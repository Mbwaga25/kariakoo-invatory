package models

import (
	"time"
)

type Tenant struct {
	ID          int
	Name        string
	PackageType string
	IsActive    bool
	CreatedAt   time.Time
}

type User struct {
	ID           int
	TenantID     *int // Nullable for SuperAdmin
	LocationID   *int // Nullable if admin or not assigned yet
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}


// GetUserByEmail returns a user by their email address
func (m *Models) GetUserByEmail(email string) (*User, error) {
	query := "SELECT id, tenant_id, location_id, name, email, password_hash, role, created_at FROM users WHERE email = ?"
	var u User
	err := m.DB.QueryRow(query, email).Scan(&u.ID, &u.TenantID, &u.LocationID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByID returns a user by their ID
func (m *Models) GetUserByID(id string) (*User, error) {
	query := "SELECT id, tenant_id, location_id, name, email, password_hash, role, created_at FROM users WHERE id = ?"
	var u User
	err := m.DB.QueryRow(query, id).Scan(&u.ID, &u.TenantID, &u.LocationID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
