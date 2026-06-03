package handlers

import (
	"log"
	"net/http"
	"time"
	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/dashboard" {
		http.NotFound(w, r)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())
	locationID := middleware.GetLocationID(r.Context())
	
	start, end := app.ParseDateRange(r)
	preset := r.URL.Query().Get("preset")
	if preset == "" {
		if r.URL.Query().Get("start_date") != "" && r.URL.Query().Get("end_date") != "" {
			preset = "custom"
		} else {
			preset = "today"
		}
	}

	// Get dashboard data (Location specific)
	dashboardData, err := app.Models.GetDashboardData(tenantID, &locationID, start, end)
	if err != nil {
		log.Printf("ERROR GetDashboardData: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get order summary (Location specific)
	role := ""
	if user != nil {
		role = user.Role
	}
	orderSummary, _ := app.Models.GetOrderSummary(tenantID, role, locationID, user.ID, start, end)

	// Get low stock alerts (Location specific)
	lowStockProducts, _ := app.Models.GetLowStockProductsByLocation(tenantID, locationID)

	// Get best selling products
	bestSelling, _ := app.Models.GetBestSellingProducts(tenantID, 5)

	app.RenderPage(w, r, "dashboard/index", struct {
		Data            *models.DashboardData
		OrderSummary    *models.OrderSummary
		LowStock        []*models.Product
		BestSelling     []*models.Product
		Start           time.Time
		End             time.Time
		Preset          string
		StartDate       string
		EndDate         string
	}{
		Data:         dashboardData,
		OrderSummary: orderSummary,
		LowStock:     lowStockProducts,
		BestSelling:  bestSelling,
		Start:        start,
		End:          end,
		Preset:       preset,
		StartDate:    start.Format("2006-01-02"),
		EndDate:      end.Format("2006-01-02"),
	})
}
