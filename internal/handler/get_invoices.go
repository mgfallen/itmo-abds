package handler

import (
	"net/http"
	"strconv"
)

func (h *Handler) HandleGetInvoices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, "Not found", 404)
		return
	}

	limit := 10
	offset := 0
	var realtyID *int64

	if l := r.URL.Query().Get("limit"); l != "" {
		parsedLimit, err := strconv.Atoi(l)
		if err != nil || parsedLimit <= 0 {
			httpError(w, "invalid \"limit\" parameter", 400)
			return
		}
		limit = parsedLimit
	}

	if p := r.URL.Query().Get("page"); p != "" {
		page, err := strconv.Atoi(p)
		if err != nil || page <= 0 {
			httpError(w, "invalid \"page\" parameter", 400)
			return
		}
		offset = (page - 1) * limit
	}

	if rid := r.URL.Query().Get("realty_id"); rid != "" {
		parsedRID, err := strconv.ParseInt(rid, 10, 64)
		if err != nil {
			httpError(w, "invalid \"realty_id\" parameter", 400)
			return
		}
		realtyID = &parsedRID
	}

	invoices, err := h.repo.GetInvoices(realtyID, limit, offset)
	if err != nil {
		httpError(w, err.Error(), 500)
		return
	}

	writeJSON(w, invoices, 200)
}
