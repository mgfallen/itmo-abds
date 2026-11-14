package handler

import (
	"net/http"
	"time"
)

func (h *Handler) HandleGetReadings(w http.ResponseWriter, r *http.Request, realtyID int64) {
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")
	if dateFrom == "" || dateTo == "" {
		httpError(w, "invalid \"date_from\" or \"date_to\" parameter", 400)
		return
	}

	if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
		httpError(w, "invalid \"date_from\" format", 400)
		return
	}
	if _, err := time.Parse("2006-01-02", dateTo); err != nil {
		httpError(w, "invalid \"date_to\" format", 400)
		return
	}

	readings, err := h.repo.GetReadings(realtyID, dateFrom, dateTo)
	if err != nil {
		httpError(w, err.Error(), 400)
		return
	}
	if len(readings) == 0 {
		httpError(w, "readings not found", 404)
		return
	}
	writeJSON(w, map[string]interface{}{
		"readings": readings,
		"count":    len(readings),
	}, 200)
}
