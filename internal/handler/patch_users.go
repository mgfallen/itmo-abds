package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandlePatchUser(w http.ResponseWriter, r *http.Request, id int64) {
	var req struct {
		Fio   *string `json:"fio"`
		Phone *string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (req.Fio == nil && req.Phone == nil) {
		httpError(w, "invalid request body", 400)
		return
	}
	if req.Fio != nil && *req.Fio == "" {
		httpError(w, "invalid \"fio\" field", 400)
		return
	}
	if req.Phone != nil && *req.Phone == "" {
		httpError(w, "invalid \"phone\" field", 400)
		return
	}

	err := h.repo.UpdateUser(id, req.Fio, req.Phone)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
