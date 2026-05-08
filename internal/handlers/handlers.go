package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

type Application struct {
	DB     *sql.DB
	Models models.Models
}

func (app *Application) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (app *Application) ParseDateRange(r *http.Request) (time.Time, time.Time) {
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	
	now := time.Now()
	// Default to today if not provided
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	if startStr != "" && endStr != "" {
		if s, err := time.Parse("2006-01-02", startStr); err == nil {
			start = s
		}
		if e, err := time.Parse("2006-01-02", endStr); err == nil {
			end = time.Date(e.Year(), e.Month(), e.Day(), 23, 59, 59, 0, e.Location())
		}
	}
	return start, end
}

func (app *Application) StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (app *Application) Float64Ptr(f float64) *float64 {
	return &f
}

// RenderPage is a generic helper to render a page with the base layout
func (app *Application) RenderPage(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC in RenderPage (%s): %v", templateName, err)
			http.Error(w, fmt.Sprintf("PANIC: %v", err), http.StatusInternalServerError)
		}
	}()

	user := middleware.GetUser(r.Context())
	userName := "User"
	role := ""
	if user != nil {
		userName = user.Name
		role = user.Role
	}

	files := []string{
		filepath.Join("ui", "html", "layouts", "base.tmpl"),
		filepath.Join("ui", "html", "pages", templateName+".tmpl"),
	}

	ts := template.New(filepath.Base(files[0])).Funcs(template.FuncMap{
		"subtract": func(a, b float64) float64 {
			return a - b
		},
		"add": func(a, b float64) float64 {
			return a + b
		},
	})
	
	ts, err := ts.ParseFiles(files...)
	if err != nil {
		log.Printf("Error parsing templates for %s: %v", templateName, err)
		http.Error(w, "PARSE ERROR: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch locations and settings if user is logged in
	var locations []*models.BusinessLocation
	activeLocationID := middleware.GetLocationID(r.Context())
	currencySymbol := "TSh"
	installedModules := make(map[string]bool)
	
	if user != nil {
		tenantID := middleware.GetTenantID(r.Context())
		locations, _ = app.Models.GetLocationsByTenant(tenantID)
		
		if settings, err := app.Models.GetBusinessSettings(tenantID); err == nil {
			currencySymbol = settings.CurrencySymbol
		}

		if modules, err := app.Models.GetTenantModules(tenantID); err == nil {
			for _, m := range modules {
				installedModules[m.Key] = m.IsInstalled
			}
		}
	}

	// Merge global data with page-specific data
	fullData := struct {
		Role             string
		UserName         string
		PageData         interface{}
		Locations        []*models.BusinessLocation
		ActiveLocationID int
		CurrencySymbol   string
		Modules          map[string]bool
	}{
		Role:             role,
		UserName:         userName,
		PageData:         data,
		Locations:        locations,
		ActiveLocationID: activeLocationID,
		CurrencySymbol:   currencySymbol,
		Modules:          installedModules,
	}

	err = ts.ExecuteTemplate(w, "base", fullData)
	if err != nil {
		log.Printf("TEMPLATE ERROR (%s): %v", templateName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
