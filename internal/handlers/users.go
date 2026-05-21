package handlers

import (
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) UserList(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

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

	var locations []*models.BusinessLocation
	if user.Role == "SuperAdmin" {
		locations, _ = app.Models.GetAllLocations()
	} else {
		locations, _ = app.Models.GetLocationsByTenant(tenantID)
	}

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

	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
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

	var tenantID int
	locationID := middleware.GetLocationID(r.Context())

	// 3. Logic for Tenant Admin (Creating a new Tenant)
	if role == "ShopAdmin" {
		if currentUser.Role != "SuperAdmin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

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
			LocationType: "shop",
		}
		locID, err := app.Models.InsertLocation(loc)
		if err != nil {
			log.Println("Error creating default location:", err)
			// Non-blocking but should be logged
		} else {
			locationID = int(locID)
		}
	} else {
		if currentUser.Role == "SuperAdmin" {
			locationID, _ = strconv.Atoi(r.FormValue("location_id"))
			if locationID == 0 {
				http.Error(w, "Please select a location for this user", http.StatusBadRequest)
				return
			}

			locations, err := app.Models.GetAllLocations()
			if err != nil {
				http.Error(w, "Unable to load locations", http.StatusInternalServerError)
				return
			}

			found := false
			for _, loc := range locations {
				if loc.ID == locationID {
					tenantID = loc.TenantID
					found = true
					break
				}
			}

			if !found {
				http.Error(w, "Selected location is invalid", http.StatusBadRequest)
				return
			}
		} else {
			// Adding a user to the current tenant
			tenantID = middleware.GetTenantID(r.Context())
			if selectedLocationID, _ := strconv.Atoi(r.FormValue("location_id")); selectedLocationID > 0 {
				locationID = selectedLocationID
			}
		}
	}

	if role == "StoreKeeper" || role == "ShopKeeper" {
		if locationID == 0 {
			http.Error(w, "Please select a location for this user", http.StatusBadRequest)
			return
		}

		if tenantID == 0 {
			http.Error(w, "Please select a tenant for this user", http.StatusBadRequest)
			return
		}

		locs, err := app.Models.GetLocationsByTenant(tenantID)
		if err != nil {
			http.Error(w, "Unable to verify location", http.StatusInternalServerError)
			return
		}

		validLocation := false
		for _, loc := range locs {
			if loc.ID == locationID {
				validLocation = true
				break
			}
		}

		if !validLocation {
			http.Error(w, "Selected location does not belong to the tenant", http.StatusBadRequest)
			return
		}
	}

	if tenantID == 0 {
		http.Error(w, "Tenant could not be determined for this user", http.StatusBadRequest)
		return
	}

	// 4. Insert User
	query := "INSERT INTO users (tenant_id, location_id, name, email, password_hash, role) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := app.DB.Exec(query, tenantID, locationID, name, email, string(hash), role)
	if err != nil {
		log.Println("Error creating user:", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}
