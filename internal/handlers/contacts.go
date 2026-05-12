package handlers

import (
	"net/http"
	"strconv"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) ContactList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	contactType := r.URL.Query().Get("type")
	
	contacts, err := app.Models.GetContactsByTenant(tenantID, contactType)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "contacts/index", struct {
		Contacts    []*models.Contact
		ContactType string
	}{
		Contacts:    contacts,
		ContactType: contactType,
	})
}

func (app *Application) ContactCreate(w http.ResponseWriter, r *http.Request) {
	contactType := r.URL.Query().Get("type")
	app.RenderPage(w, r, "contacts/create", struct {
		ContactType string
	}{
		ContactType: contactType,
	})
}

func (app *Application) ContactStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	r.ParseForm()

	balance, _ := strconv.ParseFloat(r.FormValue("opening_balance"), 64)
	
	c := &models.Contact{
		TenantID:       tenantID,
		Type:           r.FormValue("type"),
		Name:           r.FormValue("name"),
		BusinessName:   r.FormValue("business_name"),
		Email:          r.FormValue("email"),
		Mobile:         r.FormValue("mobile"),
		TaxNumber:      r.FormValue("tax_number"),
		OpeningBalance: balance,
		Address:        r.FormValue("address"),
		City:           r.FormValue("city"),
		CreatedBy:      &user.ID,
	}

	_, err := app.Models.InsertContact(c)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/contacts?type="+c.Type, http.StatusSeeOther)
}

func (app *Application) ContactStoreQuick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"success": "false", "msg": "Method not allowed"})
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	r.ParseForm()
	
	c := &models.Contact{
		TenantID:  tenantID,
		Type:      r.FormValue("type"),
		Name:      r.FormValue("name"),
		Mobile:    r.FormValue("mobile"),
		CreatedBy: &user.ID,
	}

	if c.Name == "" {
		app.jsonResponse(w, http.StatusOK, map[string]string{"success": "false", "msg": "Name is required"})
		return
	}

	id, err := app.Models.InsertContact(c)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]string{"success": "false", "msg": err.Error()})
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success": "true",
		"id":      id,
		"name":    c.Name,
		"msg":     "Customer added successfully",
	})
}
