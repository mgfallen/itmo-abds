package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort            string
	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseHttpPort string
	ClickHouseUser     string
	ClickHousePassword string
	ClickHouseDatabase string
}

func Load() *Config {
	return &Config{
		AppPort:            getEnv("APP_PORT", "8080"),
		ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "localhost"),
		ClickHousePort:     getEnv("CLICKHOUSE_PORT", "9000"),
		ClickHouseHttpPort: getEnv("CLICKHOUSE_HTTP_PORT", "8123"),
		ClickHouseUser:     getEnv("CLICKHOUSE_USER", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		ClickHouseDatabase: getEnv("CLICKHOUSE_DATABASE", "energosbyt"),
	}
}

func (c *Config) DSN() string {
	if c.ClickHouseUser != "" {
		return fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s",
			c.ClickHouseUser, c.ClickHousePassword, c.ClickHouseHost, c.ClickHousePort, c.ClickHouseDatabase)
	}
	return fmt.Sprintf("clickhouse://%s:%s/%s", c.ClickHouseHost, c.ClickHousePort, c.ClickHouseDatabase)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
