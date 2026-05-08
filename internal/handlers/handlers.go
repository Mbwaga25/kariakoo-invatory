package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

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

	// Fetch locations for switcher if user is admin
	var locations []*models.BusinessLocation
	activeLocationID := middleware.GetLocationID(r.Context())
	if user != nil {
		tenantID := middleware.GetTenantID(r.Context())
		locations, _ = app.Models.GetLocationsByTenant(tenantID)
	}

	// Merge global data with page-specific data
	fullData := struct {
		Role             string
		UserName         string
		PageData         interface{}
		Locations        []*models.BusinessLocation
		ActiveLocationID int
	}{
		Role:             role,
		UserName:         userName,
		PageData:         data,
		Locations:        locations,
		ActiveLocationID: activeLocationID,
	}

	err = ts.ExecuteTemplate(w, "base", fullData)
	if err != nil {
		log.Printf("TEMPLATE ERROR (%s): %v", templateName, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
