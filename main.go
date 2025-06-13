package main

import (
	"fmt"

	"example.com/m/pkg/cache"
	"example.com/m/pkg/config"
	"example.com/m/pkg/logger"
)

func InitApp() {
	config.InitConfig()
	logger.InitLogger()
	cache.InitCache()
}

func main() {
	fmt.Println("initializing app")
	InitApp()
	logger.Info("app initialized successfully")
}
