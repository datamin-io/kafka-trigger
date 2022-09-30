package command

import (
	"fmt"
	"kafkatrigger/config"
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
	sarama.Logger = syslog.New(log.StandardLogger().Out, "[Sarama] ", syslog.LstdFlags)

	cfg := config.Cfg()
	apiCfg := cfg.API
	kafkaCfg := cfg.Kafka
	topicMapping, err := parseTopicMapping(kafkaCfg.TopicMapping)

	if err != nil {
		return err
	}

	apiClient := workflows.NewClient(getBaseUrl(cfg.Env), apiCfg.ClientId, apiCfg.ClientSecret, apiCfg.BasicAuthUsername, apiCfg.BasicAuthPassword)

	var wg sync.WaitGroup
	for topicName, wfUuids := range topicMapping {
		g := goka.DefineGroup(
			goka.Group(kafkaCfg.ConsumerGroupName),
			goka.Input(goka.Stream(topicName), new(codec.Bytes), callDataminApi(apiClient, wfUuids)),
		)

		p, err := goka.NewProcessor(kafkaCfg.BootstrapServers, g)
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
		log.Debugf("Message received, running the following workflows: %v", workflowUuids)
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

			log.Infof("Message key: %s, run UUID: %s", ctx.Key(), runUuid)
		}

		log.Debugf("Workflows run successfully. Message key: %s", ctx.Key())
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

func getBaseUrl(env string) string {
	switch env {
	case "production":
		return "https://api.datamin.io"
	case "test":
		return "https://api-test.datamin.io"
	}

	panic("unknown environment specified")
}

var RunCommand = &cli.Command{
	Name:        "run",
	Description: "Run Datamin Kafka trigger",
	Usage:       "Run Datamin Kafka trigger",
	Action:      runHandler,
}
