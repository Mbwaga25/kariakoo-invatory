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

	rows, err := db.Query("DESCRIBE orders")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Orders Table Columns:")
	for rows.Next() {
		var field, typ, null, key, def, extra sql.NullString
		rows.Scan(&field, &typ, &null, &key, &def, &extra)
		fmt.Printf("Field: %s, Type: %s, Null: %s\n", field.String, typ.String, null.String)
	}
}
