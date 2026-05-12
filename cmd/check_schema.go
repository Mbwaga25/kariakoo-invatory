package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/invatory?parseTime=true&multiStatements=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Drop old orders-related tables and recreate
	log.Println("Dropping old orders tables...")
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	db.Exec("DROP TABLE IF EXISTS order_payments")
	db.Exec("DROP TABLE IF EXISTS order_items")
	db.Exec("DROP TABLE IF EXISTS orders")
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	// Re-run orders migration
	log.Println("Recreating orders tables...")
	content, err := os.ReadFile(filepath.Join("cmd", "migrate", "sql", "031_orders.sql"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(string(content))
	if err != nil {
		log.Fatal("Error creating orders tables:", err)
	}

	// Ensure store_management module is enabled
	db.Exec("INSERT IGNORE INTO tenant_modules (tenant_id, module_key, is_installed) VALUES (1, 'store_management', 1)")

	log.Println("Orders tables recreated successfully!")
}
