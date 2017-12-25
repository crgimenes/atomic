package config

import (
	"github.com/crgimenes/goconfig"
	log "github.com/nuveo/log"
)

type Config struct {
	Debug bool   `json:"debug"`
	Host  string `json:"host"`
	Port  int    `json:"port" cfgDefault:"8888"`
}

var Get = Config{}

func Load() (err error) {
	err = goconfig.Parse(&Get)
	if err != nil {
		return
	}

	if Get.Debug {
		log.DebugMode = Get.Debug
		log.Warningln("debug mode on")
	}
	return
}
