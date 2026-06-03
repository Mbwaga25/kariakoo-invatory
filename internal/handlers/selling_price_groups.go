package handlers

import (
	"net/http"
	"strconv"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) SellingPriceGroupList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	groups, err := app.Models.GetSellingPriceGroupsByTenant(tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "admin/selling_groups", struct {
		Groups []*models.SellingPriceGroup
	}{
		Groups: groups,
	})
}

func (app *Application) SellingPriceGroupStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/selling-price-groups", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())

	r.ParseForm()
	g := &models.SellingPriceGroup{
		TenantID:    tenantID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	if g.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	_, err := app.Models.InsertSellingPriceGroup(g)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/selling-price-groups", http.StatusSeeOther)
}

func (app *Application) SellingPriceGroupUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/selling-price-groups", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))

	r.ParseForm()
	g := &models.SellingPriceGroup{
		ID:          id,
		TenantID:    tenantID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	if g.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	err := app.Models.UpdateSellingPriceGroup(g)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/selling-price-groups", http.StatusSeeOther)
}

func (app *Application) SellingPriceGroupDelete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	err := app.Models.DeleteSellingPriceGroup(id, tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/selling-price-groups", http.StatusSeeOther)
}
