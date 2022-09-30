package config

import (
	"encoding/json"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type config struct {
	LogLevel string `split_words:"true" default:"info"`
	Env      string `split_words:"true" default:"production"`
	API      api
	Kafka    kafka
}

type kafka struct {
	BootstrapServers  []string `split_words:"true" default:"127.0.0.1:9092"`
	ConsumerGroupName string   `split_words:"true" default:"datamin_kafka_trigger"`

	// example: DTMN_KT_KAFKA_TOPIC_MAPPING="topic_name:workflow_uuid_1,topic_name:workflow_uuid_2"
	TopicMapping string `split_words:"true"`
	SASL         sasl
}

type api struct {
	BasicAuthUsername string `split_words:"true"`
	BasicAuthPassword string `split_words:"true"`
	ClientId          string `split_words:"true"`
	ClientSecret      string `split_words:"true"`
}

type sasl struct {
	// Whether or not to use SASL authentication when connecting to the broker
	// (defaults to false).
	Enable bool `default:"false"`
	// SASLMechanism is the name of the enabled SASL mechanism.
	// Possible values: OAUTHBEARER, PLAIN (defaults to PLAIN).
	Mechanism string `default:"PLAIN"`
	// Version is the SASL Protocol Version to use
	// Kafka > 1.x should use V1, except on Azure EventHub which use V0
	Version int16 `default:"1"`
	// Whether or not to send the Kafka SASL handshake first if enabled
	// (defaults to true). You should only set this to false if you're using
	// a non-Kafka SASL proxy.
	Handshake bool `default:"true"`
	// AuthIdentity is an (optional) authorization identity (authzid) to
	// use for SASL/PLAIN authentication (if different from User) when
	// an authenticated user is permitted to act as the presented
	// alternative user. See RFC4616 for details.
	AuthIdentity string `split_words:"true"`
	// User is the authentication identity (authcid) to present for
	// SASL/PLAIN or SASL/SCRAM authentication
	User string
	// Password for SASL/PLAIN authentication
	Password string
	// authz id used for SASL/SCRAM authentication
	SCRAMAuthzID string

	GSSAPI gSSAPIConfig
}

type gSSAPIConfig struct {
	AuthType           int    `split_words:"true"`
	KeyTabPath         string `split_words:"true"`
	KerberosConfigPath string `split_words:"true"`
	ServiceName        string `split_words:"true"`
	Username           string
	Password           string
	Realm              string
	DisablePAFXFAST    bool `split_words:"true"`
}

func Cfg() config {
	return c
}

var c config

func init() {
	c = new()

	lvl, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(lvl)

	formattedConfig, _ := json.MarshalIndent(c, "", "    ")
	log.Debug("Configuration: ", string(formattedConfig))
}

func new() config {
	godotenv.Load(".env.local")
	godotenv.Load()
	var c config
	err := envconfig.Process("DTMN_KT", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	return c
}
