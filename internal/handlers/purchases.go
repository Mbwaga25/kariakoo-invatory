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
	
	// Get optional pre-selected product
	preselectedID, _ := strconv.Atoi(r.URL.Query().Get("product_id"))
	
	products, _ := app.Models.GetProductsByTenant(tenantID, 0, 0, 0)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)
	suppliers, _ := app.Models.GetContactsByTenant(tenantID, "supplier")

	app.RenderPage(w, r, "purchases/create", struct {
		Products             []*models.Product
		Locations            []*models.BusinessLocation
		Suppliers            []*models.Contact
		PreselectedProductID int
	}{
		Products:             products,
		Locations:            locations,
		Suppliers:            suppliers,
		PreselectedProductID: preselectedID,
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

	p.FinalTotal = 0
	
	productIDs := r.Form["product_id[]"]
	quantities := r.Form["quantity[]"]
	prices := r.Form["purchase_price[]"]

	for i, pidStr := range productIDs {
		productID, _ := strconv.Atoi(pidStr)
		qty, _ := strconv.ParseFloat(quantities[i], 64)
		price, _ := strconv.ParseFloat(prices[i], 64)

		if productID > 0 && qty > 0 {
			item := &models.PurchaseItem{
				ProductID:     productID,
				Quantity:      qty,
				PurchasePrice: price,
				LineTotal:     qty * price,
			}
			p.Items = append(p.Items, item)
			p.FinalTotal += item.LineTotal
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
