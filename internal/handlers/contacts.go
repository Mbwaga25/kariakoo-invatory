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
	tenantID := middleware.GetTenantID(r.Context())
	contactType := r.URL.Query().Get("type")
	settings, _ := app.Models.GetBusinessSettings(tenantID)

	app.RenderPage(w, r, "contacts/create", struct {
		ContactType string
		Settings    *models.BusinessSetting
	}{
		ContactType: contactType,
		Settings:    settings,
	})
}

func (app *Application) ContactEdit(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	contact, err := app.Models.GetContactByID(id, tenantID)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	app.RenderPage(w, r, "contacts/edit", struct {
		Contact *models.Contact
	}{
		Contact: contact,
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
	
	var creditLimit *float64
	creditLimitStr := r.FormValue("credit_limit")
	if creditLimitStr != "" {
		val, err := strconv.ParseFloat(creditLimitStr, 64)
		if err == nil {
			creditLimit = &val
		}
	}

	var invoiceDueDays *int
	dueDaysStr := r.FormValue("invoice_due_days")
	if dueDaysStr != "" {
		val, err := strconv.Atoi(dueDaysStr)
		if err == nil {
			invoiceDueDays = &val
		}
	}

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
		CreditLimit:    creditLimit,
		InvoiceDueDays: invoiceDueDays,
	}

	_, err := app.Models.InsertContact(c)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/contacts?type="+c.Type, http.StatusSeeOther)
}

func (app *Application) ContactUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))

	r.ParseForm()

	balance, _ := strconv.ParseFloat(r.FormValue("opening_balance"), 64)
	
	var creditLimit *float64
	creditLimitStr := r.FormValue("credit_limit")
	if creditLimitStr != "" {
		val, err := strconv.ParseFloat(creditLimitStr, 64)
		if err == nil {
			creditLimit = &val
		}
	}

	var invoiceDueDays *int
	dueDaysStr := r.FormValue("invoice_due_days")
	if dueDaysStr != "" {
		val, err := strconv.Atoi(dueDaysStr)
		if err == nil {
			invoiceDueDays = &val
		}
	}

	c := &models.Contact{
		ID:             id,
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
		CreditLimit:    creditLimit,
		InvoiceDueDays: invoiceDueDays,
	}

	err := app.Models.UpdateContact(c)
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

func (app *Application) ContactCreditCheck(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	name := r.URL.Query().Get("name")
	if name == "" {
		app.jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	contact, err := app.Models.GetContactByName(tenantID, name)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"exists":       false,
			"credit_limit": 0,
			"balance":      0,
		})
		return
	}

	balance, err := app.Models.GetCustomerBalance(tenantID, name)
	if err != nil {
		app.jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	creditLimit := 0.0
	if contact.CreditLimit != nil {
		creditLimit = *contact.CreditLimit
	}

	app.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"exists":       true,
		"name":         contact.Name,
		"credit_limit": creditLimit,
		"balance":      balance,
	})
}
