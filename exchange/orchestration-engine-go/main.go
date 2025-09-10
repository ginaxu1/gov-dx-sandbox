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

	query := `
		query BasicInfoQuery {
		  personInfo(nic: "199512345678") {
		    name
		    address
		    profession
		    birthInfo {
		      brNo
			}
			birthRegistrationNumber
		  }
		}
	`
	_ = query

	//federator.QueryBuilder(query)

	var federationObject = federator.Initialize()

	server.RunServer(federationObject)
}
