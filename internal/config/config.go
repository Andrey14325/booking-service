package config

import "flag"

const (
	defAddress = ":8080"
)

type Config struct {
	ServAddress string
}

func ParseConfig() Config {
	var cfg Config
	flag.StringVar(&cfg.ServAddress, "addr", defAddress, "Server address")
	flag.Parse()

	return cfg
}
