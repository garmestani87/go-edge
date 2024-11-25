package configs

import (
	"errors"
	"log"
	"os"

	"github.com/spf13/viper"
)

const (
	DEV  = "dev"
	TEST = "test"
	PROD = "prod"
)

type Config struct {
	Application
	Server
	Logging
	Otel
	Banner
	Kafka
	PublicKeys  map[string]string `mapstructure:"publicKeys"`
	ValidScopes map[string]string `mapstructure:"validScopes"`
}

type Application struct {
	Name string
}

type Server struct {
	Port    int
	RunMode string
}

type Logging struct {
	FilePath string
	FileName string
	Encoding string
	Level    string
	Logger   string
	Console  bool
}

type Otel struct {
	ServiceName           string
	ServiceVersion        string
	Insecure              string
	BearerToken           string
	DeploymentEnvironment string
	Language              string
}

type Kafka struct {
	BootstrapServers      string
	SchemaRegistry        string
	MessageMaxBytes       int
	AllowAutoCreateTopics bool
	SecurityProtocol      string
	Consumer
	Producer
}

type Consumer struct {
	GroupID           string
	AutoOffsetReset   string
	MaxPollIntervalMs int
	EnableAutoCommit  bool
}

type Producer struct {
	EnableIdempotence bool
	Acks              string
	Retries           int
}

type Banner struct {
	FilePath string
}

func Get() *Config {
	path := getPath(os.Getenv("APP_ENV"))
	v, err := load(path, "yml")
	if err != nil {
		log.Fatalf("error in loading config %v \n", err)
	}
	cfg, err := parse(v)
	if err != nil {
		log.Fatalf("error in parsing config %v \n", err)
	}
	return cfg
}

func getPath(env string) (path string) {
	switch env {
	case DEV:
		path = "/configs/application-dev"
	case TEST:
		path = "/app/configs/application-test"
	case PROD:
		path = "/app/configs/application-prod"
	}
	return
}

func load(filename string, fileType string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName(filename)
	v.SetConfigType(fileType)
	v.AddConfigPath(".")
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		log.Printf("unable to read config : %v \n", err)
		configFileNotFoundError := viper.ConfigFileNotFoundError{}
		if errors.As(err, &configFileNotFoundError) {
			return nil, errors.New("config file not found")
		}
	}
	return v, nil
}

func parse(v *viper.Viper) (cfg *Config, err error) {
	err = v.Unmarshal(&cfg)
	if err != nil {
		log.Printf("unable to parse config : %v \n", err)
		return nil, err
	}
	return cfg, nil
}
