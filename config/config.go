package config

import (
	"github.com/gosidekick/goconfig"
)

type Config struct {
	Debug        bool   `json:"debug"`
	Host         string `json:"host"`
	Port         int    `json:"port" cfgDefault:"8888"`
	DatabasePath string `json:"database_path" cfgDefault:"./db"`
}

var cfg = Config{}

func Get() Config {
	return cfg
}

func init() {
	err := goconfig.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	return
}
