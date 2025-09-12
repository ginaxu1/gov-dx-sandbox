package main

import (
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/configs"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/server"
)

func main() {
	logger.Init()
	configs.LoadConfig()

	var federationObject = federator.Initialize()

	server.RunServer(federationObject)
}
