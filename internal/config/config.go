package config

import (
	_ "github.com/joho/godotenv/autoload" // use dotenv file mostly for development purposes
	"github.com/kelseyhightower/envconfig"
)

// Config describes app configuration
type Config struct {
	Debug         bool     `envconfig:"DEBUG" default:"false"`
	TasksFilePath string   `envconfig:"TASKS_FILE_PATH"`
	DatabaseDSN   string   `envconfig:"DATABASE_DSN"`
	MigrationsDir string   `envconfig:"MIGRATIONS_DIR" default:"/migrations"`
	KafkaBrokers  []string `envconfig:"KAFKA_BROKERS"`
	CertFile      string   `envconfig:"CERT_FILE" default:""`
	KeyFile       string   `envconfig:"KEY_FILE" default:""`
	CAFile        string   `envconfig:"CA_FILE" default:""`
}

// Read reads config from env variables
func Read() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// MustRead reads config from env variables or panics in case of error
func MustRead() *Config {
	cfg, err := Read()
	if err != nil {
		panic("config read error: " + err.Error())
	}
	return cfg
}
