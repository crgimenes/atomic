package config

import (
	"github.com/crgimenes/goconfig"
)

type Config struct {
	Debug bool   `json:"debug"`
	Host  string `json:"host"`
	Port  int    `json:"port" cfgDefault:"8888"`
}

var Get = Config{}

func Load() (err error) {
	err = goconfig.Parse(&Get)
	return
}
