package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logging Logging `yaml:"logging"`
}

type Logging struct {
	Level string `yaml:"level"`
	Stage string `yaml:"stage"`
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

func initWithPath(configPath string) error {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := yaml.Unmarshal(b, &appConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
