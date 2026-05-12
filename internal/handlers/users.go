package handlers

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) UserList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())
	
	var users []*models.User
	var err error

	if user.Role == "SuperAdmin" {
		users, err = app.Models.GetAllUsers()
	} else {
		users, err = app.Models.GetUsersByTenant(tenantID)
	}
	if err != nil {
		log.Printf("ERROR UserList: %v", err)
	}

	locations, _ := app.Models.GetLocationsByTenant(tenantID)

	app.RenderPage(w, r, "admin/users", struct {
		Users     []*models.User
		UserRole  string
		Locations []*models.BusinessLocation
	}{
		Users:     users,
		UserRole:  user.Role,
		Locations: locations,
	})
}

func (app *Application) UserStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	// 1. Get current user to check permissions (only SuperAdmin or ShopAdmin)
	currentUser := middleware.GetUser(r.Context())
	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Parse form
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role")
	name := r.FormValue("name")

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// 3. Logic for Tenant Admin (Creating a new Tenant)
	var tenantID int
	if role == "ShopAdmin" && currentUser.Role == "SuperAdmin" {
		// Creating a new tenant
		tID, err := app.Models.InsertTenant(name + "'s Shop")
		if err != nil {
			log.Println("Error creating tenant:", err)
			http.Error(w, "Error creating tenant", http.StatusInternalServerError)
			return
		}
		tenantID = int(tID)

		// Create DEFAULT LOCATION for this new tenant
		loc := &models.BusinessLocation{
			TenantID:   tenantID,
			Name:       "Main Branch",
			LocationID: "BL001",
			City:       "Default City",
			Country:    "Default Country",
		}
		_, err = app.Models.InsertLocation(loc)
		if err != nil {
			log.Println("Error creating default location:", err)
			// Non-blocking but should be logged
		}
	} else {
		// Adding a user to the current tenant
		tenantID = middleware.GetTenantID(r.Context())
	}

	// 4. Insert User
	query := "INSERT INTO users (tenant_id, name, email, password_hash, role) VALUES (?, ?, ?, ?, ?)"
	_, err := app.DB.Exec(query, tenantID, name, email, string(hash), role)
	if err != nil {
		log.Println("Error creating user:", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}
