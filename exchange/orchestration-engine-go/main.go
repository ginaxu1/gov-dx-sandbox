package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ginaxu1/gov-dx-sandbox/federator"
	"github.com/ginaxu1/gov-dx-sandbox/server"
)

func main() {

	file, err := os.ReadFile("./config.json")

	if err != nil {
		log.Fatalln(fmt.Errorf("error reading config.json: %s", err.Error()))
	}

	var cfg federator.FederatorOptions

	err = json.Unmarshal(file, &cfg)

	if err != nil {
		log.Fatalln(fmt.Errorf("error unmarshalling config.json: %s", err.Error()))
	}

	federatorInstance := federator.Initialize(&cfg)

	server.RunServer(federatorInstance)
}
