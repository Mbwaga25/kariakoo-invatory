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
	preset := r.URL.Query().Get("preset")
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	
	now := time.Now()
	
	if preset == "" && startStr != "" && endStr != "" {
		preset = "custom"
	}
	
	if preset == "" {
		preset = "today"
	}

	var start, end time.Time

	switch preset {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	case "yesterday":
		y := now.AddDate(0, 0, -1)
		start = time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, y.Location())
		end = time.Date(y.Year(), y.Month(), y.Day(), 23, 59, 59, 0, y.Location())
	case "weekly":
		weekday := int(now.Weekday())
		daysToMonday := weekday - 1
		if weekday == 0 {
			daysToMonday = 6
		}
		monday := now.AddDate(0, 0, -daysToMonday)
		start = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
		sunday := monday.AddDate(0, 0, 6)
		end = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())
	case "monthly":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
		lastDay := nextMonth.AddDate(0, 0, -1)
		end = time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), 23, 59, 59, 0, lastDay.Location())
	case "yearly":
		start = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
		end = time.Date(now.Year(), time.December, 31, 23, 59, 59, 0, now.Location())
	case "custom":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		
		if startStr != "" {
			if s, err := time.Parse("2006-01-02", startStr); err == nil {
				start = time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, s.Location())
			}
		}
		if endStr != "" {
			if e, err := time.Parse("2006-01-02", endStr); err == nil {
				end = time.Date(e.Year(), e.Month(), e.Day(), 23, 59, 59, 0, e.Location())
			}
		}
	default:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
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
		"derefInt": func(i *int) int {
			if i == nil {
				return 0
			}
			return *i
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
