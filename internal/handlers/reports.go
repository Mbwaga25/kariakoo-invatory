package handlers

import (
	"net/http"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) ProfitLossReport(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	start, end := app.ParseDateRange(r)

	report, err := app.Models.GetProfitLossReport(tenantID, start, end)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "reports/profit_loss", struct {
		Report *models.ProfitLossReport
		Start  time.Time
		End    time.Time
	}{
		Report: report,
		Start:  start,
		End:    end,
	})
}

func (app *Application) StockReport(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	reports, err := app.Models.GetStockReport(tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "reports/stock", struct {
		StockReports []*models.StockReport
	}{
		StockReports: reports,
	})
}

func (app *Application) RegisterReport(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	reports, err := app.Models.GetRegisterReport(tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "reports/register", struct {
		RegisterReports []*models.RegisterReport
	}{
		RegisterReports: reports,
	})
}

func (app *Application) PurchaseSellReport(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	start, end := app.ParseDateRange(r)

	report, err := app.Models.GetPurchaseSellReport(tenantID, start, end)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "reports/purchase_sell", struct {
		Report *models.PurchaseSellReport
		Start  time.Time
		End    time.Time
	}{
		Report: report,
		Start:  start,
		End:    end,
	})
}

func (app *Application) ExpenseReport(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	start, end := app.ParseDateRange(r)

	reports, err := app.Models.GetExpenseReport(tenantID, start, end)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var total float64
	for _, exp := range reports {
		total += exp.TotalAmount
	}

	app.RenderPage(w, r, "reports/expense", struct {
		Expenses []*models.ExpenseReport
		Total    float64
		Start    time.Time
		End      time.Time
	}{
		Expenses: reports,
		Total:    total,
		Start:    start,
		End:      end,
	})
}
