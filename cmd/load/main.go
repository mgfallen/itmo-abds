package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func main() {
	host := getEnv("CLICKHOUSE_HOST", "localhost")
	httpPort := getEnv("CLICKHOUSE_HTTP_PORT", "8123")
	user := getEnv("CLICKHOUSE_USER", "default")
	password := getEnv("CLICKHOUSE_PASSWORD", "")
	dbName := getEnv("CLICKHOUSE_DATABASE", "energosbyt")

	tables := []string{
		"dim_realty",
		"dim_gauges",
		"dim_tariffs",
		"dim_users",
		"fact_gauge_bindings",
		"fact_meter_readings",
		"fact_invoice",
	}

	for _, table := range tables {
		csvPath := filepath.Join("data", table+".csv")
		file, err := os.Open(csvPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("File %s not found, skipping\n", csvPath)
				continue
			}
			panic(err)
		}

		query := fmt.Sprintf("INSERT INTO %s.%s FORMAT CSV", dbName, table)
		reqURL := fmt.Sprintf("http://%s:%s/?query=%s", host, httpPort, url.QueryEscape(query))

		req, err := http.NewRequest("POST", reqURL, file)
		if err != nil {
			file.Close()
			panic(err)
		}
		if user != "" {
			req.SetBasicAuth(user, password)
		}

		resp, err := http.DefaultClient.Do(req)
		file.Close() // Закрываем файл после завершения запроса
		if err != nil {
			panic(fmt.Sprintf("failed to insert %s: %v", table, err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			panic(fmt.Sprintf("clickhouse error for %s (status %d): %s", table, resp.StatusCode, string(bodyBytes)))
		}

		fmt.Printf("Loaded table %s\n", table)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
