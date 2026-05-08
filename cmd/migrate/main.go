package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/?parseTime=true&multiStatements=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create database if not exists
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS invatory")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("USE invatory")
	if err != nil {
		log.Fatal(err)
	}

	// Clean tables before seeding
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	_, _ = db.Exec("DROP TABLE IF EXISTS users")
	_, _ = db.Exec("DROP TABLE IF EXISTS products")
	_, _ = db.Exec("DROP TABLE IF EXISTS business_locations")
	_, _ = db.Exec("DROP TABLE IF EXISTS tenants")
	_, _ = db.Exec("DROP TABLE IF EXISTS categories")
	_, _ = db.Exec("DROP TABLE IF EXISTS brands")
	_, _ = db.Exec("DROP TABLE IF EXISTS units")
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	// Read and execute SQL files from cmd/migrate/sql/
	sqlDir := filepath.Join("cmd", "migrate", "sql")
	files, err := os.ReadDir(sqlDir)
	if err != nil {
		log.Fatal("Error reading SQL directory: ", err)
	}

	// Sort files to ensure 001 runs before 002
	var filenames []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			filenames = append(filenames, f.Name())
		}
	}
	sort.Strings(filenames)

	for _, f := range filenames {
		log.Printf("Executing migration: %s", f)
		content, err := os.ReadFile(filepath.Join(sqlDir, f))
		if err != nil {
			log.Fatalf("Error reading file %s: %v", f, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Error executing migration %s: %v", f, err)
		}
	}

	// Seeding
	password := "123456"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Seed Tenant 1
	_, _ = db.Exec("INSERT IGNORE INTO tenants (id, name) VALUES (1, 'Main Business')")
	
	// Seed Location 1 for Tenant 1
	_, _ = db.Exec("INSERT IGNORE INTO business_locations (id, tenant_id, name) VALUES (1, 1, 'Headquarters')")
	_, _ = db.Exec("INSERT IGNORE INTO business_locations (id, tenant_id, name) VALUES (2, 1, 'Branch Office')")

	// Seed Users
	_, err = db.Exec("INSERT INTO users (id, tenant_id, location_id, name, email, password_hash, role) VALUES (1, NULL, NULL, 'Super Admin', 'superadmin@test.com', ?, 'SuperAdmin')", string(hash))
	if err != nil { log.Printf("Error seeding superadmin: %v", err) }
	_, err = db.Exec("INSERT INTO users (id, tenant_id, location_id, name, email, password_hash, role) VALUES (2, 1, 1, 'Shop Admin', 'shopadmin@test.com', ?, 'ShopAdmin')", string(hash))
	if err != nil { log.Printf("Error seeding shopadmin: %v", err) }
	_, err = db.Exec("INSERT INTO users (id, tenant_id, location_id, name, email, password_hash, role) VALUES (3, 1, 2, 'Shop Keeper', 'shopkeeper@test.com', ?, 'ShopKeeper')", string(hash))
	if err != nil { log.Printf("Error seeding shopkeeper: %v", err) }

	log.Println("All migrations and seeding completed successfully!")
}
