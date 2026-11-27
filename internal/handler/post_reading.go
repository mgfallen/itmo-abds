package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandlePostReading(w http.ResponseWriter, r *http.Request, realtyID int64) {
	if r.Method != http.MethodPost {
		httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserId      int64  `json:"user_id"`
		GaugeId     int64  `json:"gauge_id"`
		TariffId    int64  `json:"tariff_id"`
		ReadingDate string `json:"reading_date"`
		DayKw       uint32 `json:"day_kw"`
		NightKw     uint32 `json:"night_kw"`
	}

	// Ошибки парсинга JSON строго дают 400 Bad Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, "invalid request body", 400)
		return
	}

	if req.UserId == 0 || req.GaugeId == 0 || req.TariffId == 0 || req.ReadingDate == "" {
		httpError(w, "invalid request fields", 400)
		return
	}

	// Проверяем существование недвижимости после валидации запроса
	_, err := h.repo.GetRealty(realtyID)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}

	err = h.repo.AddReading(realtyID, req.UserId, req.GaugeId, req.TariffId, req.ReadingDate, req.DayKw, req.NightKw)
	if err != nil {
		httpError(w, err.Error(), 400)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
