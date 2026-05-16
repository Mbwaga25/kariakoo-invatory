package handlers

import (
	"net/http"
	"strconv"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) PurchaseList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())
	
	purchases, err := app.Models.GetPurchasesByTenant(tenantID, locationID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "purchases/index", struct {
		Purchases []*models.Purchase
	}{
		Purchases: purchases,
	})
}

func (app *Application) PurchaseCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	products, _ := app.Models.GetProductsByTenantFiltered(tenantID, 0, "", 0, 0)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)
	categories, _ := app.Models.GetCategoriesByTenant(tenantID)
	brands, _ := app.Models.GetBrandsByTenant(tenantID)

	app.RenderPage(w, r, "purchases/create", struct {
		Products   []*models.Product
		Locations  []*models.BusinessLocation
		Categories []*models.Category
		Brands     []*models.Brand
	}{
		Products:   products,
		Locations:  locations,
		Categories: categories,
		Brands:     brands,
	})
}


func (app *Application) PurchaseStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/purchases/create", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	r.ParseForm()

	locationID, _ := strconv.Atoi(r.FormValue("location_id"))
	refNo := r.FormValue("ref_no")
	
	p := &models.Purchase{
		TenantID:           tenantID,
		BusinessLocationID: locationID,
		RefNo:              refNo,
		PurchaseDate:       time.Now(),
		Status:             "received",
		PaymentStatus:      "paid",
		CreatedBy:          &user.ID,
	}

	// No supplier needed
	p.FinalTotal = 0
	
	productIDs := r.Form["product_id[]"]
	quantities := r.Form["quantity[]"]

	for i, pidStr := range productIDs {
		productID, _ := strconv.Atoi(pidStr)
		qty, _ := strconv.ParseFloat(quantities[i], 64)

		if productID > 0 && qty > 0 {
			item := &models.PurchaseItem{
				ProductID:     productID,
				Quantity:      qty,
				PurchasePrice: 0, // No price tracking
				LineTotal:     0,
			}
			p.Items = append(p.Items, item)
		}
	}

	if len(p.Items) == 0 {
		http.Error(w, "At least one product is required", http.StatusBadRequest)
		return
	}

	_, err := app.Models.InsertPurchase(p)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/purchases", http.StatusSeeOther)
}

func (app *Application) PurchaseView(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	purchase, err := app.Models.GetPurchaseByID(id, tenantID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	app.RenderPage(w, r, "purchases/view", struct {
		Purchase *models.Purchase
	}{
		Purchase: purchase,
	})
}

