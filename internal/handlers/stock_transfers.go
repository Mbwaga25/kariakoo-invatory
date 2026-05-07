package handlers

import (
	"net/http"
	"strconv"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) StockTransferList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	transfers, err := app.Models.GetStockTransfersByTenant(tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "stock_transfers/index", struct {
		Transfers []*models.StockTransfer
	}{
		Transfers: transfers,
	})
}

func (app *Application) StockTransferCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	products, _ := app.Models.GetProductsByTenant(tenantID, 0, 0, 0)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)

	app.RenderPage(w, r, "stock_transfers/create", struct {
		Products  []*models.Product
		Locations []*models.BusinessLocation
	}{
		Products:  products,
		Locations: locations,
	})
}

func (app *Application) StockTransferStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/stock-transfers/create", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	r.ParseForm()

	fromLocationID, _ := strconv.Atoi(r.FormValue("from_location_id"))
	toLocationID, _ := strconv.Atoi(r.FormValue("to_location_id"))
	refNo := r.FormValue("ref_no")
	
	st := &models.StockTransfer{
		TenantID:        tenantID,
		FromLocationID:  fromLocationID,
		ToLocationID:    toLocationID,
		RefNo:           refNo,
		TransactionDate: time.Now(),
		Status:          "received",
		CreatedBy:       user.ID,
	}

	productIDs := r.Form["product_id[]"]
	quantities := r.Form["quantity[]"]

	for i, pidStr := range productIDs {
		productID, _ := strconv.Atoi(pidStr)
		qty, _ := strconv.ParseFloat(quantities[i], 64)

		if productID > 0 && qty > 0 {
			item := &models.StockTransferItem{
				ProductID: productID,
				Quantity:  qty,
			}
			st.Items = append(st.Items, item)
		}
	}

	_, err := app.Models.InsertStockTransfer(st)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/stock-transfers", http.StatusSeeOther)
}

