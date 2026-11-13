package handler

import (
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) HandleUsers(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/users/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		httpError(w, "Not found", 404)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.HandleGetUsers(w, r, id)
	case http.MethodPatch:
		h.HandlePatchUser(w, r, id)
	default:
		httpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleGetUsers(w http.ResponseWriter, r *http.Request, id int64) {
	user, err := h.repo.GetUser(id)
	if err != nil {
		httpError(w, "Not found", 404)
		return
	}
	writeJSON(w, user, 200)
}
