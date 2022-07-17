package config

type Config struct {
	Debug bool   `json:"debug"`
	Host  string `json:"host"`
	Port  int    `json:"port" cfgDefault:"8888"`
}
