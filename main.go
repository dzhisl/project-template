package main

import (
	"fmt"

	// "example.com/m/internal/sol/utils"
	"example.com/m/pkg/cache"
	"example.com/m/pkg/config"
	"example.com/m/pkg/logger"
	// "github.com/gagliardetto/solana-go/rpc"
	// "github.com/spf13/viper"
)

func InitApp() {
	config.InitConfig()
	logger.InitLogger()
	cache.InitCache()

	// utils.InitRpcUtils(rpc.New(viper.GetString("RPC_URL")))
}

func main() {
	fmt.Println("initializing app")
	InitApp()
	logger.Info("app initialized successfully")
}
