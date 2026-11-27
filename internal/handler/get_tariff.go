package handler

import (
	"net/http"
	"time"
)

func (h *Handler) HandleGetTariff(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		httpError(w, "invalid \"date\" parameter", 400)
		return
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		httpError(w, "invalid \"date\" parameter format", 400)
		return
	}

	tariff, err := h.repo.GetTariffByDate(dateStr)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}

	writeJSON(w, tariff, 200)
}
