package main

import (
	"fmt"

	"example.com/m/pkg/logger"
)

func main() {
	fmt.Println("initializing app")
	logger.Info(nil, "app initialized successfully")
}
