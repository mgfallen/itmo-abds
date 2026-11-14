package handler

import (
	"encoding/json"
	"net/http"

	"energosbyt/internal/repository"
)

type Handler struct {
	repo repository.Repository
}

func NewHandler(repo repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ping", h.HandleGetPing)
	mux.HandleFunc("/users/", h.HandleUsers)
	mux.HandleFunc("/realty/", h.HandleRealty)
	mux.HandleFunc("/tariffs", h.HandleGetTariff)
	mux.HandleFunc("/invoices", h.HandleGetInvoices)
	mux.HandleFunc("/gauges", h.HandlePostGauge)
}

func httpError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func writeJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}
