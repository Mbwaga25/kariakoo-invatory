package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/invatory?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Generate real hash for '123456'
	hashBytes, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	actualHash := string(hashBytes)

	users := []struct {
		name     string
		email    string
		role     string
		tenantID interface{}
	}{
		{"Super Admin", "superadmin@test.com", "SuperAdmin", nil},
		{"Tenant Admin", "shopadmin@test.com", "ShopAdmin", 1},
		{"Shop Keeper", "shopkeeper@test.com", "ShopKeeper", 1},
		{"Store Keeper", "storekeeper@test.com", "StoreKeeper", 1},
	}

	// Ensure tenant 1 exists
	_, _ = db.Exec("INSERT IGNORE INTO tenants (id, name) VALUES (1, 'Main Shop')")
	// Ensure a default location exists for tenant 1
	_, _ = db.Exec("INSERT IGNORE INTO business_locations (id, tenant_id, name, location_id, city, country) VALUES (1, 1, 'Kariakoo Main', 'BL001', 'Dar es Salaam', 'Tanzania')")
	
	for _, u := range users {
		// Use INSERT ON DUPLICATE KEY UPDATE to avoid FK issues
		query := `INSERT INTO users (name, email, password_hash, role, tenant_id) VALUES (?, ?, ?, ?, ?)
				  ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), role = VALUES(role), name = VALUES(name)`
		_, err = db.Exec(query, u.name, u.email, actualHash, u.role, u.tenantID)
		if err != nil {
			log.Printf("Error seeding user %s: %v", u.email, err)
		} else {
			log.Printf("Successfully seeded/updated user: %s", u.email)
		}
	}

	log.Println("Seeding complete!")
}

