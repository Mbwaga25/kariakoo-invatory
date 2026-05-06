package models

import (
	"database/sql"
	"time"
)

type Order struct {
	ID           int
	TenantID     int
	OrderType    string // 'StoreOrder' or 'BulkOrder'
	OrderFrom    string // Neighboring store name
	CustomerName string
	SourceShop   string
	Status       string
	CreatedAt    time.Time
}

// Models wrapper
type Models struct {
	DB *sql.DB
}

func NewModels(db *sql.DB) Models {
	return Models{
		DB: db,
	}
}
