package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel   string `mapstructure:"LOG_LEVEL"`
	StageLevel string `mapstructure:"STAGE_ENV"`
	// TODO: Add more
}

var appConfig Config

func GetConfig() Config {
	return appConfig
}

func InitConfig() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Try to add the project root dynamically
	wd, _ := os.Getwd()
	for i := 0; i < 5; i++ { // search up to 5 levels up
		envPath := filepath.Join(wd, ".env")
		if _, err := os.Stat(envPath); err == nil {
			viper.AddConfigPath(wd)
			break
		}
		wd = filepath.Dir(wd) // move one level up
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read .env file: %v", err)
	}

	if err := viper.Unmarshal(&appConfig); err != nil {
		log.Fatalf("Error unmarshaling env vars: %v", err)
	}

}
