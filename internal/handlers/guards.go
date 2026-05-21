package handlers

import (
	"net/http"

	"kariakoo/inventory/internal/middleware"
)

func (app *Application) requireRoles(w http.ResponseWriter, r *http.Request, allowed ...string) bool {
	user := middleware.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	if len(allowed) == 0 {
		return true
	}

	for _, role := range allowed {
		if user.Role == role {
			return true
		}
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
	return false
}
