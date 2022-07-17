package config

import (
	"github.com/crgimenes/goconfig"
)

type Config struct {
	Debug bool   `json:"debug"`
	Host  string `json:"host"`
	Port  int    `json:"port" cfgDefault:"8888"`
}

func Load() (Config, error) {
	var cfg = Config{}
	err := goconfig.Parse(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
