package repository

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"energosbyt/internal/model"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIntegrationDB(t *testing.T) (*sql.DB, Repository) {
	t.Helper()
	host := os.Getenv("CLICKHOUSE_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("CLICKHOUSE_PORT")
	if port == "" {
		port = "9000"
	}
	dbName := os.Getenv("CLICKHOUSE_DATABASE")
	if dbName == "" {
		dbName = "energosbyt"
	}

	dsn := "clickhouse://" + host + ":" + port + "/" + dbName
	db, err := sql.Open("clickhouse", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	return db, NewClickHouseRepo(db)
}

func TestClickHouseRepo_GetUser(t *testing.T) {
	t.Parallel()
	db, repo := setupIntegrationDB(t)
	// Убран defer db.Close(), чтобы параллельные тесты не закрывали соединение друг другу

	baseID := int64(100000)

	tests := []struct {
		name     string
		setupFn  func(id int64)
		wantUser *models.User
		wantErr  bool
	}{
		{
			name: "1. Existing User Initial",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, 'Иванов Иван', '+79001111111', '', ?, ?)", id, now, now)
			},
			wantUser: &models.User{ID: baseID + 1, Fio: "Иванов Иван", Phone: "+79001111111", PhonePrev: ""},
			wantErr:  false,
		},
		{
			name:     "2. Non-existent User",
			setupFn:  func(id int64) {},
			wantUser: nil,
			wantErr:  true,
		},
		{
			name: "3. User after SCD1 FIO update",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, 'Старое ФИО', '+79002222222', '', ?, ?)", id, now, now)
				newFio := "Новое ФИО"
				_ = repo.UpdateUser(id, &newFio, nil)
			},
			wantUser: &models.User{ID: baseID + 3, Fio: "Новое ФИО", Phone: "+79002222222", PhonePrev: ""},
			wantErr:  false,
		},
		{
			name: "4. User after SCD3 Phone update",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, 'Петров Петр', '+79003333333', '', ?, ?)", id, now, now)
				newPhone := "+79004444444"
				_ = repo.UpdateUser(id, nil, &newPhone)
			},
			wantUser: &models.User{ID: baseID + 4, Fio: "Петров Петр", Phone: "+79004444444", PhonePrev: "+79003333333"},
			wantErr:  false,
		},
		{
			name: "5. User with special characters in FIO",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, 'О`Коннор-Смит Анна-Мария', '+79005555555', '', ?, ?)", id, now, now)
			},
			wantUser: &models.User{ID: baseID + 5, Fio: "О`Коннор-Смит Анна-Мария", Phone: "+79005555555", PhonePrev: ""},
			wantErr:  false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			currentID := baseID + int64(i+1)
			tt.setupFn(currentID)

			user, err := repo.GetUser(currentID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.wantUser.Fio, user.Fio)
				assert.Equal(t, tt.wantUser.Phone, user.Phone)
				assert.Equal(t, tt.wantUser.PhonePrev, user.PhonePrev)
			}
		})
	}
}

