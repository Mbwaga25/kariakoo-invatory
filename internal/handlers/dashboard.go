package handlers

import (
	"log"
	"net/http"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	
	// Default to today
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	if startStr != "" && endStr != "" {
		if s, err := time.Parse("2006-01-02", startStr); err == nil {
			start = s
		}
		if e, err := time.Parse("2006-01-02", endStr); err == nil {
			end = time.Date(e.Year(), e.Month(), e.Day(), 23, 59, 59, 0, e.Location())
		}
	}

	// Get dashboard data
	dashboardData, err := app.Models.GetDashboardData(tenantID, nil, start, end)
	if err != nil {
		log.Printf("ERROR GetDashboardData: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "dashboard/index", struct {
		Data  *models.DashboardData
		Start time.Time
		End   time.Time
	}{
		Data:  dashboardData,
		Start: start,
		End:   end,
	})
}
