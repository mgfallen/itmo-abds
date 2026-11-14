package repository

import (
	"database/sql"
	models "energosbyt/internal/model"
	"time"
)

//go:generate mockgen -source=repo.go -destination=mocks/mock_repository.go -package=mocks

type Repository interface {
	GetUser(id int64) (*models.User, error)
	UpdateUser(id int64, fio, phone *string) error
	GetRealty(id int64) (*models.Realty, error)
	GetReadings(realtyID int64, dateFrom, dateTo string) ([]models.Reading, error)
	AddReading(realtyID int64, userId, gaugeId, tariffId int64, readingDate string, dayKw, nightKw uint32) error
	AddGaugeBinding(realtyID, gaugeId int64) error
	GetTariffByDate(dateStr string) (*models.Tariff, error)
	GetInvoices(realtyID *int64, limit, offset int) ([]models.Invoice, error)
	GetInvoicesCount(realtyID *int64) (int, error)
}

type ClickHouseRepo struct {
	db *sql.DB
}

func NewClickHouseRepo(db *sql.DB) Repository {
	return &ClickHouseRepo{db: db}
}

func (r *ClickHouseRepo) GetUser(id int64) (*models.User, error) {
	var u models.User
	u.ID = id
	err := r.db.QueryRow("SELECT fio, phone, phone_prev FROM energosbyt.dim_users FINAL WHERE id = ?", id).
		Scan(&u.Fio, &u.Phone, &u.PhonePrev)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *ClickHouseRepo) UpdateUser(id int64, fio, phone *string) error {
	var curFio, curPhone, curPhonePrev string
	err := r.db.QueryRow("SELECT fio, phone, phone_prev FROM energosbyt.dim_users FINAL WHERE id = ?", id).
		Scan(&curFio, &curPhone, &curPhonePrev)
	if err != nil {
		return err
	}

	newFio := curFio
	if fio != nil {
		newFio = *fio
	}

	newPhone := curPhone
	newPhonePrev := curPhonePrev
	if phone != nil && *phone != curPhone {
		newPhonePrev = curPhone
		newPhone = *phone
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = r.db.Exec("INSERT INTO energosbyt.dim_users (id, fio, phone, phone_prev, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, newFio, newPhone, newPhonePrev, now, now)
	return err
}

func (r *ClickHouseRepo) GetRealty(id int64) (*models.Realty, error) {
	var realty models.Realty
	realty.ID = id
	err := r.db.QueryRow("SELECT address, square, type, number, rooms, floor, porch FROM energosbyt.dim_realty WHERE id = ?", id).
		Scan(&realty.Address, &realty.Square, &realty.Type, &realty.Number, &realty.Rooms, &realty.Floor, &realty.Porch)
	if err != nil {
		return nil, err
	}
	return &realty, nil
}

func (r *ClickHouseRepo) GetReadings(realtyID int64, dateFrom, dateTo string) ([]models.Reading, error) {
	rows, err := r.db.Query(`
		SELECT reading_date, user_id, tariff_id, gauge_id, day_kw, night_kw 
		FROM energosbyt.fact_meter_readings FINAL 
		WHERE realty_id = ? AND reading_date >= ? AND reading_date <= ? 
		ORDER BY reading_date`, realtyID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []models.Reading
	for rows.Next() {
		var rd models.Reading
		var dt time.Time
		if err := rows.Scan(&dt, &rd.UserId, &rd.TariffId, &rd.GaugeId, &rd.DayKw, &rd.NightKw); err != nil {
			return nil, err
		}
		rd.ReadingDate = dt.Format("2006-01-02")
		readings = append(readings, rd)
	}
	if readings == nil {
		readings = []models.Reading{}
	}
	return readings, nil
}

func (r *ClickHouseRepo) AddReading(realtyID int64, userId, gaugeId, tariffId int64, readingDate string, dayKw, nightKw uint32) error {
	createdAt := time.Now().Format("2006-01-02 15:04:05")
	_, err := r.db.Exec(`
		INSERT INTO energosbyt.fact_meter_readings 
		(realty_id, user_id, tariff_id, gauge_id, reading_date, day_kw, night_kw, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		realtyID, userId, tariffId, gaugeId, readingDate, dayKw, nightKw, createdAt,
	)
	return err
}

func (r *ClickHouseRepo) AddGaugeBinding(realtyID, gaugeId int64) error {
	var activeGaugeId uint64
	var maxVer uint64
	err := r.db.QueryRow(`
		SELECT gauge_id, version FROM energosbyt.fact_gauge_bindings FINAL 
		WHERE realty_id = ? GROUP BY gauge_id, version HAVING sum(sign) > 0`, realtyID).Scan(&activeGaugeId, &maxVer)

	now := time.Now().Format("2006-01-02 15:04:05")
	if err == nil && activeGaugeId != 0 {
		_, _ = r.db.Exec(`
			INSERT INTO energosbyt.fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) 
			VALUES (?, ?, ?, -1, ?)`, activeGaugeId, realtyID, now, maxVer)
		maxVer++
	} else {
		maxVer = 1
	}

	_, err = r.db.Exec(`
		INSERT INTO energosbyt.fact_gauge_bindings (gauge_id, realty_id, created_at, sign, version) 
		VALUES (?, ?, ?, 1, ?)`, gaugeId, realtyID, now, maxVer)
	return err
}

func (r *ClickHouseRepo) GetTariffByDate(dateStr string) (*models.Tariff, error) {
	var t models.Tariff
	var validFrom, validTo time.Time
	err := r.db.QueryRow(`
		SELECT id, name, price_day_kwph, price_night_kwph, valid_from, valid_to 
		FROM energosbyt.dim_tariffs FINAL 
		WHERE valid_from <= ? AND valid_to > ? LIMIT 1`, dateStr+" 00:00:00", dateStr+" 00:00:00").Scan(
		&t.ID, &t.Name, &t.PriceDayKwph, &t.PriceNightKwph, &validFrom, &validTo,
	)
	if err != nil {
		return nil, err
	}
	t.ValidFrom = validFrom.Format("2006-01-02 15:04:05")
	t.ValidTo = validTo.Format("2006-01-02 15:04:05")
	return &t, nil
}

func (r *ClickHouseRepo) GetInvoices(realtyID *int64, limit, offset int) ([]models.Invoice, error) {
	query := "SELECT realty_id, user_id, tariff_id, gauge_id, day_kw, night_kw, price, deadline, created_at FROM energosbyt.fact_invoice"
	var args []interface{}

	if realtyID != nil {
		query += " WHERE realty_id = ?"
		args = append(args, *realtyID)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []models.Invoice
	for rows.Next() {
		var inv models.Invoice
		var deadline time.Time
		var createdAt time.Time
		if err = rows.Scan(&inv.RealtyID, &inv.UserID, &inv.TariffID, &inv.GaugeID, &inv.DayKw, &inv.NightKw, &inv.Price, &deadline, &createdAt); err != nil {
			return nil, err
		}
		inv.Deadline = deadline.Format("2006-01-02")
		inv.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		invoices = append(invoices, inv)
	}

	if invoices == nil {
		invoices = []models.Invoice{}
	}
	return invoices, nil
}

func (r *ClickHouseRepo) GetInvoicesCount(realtyID *int64) (int, error) {
	query := "SELECT count(*) FROM energosbyt.fact_invoice"
	var args []interface{}
	if realtyID != nil {
		query += " WHERE realty_id = ?"
		args = append(args, *realtyID)
	}
	var count uint64
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
