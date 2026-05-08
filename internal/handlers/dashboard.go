package handlers

import (
	"net/http"
	"kariakoo/inventory/internal/middleware"
)

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	
	// Get dashboard data
	dashboardData, err := app.Models.GetDashboardData(tenantID, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "dashboard/index", dashboardData)
}
