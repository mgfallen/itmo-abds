package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandlePostGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, "Not found", 404)
		return
	}

	var req struct {
		GaugeId int64  `json:"gauge_id"`
		Name    string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.GaugeId == 0 && req.Name == "") {
		httpError(w, "invalid request body", 400)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
