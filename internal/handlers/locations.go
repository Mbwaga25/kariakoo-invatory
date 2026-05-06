package handlers

import (
	"log"
	"net/http"
	"strconv"
	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) LocationSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	locations, err := app.Models.GetLocationsByTenant(tenantID)
	if err != nil {
		log.Printf("ERROR GetLocationsByTenant: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "admin/locations", struct {
		Locations []*models.BusinessLocation
	}{
		Locations: locations,
	})
}

func (app *Application) LocationStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/business-location", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	
	l := &models.BusinessLocation{
		TenantID:   tenantID,
		Name:       r.FormValue("name"),
		LocationID: r.FormValue("location_id"),
		City:       r.FormValue("city"),
		State:      r.FormValue("state"),
		Country:    r.FormValue("country"),
		ZipCode:    r.FormValue("zip_code"),
	}

	_, err := app.Models.InsertLocation(l)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/business-location", http.StatusSeeOther)
}

func (app *Application) LocationUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/business-location", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))
	
	l := &models.BusinessLocation{
		ID:         id,
		TenantID:   tenantID,
		Name:       r.FormValue("name"),
		LocationID: r.FormValue("location_id"),
		City:       r.FormValue("city"),
		State:      r.FormValue("state"),
		Country:    r.FormValue("country"),
		ZipCode:    r.FormValue("zip_code"),
	}

	err := app.Models.UpdateLocation(l)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/business-location", http.StatusSeeOther)
}

func (app *Application) LocationDelete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	err := app.Models.DeleteLocation(id, tenantID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/business-location", http.StatusSeeOther)
}

func (app *Application) BusinessSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	settings, _ := app.Models.GetBusinessSettings(tenantID)
	if settings == nil {
		settings = &models.BusinessSetting{
			BusinessName: "My Shop",
			Currency:     "USD",
		}
	}

	app.RenderPage(w, r, "admin/settings", struct {
		Settings *models.BusinessSetting
	}{
		Settings: settings,
	})
}

func (app *Application) BusinessSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/business-settings", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	
	s := &models.BusinessSetting{
		TenantID:           tenantID,
		BusinessName:       r.FormValue("business_name"),
		Currency:           r.FormValue("currency"),
		CurrencySymbol:     r.FormValue("currency_symbol"),
		TimeZone:           r.FormValue("time_zone"),
		TaxNumber:          r.FormValue("tax_number"),
		TaxName:            r.FormValue("tax_name"),
		FinancialYearStart: r.FormValue("financial_year_start"),
		StockExpirySetting: r.FormValue("stock_expiry_setting"),
	}

	err := app.Models.UpdateBusinessSettings(s)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/business-settings", http.StatusSeeOther)
}

func (app *Application) LocationSwitch(w http.ResponseWriter, r *http.Request) {
	locationID := r.URL.Query().Get("location_id")
	if locationID != "" {
		cookie := http.Cookie{
			Name:     "location_id",
			Value:    locationID,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   3600 * 24 * 30, // 30 days
		}
		http.SetCookie(w, &cookie)
	}

	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}
