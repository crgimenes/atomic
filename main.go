package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/server"
	log "github.com/nuveo/logSys"
)

func main() {

	err := config.Load()
	if err != nil {
		log.Errorln(err)
	}

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt)
		<-sc

		log.Warningln("\r\nshutting down...\r\n")

		// clear

		fmt.Printf("have a nice day!\r\n")
		os.Exit(0)
	}()

	server.Run()

}
