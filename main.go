package main

import (
	"context"
	"log"
	"os"
	"os/signal"

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

	int := make(chan os.Signal, 1)
	signal.Notify(int, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { <-int; cancel() }()
	err = server.Run(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
}
