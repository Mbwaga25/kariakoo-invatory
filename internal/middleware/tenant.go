package middleware

import (
	"context"
	"net/http"
	"strconv"
	"kariakoo/inventory/internal/models"
)

type contextKey string

const tenantContextKey = contextKey("tenantID")
const locationContextKey = contextKey("locationID")
const roleContextKey = contextKey("userRole")
const userContextKey = contextKey("user")

// TenantContext injects the tenant context into the request.
func TenantContext(models *models.Models) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("user_id")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := models.GetUserByID(cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			if user.TenantID != nil {
				ctx = context.WithValue(ctx, tenantContextKey, *user.TenantID)
			}

			// Determine Active Location: Cookie > User DB Default
			activeLocationID := 0
			if locCookie, err := r.Cookie("location_id"); err == nil {
				activeLocationID, _ = strconv.Atoi(locCookie.Value)
			}
			
			if activeLocationID == 0 && user.LocationID != nil {
				activeLocationID = *user.LocationID
			}

			// Existing users created before location assignment support should still land on a real location.
			if activeLocationID == 0 && user.TenantID != nil {
				if locations, err := models.GetLocationsByTenant(*user.TenantID); err == nil && len(locations) > 0 {
					activeLocationID = locations[0].ID
				}
			}

			if activeLocationID != 0 {
				ctx = context.WithValue(ctx, locationContextKey, activeLocationID)
			}
			
			ctx = context.WithValue(ctx, roleContextKey, user.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuthentication redirects unauthenticated users to the login page
func RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if cookie.Value == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetTenantID is a helper to retrieve the tenant ID from context
func GetTenantID(ctx context.Context) int {
	val, ok := ctx.Value(tenantContextKey).(int)
	if !ok {
		return 0
	}
	return val
}

// GetLocationID is a helper to retrieve the location ID from context
func GetLocationID(ctx context.Context) int {
	val, ok := ctx.Value(locationContextKey).(int)
	if !ok {
		return 0
	}
	return val
}

// GetRole is a helper to retrieve the user role from context
func GetRole(ctx context.Context) string {
	val, ok := ctx.Value(roleContextKey).(string)
	if !ok {
		return ""
	}
	return val
}
// GetUser is a helper to retrieve the full user object from context
func GetUser(ctx context.Context) *models.User {
	val, ok := ctx.Value(userContextKey).(*models.User)
	if !ok {
		return nil
	}
	return val
}
