package config

import (
	"fmt"
	"log"

	"os"
	"sync"

	"github.com/gagliardetto/solana-go/rpc"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Logging Logging `yaml:"logging"`
	Solana  Solana  `yaml:"solana"`
}

type Logging struct {
	Level string `yaml:"level"`
	Stage string `yaml:"stage"`
}

type Solana struct {
	URL    string `yaml:"rpc-endpoint"`
	Client *rpc.Client
}

var (
	appConfig Config
	once      sync.Once
)

func Get() Config {
	once.Do(
		func() {
			path := os.Getenv("APP_CONFIG_PATH")
			if path == "" {
				path = "config.yaml"
			}
			if err := initWithPath(path); err != nil {
				log.Fatalf("failed to init config: %s", err)
			}
		},
	)
	return appConfig
}

func (c Config) Client() *rpc.Client {
	return c.Solana.Client
}

func initWithPath(configPath string) error {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := yaml.Unmarshal(b, &appConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	appConfig.Solana.Client = rpc.New(appConfig.Solana.URL)
	return nil
}
