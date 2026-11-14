package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

func main() {
	if err := os.MkdirAll("data", 0755); err != nil {
		panic(err)
	}

	// 1. dim_realty (1,000 rows)
	generateRealtyCSV("data/dim_realty.csv")

	// 2. dim_gauges (1,050 rows)
	generateGaugesCSV("data/dim_gauges.csv")

	// 3. dim_tariffs (2+ rows)
	generateTariffsCSV("data/dim_tariffs.csv")

	// 4. dim_users (1,000 rows)
	generateUsersCSV("data/dim_users.csv")

	// 5. fact_gauge_bindings (1,000+ rows)
	generateBindingsCSV("data/fact_gauge_bindings.csv")

	// 6. fact_meter_readings (~1,100,000 rows)
	generateReadingsCSV("data/fact_meter_readings.csv")

	// 7. fact_invoice (36,000 rows)
	generateInvoicesCSV("data/fact_invoice.csv")

	println("Datasets generated successfully in data/")
}

func generateRealtyCSV(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i := 1; i <= 1000; i++ {
		realtyType := "flat"
		number := fmt.Sprintf("%d", i)
		if i > 900 {
			realtyType = "nrcp"
			number = ""
		}
		writer.Write([]string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("г. Москва, ул. Пушкина, д. %d", i),
			"50",
			realtyType,
			number,
			"2",
			"5",
			"1",
			"2023-10-01 00:00:00",
		})
	}
}

func generateGaugesCSV(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i := 1; i <= 1050; i++ {
		writer.Write([]string{
			fmt.Sprintf("%d", i),
			"Merkur-230",
			"Russia",
			"2023-10-01 00:00:00",
		})
	}
}

func generateTariffsCSV(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Tariff 1: version 1 (July - Oct 2026)
	writer.Write([]string{"1", "День-Ночь", "500", "250", "2023-01-01 00:00:00", "2026-11-01 00:00:00", "2023-01-01 00:00:00", "2023-01-01 00:00:00"})
	// Tariff 1: version 2 (from Nov 1, 2026 onwards)
	writer.Write([]string{"1", "День-Ночь", "620", "310", "2026-11-01 00:00:00", "2099-01-01 00:00:00", "2026-10-27 00:00:00", "2026-10-27 00:00:00"})
}

func generateUsersCSV(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i := 1; i <= 1000; i++ {
		writer.Write([]string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("Иванов Иван %d-й", i),
			fmt.Sprintf("+7900000%04d", i),
			"",
			"2023-10-01 00:00:00",
			"2023-10-01 00:00:00",
		})
	}
}

func generateBindingsCSV(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i := 1; i <= 1000; i++ {
		writer.Write([]string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("%d", i),
			"2023-10-01 00:00:00",
			"1",
			"1",
		})
	}
	// Add some replacements
	for i := 1; i <= 50; i++ {
		writer.Write([]string{fmt.Sprintf("%d", i), fmt.Sprintf("%d", i), "2024-01-01 00:00:00", "-1", "1"})
		writer.Write([]string{fmt.Sprintf("%d", 1000+i), fmt.Sprintf("%d", i), "2024-01-01 00:00:00", "1", "2"})
	}
}

func generateReadingsCSV(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	start := time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

	for realty := 1; realty <= 1000; realty++ {
		cur := start
		dayVal := uint32(100)
		nightVal := uint32(50)
		for !cur.After(end) {
			dayVal += 5
			nightVal += 2
			dateStr := cur.Format("2006-01-02")
			writer.Write([]string{
				fmt.Sprintf("%d", realty),
				fmt.Sprintf("%d", realty),
				"1",
				fmt.Sprintf("%d", realty),
				dateStr,
				fmt.Sprintf("%d", dayVal),
				fmt.Sprintf("%d", nightVal),
				dateStr + " 10:00:00",
			})
			cur = cur.AddDate(0, 0, 1)
		}
	}
}

func generateInvoicesCSV(filename string) {
	file, _ := os.Create(filename)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	startYear := 2023
	startMonth := 11
	for realty := 1; realty <= 1000; realty++ {
		y := startYear
		m := startMonth
		for i := 0; i < 36; i++ {
			nextMonth := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
			deadline := fmt.Sprintf("%04d-%02d-10", nextMonth.Year(), nextMonth.Month())
			createdAt := fmt.Sprintf("%04d-%02d-25 10:00:00", y, m)

			writer.Write([]string{
				fmt.Sprintf("%d", realty),
				fmt.Sprintf("%d", realty),
				"1",
				fmt.Sprintf("%d", realty),
				"150",
				"60",
				"93000",
				deadline,
				createdAt,
			})

			m++
			if m > 12 {
				m = 1
				y++
			}
		}
	}
}
