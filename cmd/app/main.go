package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type App struct {
	DB *sql.DB
}

func main() {
	host := getEnv("CLICKHOUSE_HOST", "localhost")
	port := getEnv("CLICKHOUSE_PORT", "9000")
	dbName := getEnv("CLICKHOUSE_DATABASE", "energosbyt")
	appPort := getEnv("APP_PORT", "8080")

	dsn := fmt.Sprintf("clickhouse://%s:%s/%s", host, port, dbName)
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer db.Close()

	app := &App{DB: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", app.handlePing)
	mux.HandleFunc("/users/", app.handleUsers)
	mux.HandleFunc("/realty/", app.handleRealty)
	mux.HandleFunc("/tariffs", app.handleTariffs)

	log.Printf("Starting server on port %s...", appPort)
	if err := http.ListenAndServe(":"+appPort, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func (app *App) handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (app *App) handleUsers(w http.ResponseWriter, r *http.Request) {
	// Parse /users/{id}
	var id int64
	_, err := fmt.Sscanf(r.URL.Path, "/users/%d", &id)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}

	if r.Method == http.MethodGet {
		var fio, phone, phonePrev string
		err := app.DB.QueryRow("SELECT fio, phone, phone_prev FROM energosbyt.dim_users FINAL WHERE id = ?", id).Scan(&fio, &phone, &phonePrev)
		if err == sql.ErrNoRows {
			httpError(w, "Not found", 404)
			return
		} else if err != nil {
			httpError(w, err.Error(), 500)
			return
		}
		writeJSON(w, map[string]interface{}{
			"id":         id,
			"fio":        fio,
			"phone":      phone,
			"phone_prev": phonePrev,
		}, 200)
	} else if r.Method == http.MethodPatch {
		var req struct {
			Fio   *string `json:"fio"`
			Phone *string `json:"phone"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Fio == nil && req.Phone == nil) {
			httpError(w, "invalid request body", 400)
			return
		}

		// Check user exists and get current values
		var curFio, curPhone, curPhonePrev string
		err := app.DB.QueryRow("SELECT fio, phone, phone_prev FROM energosbyt.dim_users FINAL WHERE id = ?", id).Scan(&curFio, &curPhone, &curPhonePrev)
		if err == sql.ErrNoRows {
			httpError(w, "Not found", 404)
			return
		}

		newFio := curFio
		if req.Fio != nil {
			if *req.Fio == "" {
				httpError(w, "invalid \"fio\" field", 400)
				return
			}
			newFio = *req.Fio
		}

		newPhone := curPhone
		newPhonePrev := curPhonePrev
		if req.Phone != nil {
			if *req.Phone == "" {
				httpError(w, "invalid \"phone\" field", 400)
				return
			}
			if *req.Phone != curPhone {
				newPhonePrev = curPhone
				newPhone = *req.Phone
			}
		}

		now := time.Now().Format("2006-01-02 15:04:05")
		_, err = app.DB.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			id, newFio, newPhone, newPhonePrev, now, now)
		if err != nil {
			httpError(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (app *App) handleRealty(w http.ResponseWriter, r *http.Request) {
	// Routing logic for /realty/{id} and subpaths
	var id int64
	var sub string
	n, _ := fmt.Sscanf(r.URL.Path, "/realty/%d/%s", &id, &sub)

	if n == 1 {
		if r.Method == http.MethodGet {
			var address, realtyType, number string
			var square uint32
			var rooms, floor, porch uint8
			err := app.DB.QueryRow("SELECT address, square, type, number, rooms, floor, porch FROM energosbyt.dim_realty WHERE id = ?", id).Scan(
				&address, &square, &realtyType, &number, &rooms, &floor, &porch,
			)
			if err == sql.ErrNoRows {
				httpError(w, "Not found", 404)
				return
			}
			writeJSON(w, map[string]interface{}{
				"id":      id,
				"address": address,
				"square":  square,
				"type":    realtyType,
				"number":  number,
				"rooms":   rooms,
				"floor":   floor,
				"porch":   porch,
			}, 200)
		}
		return
	}

	if sub == "readings" {
		if r.Method == http.MethodGet {
			dateFrom := r.URL.Query().Get("date_from")
			dateTo := r.URL.Query().Get("date_to")
			if dateFrom == "" || dateTo == "" {
				httpError(w, "invalid \"date_from\" parameter", 400)
				return
			}
			rows, err := app.DB.Query(`
				SELECT reading_date, user_id, tariff_id, gauge_id, day_kw, night_kw 
				FROM energosbyt.fact_meter_readings FINAL 
				WHERE realty_id = ? AND reading_date >= ? AND reading_date <= ? 
				ORDER BY reading_date`, id, dateFrom, dateTo)
			if err != nil {
				httpError(w, err.Error(), 500)
				return
			}
			defer rows.Close()

			type Reading struct {
				ReadingDate string `json:"reading_date"`
				UserId      int64  `json:"user_id"`
				TariffId    int64  `json:"tariff_id"`
				GaugeId     int64  `json:"gauge_id"`
				DayKw       uint32 `json:"day_kw"`
				NightKw     uint32 `json:"night_kw"`
			}
			readings := []Reading{}
			for rows.Next() {
				var r Reading
				var dt time.Time
				rows.Scan(&dt, &r.UserId, &r.TariffId, &r.GaugeId, &r.DayKw, &r.NightKw)
				r.ReadingDate = dt.Format("2006-01-02")
				readings = append(readings, r)
			}
			writeJSON(w, map[string]interface{}{
				"readings": readings,
				"count":    len(readings),
			}, 200)
		} else if r.Method == http.MethodPost {
			var req struct {
				UserId      int64  `json:"user_id"`
				GaugeId     int64  `json:"gauge_id"`
				TariffId    int64  `json:"tariff_id"`
				ReadingDate string `json:"reading_date"`
				DayKw       uint32 `json:"day_kw"`
				NightKw     uint32 `json:"night_kw"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httpError(w, "invalid request", 400)
				return
			}
			createdAt := time.Now().Format("2006-01-02 15:04:05")
			_, err := app.DB.Exec(`
				INSERT INTO energosbyt.fact_meter_readings 
				(realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				id, req.UserId, req.TariffId, req.GaugeId, req.ReadingDate, req.DayKw, req.NightKw, createdAt,
			)
			if err != nil {
				httpError(w, err.Error(), 500)
				return
			}
			w.WriteHeader(http.StatusCreated)
		}
	} else if sub == "gauges" && r.Method == http.MethodPost {
		var req struct {
			GaugeId int64 `json:"gauge_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GaugeId == 0 {
			httpError(w, "invalid \"gauge_id\" field", 400)
			return
		}

		// Find active gauge and version
		var activeGaugeId uint64
		var maxVer uint64
		err := app.DB.QueryRow(`
			SELECT gauge_id, version FROM energosbyt.fact_gauge_bindings FINAL 
			WHERE realty_id = ? GROUP BY gauge_id, version HAVING sum(sign) > 0`, id).Scan(&activeGaugeId, &maxVer)

		now := time.Now().Format("2006-01-02 15:04:05")
		if err == nil && activeGaugeId != 0 {
			// Cancel previous
			_, _ = app.DB.Exec(`
				INSERT INTO energosbyt.fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) 
				VALUES (?, ?, ?, -1, ?)`, activeGaugeId, id, now, maxVer)
			maxVer++
		} else {
			maxVer = 1
		}

		// Insert new
		_, err = app.DB.Exec(`
			INSERT INTO energosbyt.fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) 
			VALUES (?, ?, ?, 1, ?)`, req.GaugeId, id, now, maxVer)
		if err != nil {
			httpError(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func (app *App) handleTariffs(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		httpError(w, "invalid \"date\" parameter", 400)
		return
	}

	var id uint64
	var name string
	var priceDay, priceNight uint32
	var validFrom, validTo time.Time

	err := app.DB.QueryRow(`
		SELECT id, name, price_day_kwph, price_night_kwph, valid_from, valid_to 
		FROM energosbyt.dim_tariffs FINAL 
		WHERE valid_from <= ? AND valid_to > ? LIMIT 1`, dateStr+" 00:00:00", dateStr+" 00:00:00").Scan(
		&id, &name, &priceDay, &priceNight, &validFrom, &validTo,
	)
	if err == sql.ErrNoRows {
		httpError(w, "Not found", 404)
		return
	} else if err != nil {
		httpError(w, err.Error(), 500)
		return
	}

	writeJSON(w, map[string]interface{}{
		"id":               id,
		"name":             name,
		"price_day_kwph":   priceDay,
		"price_night_kwph": priceNight,
		"valid_from":       validFrom.Format("2006-01-02 15:04:05"),
		"valid_to":         validTo.Format("2006-01-02 15:04:05"),
	}, 200)
}

func httpError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func writeJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
