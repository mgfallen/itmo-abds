package handler

import "net/http"

func (h *Handler) HandleGetPing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
