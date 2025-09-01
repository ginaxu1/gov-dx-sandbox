package main

import (
	"github.com/ginaxu1/gov-dx-sandbox/logger"
	"github.com/ginaxu1/gov-dx-sandbox/server"
)

func main() {
	logger.Init()
	server.RunServer()
}
