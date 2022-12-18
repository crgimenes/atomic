package config

import (
	"path/filepath"

	"crg.eti.br/go/config"
	_ "crg.eti.br/go/config/ini"
)

type Config struct {
	Debug      bool   `json:"debug" ini:"debug" cfg:"debug" cfgDefault:"false"`
	Listen     string `json:"listen" ini:"listen" cfg:"listen" cfgDefault:"0.0.0.0:2200"`
	PrivateKey string `json:"private_key" ini:"private_key" cfg:"private_key" cfgDefault:"id_rsa"`
	BaseBBSDir string `json:"base_bbs_dir" ini:"base_bbs_dir" cfg:"base_bbs_dir" cfgDefault:"./"`
}

func Load() (Config, error) {
	var cfg = Config{}
	config.PrefixEnv = "ATOMIC"
	config.File = "config.ini"
	err := config.Parse(&cfg)
	if err != nil {
		return Config{}, err
	}

	cfg.BaseBBSDir, err = filepath.Abs(cfg.BaseBBSDir)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
