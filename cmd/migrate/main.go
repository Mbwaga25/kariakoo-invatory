package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables from system")
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "invatory"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&multiStatements=true", dbUser, dbPass, dbHost, dbPort)
	log.Printf("Connecting to database at %s:%s (without DB select)...", dbHost, dbPort)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create database if not exists
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(fmt.Sprintf("USE %s", dbName))
	if err != nil {
		log.Fatal(err)
	}

	// Clean tables before seeding ONLY if DB_CLEAN is explicitly set to true
	if os.Getenv("DB_CLEAN") == "true" {
		log.Println("DB_CLEAN is set to true. Dropping tables for a clean seed...")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, _ = db.Exec("DROP TABLE IF EXISTS users")
		_, _ = db.Exec("DROP TABLE IF EXISTS products")
		_, _ = db.Exec("DROP TABLE IF EXISTS business_locations")
		_, _ = db.Exec("DROP TABLE IF EXISTS tenants")
		_, _ = db.Exec("DROP TABLE IF EXISTS categories")
		_, _ = db.Exec("DROP TABLE IF EXISTS brands")
		_, _ = db.Exec("DROP TABLE IF EXISTS units")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	} else {
		log.Println("DB_CLEAN is not true. Skipping table drops to prevent data loss.")
	}

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

	// Run seeders from cmd/migrate/seed/
	seedDir := filepath.Join("cmd", "migrate", "seed")
	if _, err := os.Stat(seedDir); err == nil {
		seedFiles, _ := os.ReadDir(seedDir)
		for _, f := range seedFiles {
			if filepath.Ext(f.Name()) == ".sql" {
				log.Printf("Executing seeder: %s", f.Name())
				content, err := os.ReadFile(filepath.Join(seedDir, f.Name()))
				if err == nil {
					_, err = db.Exec(string(content))
					if err != nil {
						log.Printf("Error executing seeder %s: %v", f.Name(), err)
					}
				}
			}
		}
	}

	log.Println("All migrations and seeding completed successfully!")
}
