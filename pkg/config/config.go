package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	LogLevel   string `mapstructure:"LOG_LEVEL"`
	StageLevel string `mapstructure:"STAGE_ENV"`
	// TODO: Add more
}

var AppConfig Config

func InitConfig() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read .env file: %v", err)
	}

	// Unmarshal env vars into the struct using mapstructure tags
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Error unmarshaling env vars: %v", err)
	}

}
