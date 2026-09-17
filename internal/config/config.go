package config

import "flag"

const (
	defAddress  = ":8080"
	dafDatabase = "postgres://postgres:postgres@localhost:5432/booking"
)

type Config struct {
	ServAddress string
	Database    string
}

func ParseConfig() Config {
	var cfg Config
	flag.StringVar(&cfg.ServAddress, "addr", defAddress, "Server address")
	flag.StringVar(&cfg.Database, "database", dafDatabase, "Databes address")
	flag.Parse()

	return cfg
}
