package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/invatory?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, email, role, location_id FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Users with Locations:")
	for rows.Next() {
		var id int
		var email, role string
		var locID sql.NullInt64
		rows.Scan(&id, &email, &role, &locID)
		fmt.Printf("ID: %d, Email: %s, Role: %s, LocationID: %v\n", id, email, role, locID)
	}
}
