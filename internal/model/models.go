package models

type User struct {
	ID        int64  `json:"id"`
	Fio       string `json:"fio"`
	Phone     string `json:"phone"`
	PhonePrev string `json:"phone_prev"`
}

type Realty struct {
	ID      int64  `json:"id"`
	Address string `json:"address"`
	Square  uint32 `json:"square"`
	Type    string `json:"type"`
	Number  string `json:"number"`
	Rooms   uint8  `json:"rooms"`
	Floor   uint8  `json:"floor"`
	Porch   uint8  `json:"porch"`
}

type Reading struct {
	ReadingDate string `json:"reading_date"`
	UserId      int64  `json:"user_id"`
	TariffId    int64  `json:"tariff_id"`
	GaugeId     int64  `json:"gauge_id"`
	DayKw       uint32 `json:"day_kw"`
	NightKw     uint32 `json:"night_kw"`
}

type Tariff struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	PriceDayKwph   uint32 `json:"price_day_kwph"`
	PriceNightKwph uint32 `json:"price_night_kwph"`
	ValidFrom      string `json:"valid_from"`
	ValidTo        string `json:"valid_to"`
}

type Invoice struct {
	RealtyID  uint64 `json:"realty_id"`
	UserID    uint64 `json:"user_id"`
	TariffID  uint64 `json:"tariff_id"`
	GaugeID   uint64 `json:"gauge_id"`
	DayKw     uint32 `json:"day_kw"`
	NightKw   uint32 `json:"night_kw"`
	Price     uint64 `json:"price"`
	Deadline  string `json:"deadline"`
	CreatedAt string `json:"created_at"`
}
