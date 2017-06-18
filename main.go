package main

import "github.com/crgimenes/goConfig"
import log "github.com/nuveo/logSys"

type config struct {
	Debug bool
}

func main() {
	cfg := config{}

	err := goConfig.Parse(&cfg)
	if err != nil {
		log.Errorln(err.Error())
		return
	}

	if cfg.Debug {
		log.DebugMode = cfg.Debug
		log.Warningln("debug mode on")
	}
}
