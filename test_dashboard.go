package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"kariakoo/inventory/internal/models"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	m := models.NewModels(db)

	locationID := 1
	start := time.Now().AddDate(0, 0, -30)
	end := time.Now()

	fmt.Println("Running GetDashboardData for tenant 1, location 1...")
	data, err := m.GetDashboardData(1, &locationID, start, end)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("Success! Sales: %f, Alerts Count: %d\n", data.TotalSales, data.StockAlertsCount)
}
