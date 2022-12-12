package config

import (
	goconfig "crg.eti.br/go/config"
	_ "crg.eti.br/go/config/ini"
)

type Config struct {
	Debug       bool   `json:"debug" ini:"debug" cfg:"debug" cfgDefault:"false"`
	Listen      string `json:"listen" ini:"listen" cfg:"listen" cfgDefault:"0.0.0.0:2200"`
	InitBBSFile string `json:"init_bbs_file" ini:"init_bbs_file" cfg:"init_bbs_file" cfgDefault:"init.lua"`
	PrivateKey  string `json:"private_key" ini:"private_key" cfg:"private_key" cfgDefault:"id_rsa"`
}

func Load() (Config, error) {
	var cfg = Config{}
	goconfig.PrefixEnv = "ATOMIC"
	goconfig.File = "config.ini"
	err := goconfig.Parse(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