func TestClickHouseRepo_UpdateUser(t *testing.T) {
	t.Parallel()
	db, repo := setupIntegrationDB(t)

	baseID := int64(200000)

	tests := []struct {
		name    string
		setupFn func(id int64)
		fio     *string
		phone   *string
		wantErr bool
		checkFn func(t *testing.T, id int64)
	}{
		{
			name: "1. Update only FIO (SCD1)",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, 'Иван', '+79111111111', '', ?, ?)", id, now, now)
			},
			fio:     lo.ToPtr("Степан"),
			phone:   nil,
			wantErr: false,
			checkFn: func(t *testing.T, id int64) {
				u, _ := repo.GetUser(id)
				assert.Equal(t, "Степан", u.Fio)
				assert.Equal(t, "+79111111111", u.Phone)
				assert.Equal(t, "", u.PhonePrev)
			},
		},
		{
			name: "2. Update only Phone (SCD3)",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, 'Алексей', '+79222222222', '', ?, ?)", id, now, now)
			},
			fio:     nil,
			phone:   lo.ToPtr("+79333333333"),
			wantErr: false,
			checkFn: func(t *testing.T, id int64) {
				u, _ := repo.GetUser(id)
				assert.Equal(t, "Алексей", u.Fio)
				assert.Equal(t, "+79333333333", u.Phone)
				assert.Equal(t, "+79222222222", u.PhonePrev)
			},
		},
		{
			name: "3. Update both FIO and Phone simultaneously",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, 'Дмитрий', '+79444444444', '', ?, ?)", id, now, now)
			},
			fio:     lo.ToPtr("Дмитрий Новый"),
			phone:   lo.ToPtr("+79555555555"),
			wantErr: false,
			checkFn: func(t *testing.T, id int64) {
				u, _ := repo.GetUser(id)
				assert.Equal(t, "Дмитрий Новый", u.Fio)
				assert.Equal(t, "+79555555555", u.Phone)
				assert.Equal(t, "+79444444444", u.PhonePrev)
			},
		},
		{
			name:    "4. Update non-existent user",
			setupFn: func(id int64) {},
			fio:     lo.ToPtr("Призрак"),
			phone:   nil,
			wantErr: true,
			checkFn: func(t *testing.T, id int64) {},
		},
		{
			name: "5. Update with same phone (should not overwrite phone_prev)",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, 'Ольга', '+79666666666', ' старый ', ?, ?)", id, now, now)
			},
			fio:     nil,
			phone:   lo.ToPtr("+79666666666"),
			wantErr: false,
			checkFn: func(t *testing.T, id int64) {
				u, _ := repo.GetUser(id)
				assert.Equal(t, "+79666666666", u.Phone)
				assert.Equal(t, " старый ", u.PhonePrev)
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			currentID := baseID + int64(i+1)
			tt.setupFn(currentID)

			err := repo.UpdateUser(currentID, tt.fio, tt.phone)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.checkFn(t, currentID)
			}
		})
	}
}

func TestClickHouseRepo_GetRealty(t *testing.T) {
	t.Parallel()
	db, repo := setupIntegrationDB(t)

	baseID := int64(300000)

	tests := []struct {
		name       string
		setupFn    func(id int64)
		wantRealty *models.Realty
		wantErr    bool
	}{
		{
			name: "1. Existing flat realty",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_realty (id, address, square, type, number, rooms, floor, porch, created_at) VALUES (?, 'г. Москва, ул. Ленина, 1', 65, 'flat', '42', 3, 9, 2, ?)", id, now)
			},
			wantRealty: &models.Realty{ID: baseID + 1, Address: "г. Москва, ул. Ленина, 1", Square: 65, Type: "flat", Number: "42", Rooms: 3, Floor: 9, Porch: 2},
			wantErr:    false,
		},
		{
			name: "2. Existing nrcp (non-residential) realty",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_realty (id, address, square, type, number, rooms, floor, porch, created_at) VALUES (?, 'г. Москва, ул. Ленина, 1 стр 2', 120, 'nrcp', '', 0, 1, 1, ?)", id, now)
			},
			wantRealty: &models.Realty{ID: baseID + 2, Address: "г. Москва, ул. Ленина, 1 стр 2", Square: 120, Type: "nrcp", Number: "", Rooms: 0, Floor: 1, Porch: 1},
			wantErr:    false,
		},
		{
			name:       "3. Non-existent realty ID",
			setupFn:    func(id int64) {},
			wantRealty: nil,
			wantErr:    true,
		},
		{
			name: "4. Realty with small square and 1 room",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_realty (id, address, square, type, number, rooms, floor, porch, created_at) VALUES (?, 'г. Санкт-Петербург, ул. Мира, 5', 25, 'flat', '12', 1, 2, 1, ?)", id, now)
			},
			wantRealty: &models.Realty{ID: baseID + 4, Address: "г. Санкт-Петербург, ул. Мира, 5", Square: 25, Type: "flat", Number: "12", Rooms: 1, Floor: 2, Porch: 1},
			wantErr:    false,
		},
		{
			name: "5. Realty with high floor and porch numbers",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.dim_realty (id, address, square, type, number, rooms, floor, porch, created_at) VALUES (?, 'г. Казань, пр. Победы, 100', 90, 'flat', '555', 4, 25, 6, ?)", id, now)
			},
			wantRealty: &models.Realty{ID: baseID + 5, Address: "г. Казань, пр. Победы, 100", Square: 90, Type: "flat", Number: "555", Rooms: 4, Floor: 25, Porch: 6},
			wantErr:    false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			currentID := baseID + int64(i+1)
			tt.setupFn(currentID)

			realty, err := repo.GetRealty(currentID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, realty)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, realty)
				assert.Equal(t, tt.wantRealty.Address, realty.Address)
				assert.Equal(t, tt.wantRealty.Square, realty.Square)
				assert.Equal(t, tt.wantRealty.Type, realty.Type)
				assert.Equal(t, tt.wantRealty.Rooms, realty.Rooms)
			}
		})
	}
}

