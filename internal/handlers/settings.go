package handlers

import (
	"net/http"
	"kariakoo/inventory/internal/middleware"
)

func (app *Application) SettingsModules(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		
		if action == "install_all" {
			allModuleKeys := []string{"sales", "purchases", "stock_adjustments", "stock_transfers", "expenses", "reports"}
			for _, k := range allModuleKeys {
				app.Models.ToggleModule(tenantID, k, true)
			}
		} else {
			moduleKey := r.FormValue("module_key")
			status := action == "install"
			
			err := app.Models.ToggleModule(tenantID, moduleKey, status)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			// If uninstall and "clear_data" is checked
			if action == "uninstall" && r.FormValue("clear_data") == "1" {
				app.ClearModuleData(tenantID, moduleKey)
			}
		}

		http.Redirect(w, r, "/settings/modules", http.StatusSeeOther)
		return
	}

	modules, err := app.Models.GetTenantModules(tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "settings/modules", struct {
		Modules interface{}
	}{
		Modules: modules,
	})
}

func (app *Application) ClearModuleData(tenantID int, moduleKey string) {
	switch moduleKey {
	case "sales":
		app.DB.Exec("DELETE FROM sales WHERE tenant_id = ?", tenantID)
	case "purchases":
		app.DB.Exec("DELETE FROM purchases WHERE tenant_id = ?", tenantID)
	case "expenses":
		app.DB.Exec("DELETE FROM expenses WHERE tenant_id = ?", tenantID)
	case "stock_adjustments":
		app.DB.Exec("DELETE FROM stock_adjustments WHERE tenant_id = ?", tenantID)
	case "stock_transfers":
		app.DB.Exec("DELETE FROM stock_transfers WHERE tenant_id = ?", tenantID)
	}
}
