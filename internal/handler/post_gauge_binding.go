package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandlePostGaugeBinding(w http.ResponseWriter, r *http.Request, realtyID int64) {
	if r.Method != http.MethodPost {
		httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GaugeId int64 `json:"gauge_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GaugeId == 0 {
		httpError(w, "invalid \"gauge_id\" field", 400)
		return
	}

	_, err := h.repo.GetRealty(realtyID)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}

	err = h.repo.AddGaugeBinding(realtyID, req.GaugeId)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
