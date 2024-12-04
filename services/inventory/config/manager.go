package config

import (
	"os"
)

type ConfigurationManager struct {
	PostgreSqlConfig Config
}

func NewConfigurationManager() *ConfigurationManager {
	return &ConfigurationManager{
		PostgreSqlConfig: getPostgreSqlConfig(),
	}
}

func getPostgreSqlConfig() Config {
	return Config{
		Host:                  os.Getenv("DB_HOST"),
		Port:                  os.Getenv("DB_PORT"),
		User:                  os.Getenv("DB_USER"),
		Password:              os.Getenv("DB_PASSWORD"),
		DbName:                os.Getenv("DB_NAME"),
		MaxConnections:        "10",
		MaxConnectionIdleTime: "30s",
	}
}
