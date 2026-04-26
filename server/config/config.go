package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DBPath        string
	RetentionDays int
}

func Load() Config {
	retentionDays := 30
	if v := os.Getenv("POMELO_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			retentionDays = n
		}
	}
	dbPath := os.Getenv("POMELO_DB_PATH")
	if dbPath == "" {
		dbPath = "pomelodata.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port, DBPath: dbPath, RetentionDays: retentionDays}
}
