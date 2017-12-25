package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/server"
	log "github.com/nuveo/log"
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

		log.Warningln("shutting down...")

		// clear

		os.Exit(0)
	}()

	/*
	   0        1          2        3        4        5        6        7        8
	   123456789012234567890234567890234567890234567890234567890234567890234567890
	*/
	fmt.Println(`
 ██████╗██████╗  ██████╗    ███████╗████████╗██╗   ██████╗ ██████╗ 
██╔════╝██╔══██╗██╔════╝    ██╔════╝╚══██╔══╝██║   ██╔══██╗██╔══██╗
██║     ██████╔╝██║  ███╗   █████╗     ██║   ██║   ██████╔╝██████╔╝
██║     ██╔══██╗██║   ██║   ██╔══╝     ██║   ██║   ██╔══██╗██╔══██╗
╚██████╗██║  ██║╚██████╔╝██╗███████╗   ██║   ██║██╗██████╔╝██║  ██║
 ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝╚══════╝   ╚═╝   ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝
`)

	server.Run()

}
