package handlers

import (
	"net/http"
	"strconv"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) StockAdjustmentList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())
	start, end := app.ParseDateRange(r)
	
	adjustments, err := app.Models.GetStockAdjustmentsByTenant(tenantID, locationID, start, end)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "stock_adjustments/index", struct {
		Adjustments []*models.StockAdjustment
		Start       time.Time
		End         time.Time
	}{
		Adjustments: adjustments,
		Start:       start,
		End:         end,
	})
}

func (app *Application) StockAdjustmentCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	products, _ := app.Models.GetProductsByTenant(tenantID, 0, 0, 0)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)

	app.RenderPage(w, r, "stock_adjustments/create", struct {
		Products  []*models.Product
		Locations []*models.BusinessLocation
	}{
		Products:  products,
		Locations: locations,
	})
}

func (app *Application) StockAdjustmentStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/stock-adjustments/create", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	r.ParseForm()

	locationID, _ := strconv.Atoi(r.FormValue("location_id"))
	refNo := r.FormValue("ref_no")
	adjType := r.FormValue("adjustment_type")
	
	sa := &models.StockAdjustment{
		TenantID:           tenantID,
		BusinessLocationID: locationID,
		RefNo:              refNo,
		TransactionDate:    time.Now(),
		AdjustmentType:     adjType,
		CreatedBy:          &user.ID,
	}

	productIDs := r.Form["product_id[]"]
	quantities := r.Form["quantity[]"]

	for i, pidStr := range productIDs {
		productID, _ := strconv.Atoi(pidStr)
		qty, _ := strconv.ParseFloat(quantities[i], 64)

		if productID > 0 && qty > 0 {
			item := &models.StockAdjustmentItem{
				ProductID: productID,
				Quantity:  qty,
			}
			sa.Items = append(sa.Items, item)
		}
	}

	_, err := app.Models.InsertStockAdjustment(sa)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/stock-adjustments", http.StatusSeeOther)
}

