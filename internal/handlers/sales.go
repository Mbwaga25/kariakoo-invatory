package handlers

import (
	"log"
	"net/http"
	"strconv"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) SalesList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())
	
	sales, err := app.Models.GetSalesByTenant(tenantID, locationID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "sales/index", struct {
		Sales []*models.Sale
	}{
		Sales: sales,
	})
}

func (app *Application) SalesCreate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	tenantID := middleware.GetTenantID(r.Context())
	
	locations, err := app.Models.GetLocationsByTenant(tenantID)
	if err != nil {
		log.Printf("ERROR GetLocationsByTenant (SalesCreate): %v", err)
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Check if any locations exist
	if len(locations) == 0 {
		app.RenderPage(w, r, "error", struct {
			Message string
		}{
			Message: "No shop locations found. Please add a business location in settings before opening the POS.",
		})
		return
	}

	// Default location if only one
	var selectedLocationID int
	if len(locations) == 1 {
		selectedLocationID = locations[0].ID
	} else {
		selectedLocationID, _ = strconv.Atoi(r.URL.Query().Get("location_id"))
	}

	// If multiple locations and none selected, show location selector
	if len(locations) > 1 && selectedLocationID == 0 {
		app.RenderPage(w, r, "sales/select_location", struct {
			Locations []*models.BusinessLocation
		}{
			Locations: locations,
		})
		return
	}

	// Check for open register
	register, _ := app.Models.GetOpenRegister(user.ID, selectedLocationID)
	if register == nil {
		app.RenderPage(w, r, "sales/open_register", struct {
			LocationID int
		}{
			LocationID: selectedLocationID,
		})
		return
	}
	products, _ := app.Models.GetProductsByTenant(tenantID, selectedLocationID, 0, 0)

	app.RenderPage(w, r, "sales/create", struct {
		Products   []*models.Product
		LocationID int
	}{
		Products:   products,
		LocationID: selectedLocationID,
	})
}

func (app *Application) RegisterOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/pos/create", http.StatusSeeOther)
		return
	}

	user := middleware.GetUser(r.Context())
	tenantID := middleware.GetTenantID(r.Context())
	locationID, _ := strconv.Atoi(r.FormValue("location_id"))
	openingAmount, _ := strconv.ParseFloat(r.FormValue("opening_amount"), 64)

	if locationID <= 0 {
		http.Error(w, "Invalid Location ID. Please select a valid shop location.", http.StatusBadRequest)
		return
	}

	reg := &models.CashRegister{
		TenantID:           tenantID,
		BusinessLocationID: locationID,
		UserID:             user.ID,
		OpeningAmount:      openingAmount,
	}

	err := app.Models.OpenRegister(reg)
	if err != nil {
		log.Printf("ERROR OpenRegister: %v (Tenant: %d, Location: %d, User: %d)", err, tenantID, locationID, user.ID)
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/pos/create?location_id="+strconv.Itoa(locationID), http.StatusSeeOther)
}