func TestClickHouseRepo_GetReadings(t *testing.T) {
	t.Parallel()
	db, repo := setupIntegrationDB(t)

	baseID := int64(400000)

	tests := []struct {
		name          string
		setupFn       func(id int64)
		dateFrom      string
		dateTo        string
		expectedCount int
	}{
		{
			name: "1. Get single day reading",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.fact_meter_readings (realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) VALUES (?, 1, 1, 1, '2026-10-01', 100, 50, ?)", id, now)
			},
			dateFrom:      "2026-10-01",
			dateTo:        "2026-10-01",
			expectedCount: 1,
		},
		{
			name: "2. Get range with multiple readings",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.fact_meter_readings (realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) VALUES (?, 1, 1, 1, '2026-10-01', 100, 50, ?)", id, now)
				_, _ = db.Exec("INSERT INTO energosbyt.fact_meter_readings (realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) VALUES (?, 1, 1, 1, '2026-10-02', 110, 55, ?)", id, now)
			},
			dateFrom:      "2026-10-01",
			dateTo:        "2026-10-05",
			expectedCount: 2,
		},
		{
			name: "3. Readings with correction (ReplacingMergeTree FINAL check)",
			setupFn: func(id int64) {
				_, _ = db.Exec("INSERT INTO energosbyt.fact_meter_readings (realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) VALUES (?, 1, 1, 1, '2026-10-10', 200, 100, '2026-10-10 10:00:00')", id)
				_, _ = db.Exec("INSERT INTO energosbyt.fact_meter_readings (realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) VALUES (?, 1, 1, 1, '2026-10-10', 250, 120, '2026-10-10 15:00:00')", id)
			},
			dateFrom:      "2026-10-10",
			dateTo:        "2026-10-10",
			expectedCount: 1,
		},
		{
			name:          "4. Get range with no data",
			setupFn:       func(id int64) {},
			dateFrom:      "2025-01-01",
			dateTo:        "2025-01-31",
			expectedCount: 0,
		},
		{
			name: "5. Get readings outside query date range filter",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.fact_meter_readings (realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) VALUES (?, 1, 1, 1, '2026-10-15', 300, 150, ?)", id, now)
			},
			dateFrom:      "2026-10-01",
			dateTo:        "2026-10-05",
			expectedCount: 0,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			currentID := baseID + int64(i+1)
			tt.setupFn(currentID)

			readings, err := repo.GetReadings(currentID, tt.dateFrom, tt.dateTo)
			assert.NoError(t, err)
			assert.Len(t, readings, tt.expectedCount)
		})
	}
}

