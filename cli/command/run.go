package command

import (
	"crypto/tls"
	"fmt"
	"kafkatrigger/config"
	"kafkatrigger/services/kafka/sasl"
	"kafkatrigger/services/workflows"
	syslog "log"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var runHandler cli.ActionFunc = func(c *cli.Context) error {
	log.Info("Starting kafka listener...")
	sarama.Logger = syslog.New(log.StandardLogger().Out, "[Datamin Kafka trigger] ", syslog.LstdFlags)

	cfg := config.Cfg()
	apiCfg := cfg.API
	kafkaCfg := cfg.Kafka
	topicMapping, err := parseTopicMapping(kafkaCfg.TopicMapping)

	if err != nil {
		return err
	}

	replaceGokaConfig()

	baseAPIUrl := apiCfg.Url 
	if baseAPIUrl == "" {
		baseAPIUrl = getBaseUrlByEnv(cfg.Env)
	}

	apiClient := workflows.NewClient(baseAPIUrl, apiCfg.ClientId, apiCfg.ClientSecret, apiCfg.BasicAuthUsername, apiCfg.BasicAuthPassword)

	var wg sync.WaitGroup
	for topicName, wfUuids := range topicMapping {
		g := goka.DefineGroup(
			goka.Group(kafkaCfg.ConsumerGroupName),
			goka.Input(goka.Stream(topicName), new(codec.Bytes), callDataminApi(apiClient, wfUuids)),
		)

		p, err := goka.NewProcessor(
			kafkaCfg.BootstrapServers,
			g,
		)
		if err != nil {
			log.Error(err)
			continue
		} else {
			log.Info("Started.")
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err = p.Run(c.Context); err != nil {
				log.Error("Error running processor: %s", err.Error())
			} else {
				log.Info("Processor shut down cleanly")
			}
		}()
	}

	wg.Wait()

	return nil
}

func callDataminApi(apiClient workflows.Client, workflowUuids []string) goka.ProcessCallback {
	return func(ctx goka.Context, msg interface{}) {
		log.Debugf("Message received, running the following pipelines: %v", workflowUuids)
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic while transmitting a message, recovered and skipped\n", r)
			}
		}()

		for _, wfUuid := range workflowUuids {
			runUuid, err := apiClient.RunWorkflow(wfUuid, msg.([]byte))
			if err != nil {
				log.Error(err)
				return
			}

			log.Infof("Message key: %s, pipeline UUID: %s, pipeline run UUID: %s", ctx.Key(), wfUuid, runUuid)
		}

		log.Debugf("Pipelines run successfully. Message key: %s", ctx.Key())
	}
}

func parseTopicMapping(raw string) (map[string][]string, error) {
	result := make(map[string][]string, 0)
	items := strings.Split(raw, ",")
	for _, v := range items {
		mappingItem := strings.Split(v, ":")
		if len(mappingItem) != 2 {
			return nil, fmt.Errorf("invalid mapping format: %s", v)
		}

		topicName := strings.TrimSpace(mappingItem[0])
		workflowUuid := strings.TrimSpace(mappingItem[1])
		if len(topicName) == 0 || len(workflowUuid) == 0 {
			return nil, fmt.Errorf("invalid mapping format: %s", v)
		}
		if result[topicName] == nil {
			result[topicName] = make([]string, 0)
		}

		result[topicName] = append(result[topicName], workflowUuid)
	}

	return result, nil
}

func getBaseUrlByEnv(env string) string {
	switch env {
	case "production":
		return "https://api.datamin.io"
	case "test":
		return "https://api-test.datamin.io"
	}

	panic("unknown environment specified")
}

func replaceGokaConfig() {
	kc := config.Cfg().Kafka
	c := goka.DefaultConfig()
	c.Version = kafkaVersionFromString(kc.Version)

	c.Net.TLS.Enable = kc.TLS.Enable
	if c.Net.TLS.Enable {
		c.Net.TLS.Config = newTlsConfig()
	}

	c.Net.SASL.Enable = kc.SASL.Enable
	c.Net.SASL.Mechanism = sarama.SASLMechanism(kc.SASL.Mechanism)
	c.Net.SASL.Version = kc.SASL.Version
	c.Net.SASL.Handshake = kc.SASL.Handshake
	c.Net.SASL.AuthIdentity = kc.SASL.AuthIdentity
	c.Net.SASL.User = kc.SASL.User
	c.Net.SASL.Password = kc.SASL.Password
	c.Net.SASL.SCRAMAuthzID = kc.SASL.SCRAMAuthzID

	if c.Net.SASL.Mechanism == sarama.SASLTypeSCRAMSHA512 {
		c.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &sasl.XDGSCRAMClient{HashGeneratorFcn: sasl.SHA512} }
		c.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	} else if c.Net.SASL.Mechanism == sarama.SASLTypeSCRAMSHA256 {
		c.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &sasl.XDGSCRAMClient{HashGeneratorFcn: sasl.SHA256} }
		c.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	}

	c.Net.SASL.GSSAPI.AuthType = kc.SASL.GSSAPI.AuthType
	c.Net.SASL.GSSAPI.KeyTabPath = kc.SASL.GSSAPI.KeyTabPath
	c.Net.SASL.GSSAPI.KerberosConfigPath = kc.SASL.GSSAPI.KerberosConfigPath
	c.Net.SASL.GSSAPI.ServiceName = kc.SASL.GSSAPI.ServiceName
	c.Net.SASL.GSSAPI.Username = kc.SASL.GSSAPI.Username
	c.Net.SASL.GSSAPI.Password = kc.SASL.GSSAPI.Password
	c.Net.SASL.GSSAPI.Realm = kc.SASL.GSSAPI.Realm
	c.Net.SASL.GSSAPI.DisablePAFXFAST = kc.SASL.GSSAPI.DisablePAFXFAST

	goka.ReplaceGlobalConfig(c)
}

func newTlsConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}

func kafkaVersionFromString(v string) sarama.KafkaVersion {
	ver, err := sarama.ParseKafkaVersion(v)
	if err != nil {
		panic(err)
	}

	return ver
}

var RunCommand = &cli.Command{
	Name:        "run",
	Description: "Run Datamin Kafka trigger",
	Usage:       "Run Datamin Kafka trigger",
	Action:      runHandler,
}
