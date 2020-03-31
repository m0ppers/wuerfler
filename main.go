package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/m0ppers/wuerfler/config"
	"github.com/m0ppers/wuerfler/server"
)

func main() {
	var conf config.Config
	err := envconfig.Process("wuerfler", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}
	server := server.NewServer(conf)
	server.Run()
}