func TestClickHouseRepo_AddReading(t *testing.T) {
	t.Parallel()
	_, repo := setupIntegrationDB(t)
	baseID := int64(500000)

	tests := []struct {
		name        string
		realtyID    int64
		userId      int64
		gaugeId     int64
		tariffId    int64
		readingDate string
		dayKw       uint32
		nightKw     uint32
		wantErr     bool
	}{
		{
			name:        "1. Add standard daily reading",
			realtyID:    baseID + 1,
			userId:      1,
			gaugeId:     1,
			tariffId:    1,
			readingDate: "2026-10-01",
			dayKw:       150,
			nightKw:     60,
			wantErr:     false,
		},
		{
			name:        "2. Add reading with zero consumption",
			realtyID:    baseID + 2,
			userId:      1,
			gaugeId:     1,
			tariffId:    1,
			readingDate: "2026-10-02",
			dayKw:       0,
			nightKw:     0,
			wantErr:     false,
		},
		{
			name:        "3. Add correction reading for same date",
			realtyID:    baseID + 3,
			userId:      1,
			gaugeId:     1,
			tariffId:    1,
			readingDate: "2026-10-03",
			dayKw:       500,
			nightKw:     200,
			wantErr:     false,
		},
		{
			name:        "4. Add reading with large kw values",
			realtyID:    baseID + 4,
			userId:      1,
			gaugeId:     1,
			tariffId:    1,
			readingDate: "2026-10-04",
			dayKw:       999999,
			nightKw:     888888,
			wantErr:     false,
		},
		{
			name:        "5. Add reading with future date",
			realtyID:    baseID + 5,
			userId:      1,
			gaugeId:     1,
			tariffId:    1,
			readingDate: "2030-12-31",
			dayKw:       100,
			nightKw:     50,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := repo.AddReading(tt.realtyID, tt.userId, tt.gaugeId, tt.tariffId, tt.readingDate, tt.dayKw, tt.nightKw)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClickHouseRepo_AddGaugeBinding(t *testing.T) {
	t.Parallel()
	db, repo := setupIntegrationDB(t)
	baseID := int64(600000)

	tests := []struct {
		name    string
		setupFn func(id int64)
		gaugeID int64
		wantErr bool
	}{
		{
			name:    "1. Initial gauge installation (version 1)",
			setupFn: func(id int64) {},
			gaugeID: 101,
			wantErr: false,
		},
		{
			name: "2. Gauge replacement (version 1 to version 2)",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) VALUES (102, ?, ?, 1, 1)", id, now)
			},
			gaugeID: 103,
			wantErr: false,
		},
		{
			name: "3. Multiple sequential replacements",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) VALUES (104, ?, ?, 1, 1)", id, now)
			},
			gaugeID: 105,
			wantErr: false,
		},
		{
			name:    "4. Install gauge with high ID",
			setupFn: func(id int64) {},
			gaugeID: 999999,
			wantErr: false,
		},
		{
			name: "5. Re-binding after prior cancellation",
			setupFn: func(id int64) {
				now := time.Now().Format("2006-01-02 15:04:05")
				_, _ = db.Exec("INSERT INTO energosbyt.fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) VALUES (106, ?, ?, 1, 1)", id, now)
				_, _ = db.Exec("INSERT INTO energosbyt.fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) VALUES (106, ?, ?, -1, 1)", id, now)
			},
			gaugeID: 107,
			wantErr: false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			currentID := baseID + int64(i+1)
			tt.setupFn(currentID)

			err := repo.AddGaugeBinding(currentID, tt.gaugeID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClickHouseRepo_Integration_GetTariffByDate(t *testing.T) {
	t.Parallel()
	db, repo := setupIntegrationDB(t)

	// Очищаем таблицу перед тестом во избежание дублирования записей с id = 777
	_, err := db.Exec("TRUNCATE TABLE energosbyt.dim_tariffs")
	require.NoError(t, err)

	now := time.Now().Format("2006-01-02 15:04:05")
	_, _ = db.Exec("INSERT INTO energosbyt.dim_tariffs (id, name, price_day_kwph, price_night_kwph, valid_from, valid_to, created_at, updated_at) VALUES (777, 'Тестовый тариф v1', 500, 250, '2026-01-01 00:00:00', '2026-07-01 00:00:00', ?, ?)", now, now)
	_, _ = db.Exec("INSERT INTO energosbyt.dim_tariffs (id, name, price_day_kwph, price_night_kwph, valid_from, valid_to, created_at, updated_at) VALUES (777, 'Тестовый тариф v2', 600, 300, '2026-07-01 00:00:00', '2099-01-01 00:00:00', ?, ?)", now, now)

	tests := []struct {
		name      string
		dateStr   string
		wantID    uint64
		wantPrice uint32
		wantErr   bool
	}{
		{
			name:      "1. Date within Tariff v1 range",
			dateStr:   "2026-03-15",
			wantID:    777,
			wantPrice: 500,
			wantErr:   false,
		},
		{
			name:      "2. Date within Tariff v2 range",
			dateStr:   "2026-10-27",
			wantID:    777,
			wantPrice: 600,
			wantErr:   false,
		},
		{
			name:      "3. Exact boundary date valid_from of v2",
			dateStr:   "2026-07-01",
			wantID:    777,
			wantPrice: 600,
			wantErr:   false,
		},
		{
			name:      "4. Date before any tariff version",
			dateStr:   "2020-01-01",
			wantID:    0,
			wantPrice: 0,
			wantErr:   true,
		},
		{
			name:      "5. Date in far future (open-ended version)",
			dateStr:   "2050-01-01",
			wantID:    777,
			wantPrice: 600,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tariff, err := repo.GetTariffByDate(tt.dateStr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tariff)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, tariff)
				assert.Equal(t, tt.wantID, tariff.ID)
				assert.Equal(t, tt.wantPrice, tariff.PriceDayKwph)
			}
		})
	}
}
