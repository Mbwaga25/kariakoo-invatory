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
	
	products, _ := app.Models.GetProductsByTenant(tenantID, 0, 0, 0)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)
	suppliers, _ := app.Models.GetContactsByTenant(tenantID, "supplier")

	app.RenderPage(w, r, "purchases/create", struct {
		Products  []*models.Product
		Locations []*models.BusinessLocation
		Suppliers []*models.Contact
	}{
		Products:  products,
		Locations: locations,
		Suppliers: suppliers,
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
	supplierID, _ := strconv.Atoi(r.FormValue("supplier_id"))
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

	if supplierID > 0 {
		p.SupplierID = &supplierID
	}

	// Simplified: In a real app, we would loop through items from the form
	// For now, let's assume one item or a simplified structure
	// This is a placeholder for the actual form processing logic
	
	productID, _ := strconv.Atoi(r.FormValue("product_id"))
	qty, _ := strconv.ParseFloat(r.FormValue("quantity"), 64)
	price, _ := strconv.ParseFloat(r.FormValue("purchase_price"), 64)

	if productID > 0 && qty > 0 {
		item := &models.PurchaseItem{
			ProductID:     productID,
			Quantity:      qty,
			PurchasePrice: price,
			LineTotal:     qty * price,
		}
		p.Items = append(p.Items, item)
		p.FinalTotal = item.LineTotal
	}

	_, err := app.Models.InsertPurchase(p)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/purchases", http.StatusSeeOther)
}
