package handler

import (
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) HandleRealty(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/realty/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		httpError(w, "Not found", 404)
		return
	}

	realtyID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			h.HandleGetRealty(w, r, realtyID)
			return
		}
		httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sub := parts[1]
	switch sub {
	case "readings":
		if r.Method == http.MethodGet {
			h.HandleGetReadings(w, r, realtyID)
		} else if r.Method == http.MethodPost {
			h.HandlePostReading(w, r, realtyID)
		} else {
			httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "gauges":
		if r.Method == http.MethodPost {
			h.HandlePostGaugeBinding(w, r, realtyID)
		} else {
			httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "invoices":
		if r.Method == http.MethodGet {
			h.HandleGetInvoicesForRealty(w, r, realtyID)
		} else {
			httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		httpError(w, "Not found", 404)
	}
}

func (h *Handler) HandleGetRealty(w http.ResponseWriter, r *http.Request, realtyID int64) {
	realty, err := h.repo.GetRealty(realtyID)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}
	writeJSON(w, realty, 200)
}

func (h *Handler) HandleGetInvoicesForRealty(w http.ResponseWriter, r *http.Request, realtyID int64) {
	_, err := h.repo.GetRealty(realtyID)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}

	limit := 10
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err != nil || parsedLimit <= 0 {
			httpError(w, "invalid \"limit\" parameter", 400)
			return
		}
		limit = parsedLimit
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		parsedOffset, err := strconv.Atoi(o)
		if err != nil || parsedOffset < 0 {
			httpError(w, "invalid \"offset\" parameter", 400)
			return
		}
		offset = parsedOffset
	}

	invoices, err := h.repo.GetInvoices(&realtyID, limit, offset)
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}

	count, err := h.repo.GetInvoicesCount(&realtyID)
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}

	writeJSON(w, map[string]interface{}{
		"invoices": invoices,
		"count":    count,
	}, 200)
}
