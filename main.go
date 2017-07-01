package main

import (
	"github.com/crgimenes/atomic/config"
	"github.com/crgimenes/atomic/server"
	log "github.com/nuveo/logSys"
)

func main() {

	err := config.Load()
	if err != nil {
		log.Errorln(err)
	}
	server.Run()

}
