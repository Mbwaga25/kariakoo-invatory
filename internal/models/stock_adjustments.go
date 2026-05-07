package models

import (
	"time"
)

type StockAdjustment struct {
	ID                  int
	TenantID            int
	BusinessLocationID  int
	RefNo               string
	TransactionDate     time.Time
	AdjustmentType      string
	FinalTotal          float64
	TotalAmountRecovered float64
	AdditionalNotes     string
	CreatedBy           *int
	CreatedAt           time.Time
	Items               []*StockAdjustmentItem
	LocationName        string
}

type StockAdjustmentItem struct {
	ID                int
	StockAdjustmentID int
	ProductID         int
	Quantity          float64
	UnitPrice         float64
	ProductName       string
	SKU               string
}

func (m *Models) GetStockAdjustmentsByTenant(tenantID int, locationID int) ([]*StockAdjustment, error) {
	query := `SELECT sa.id, sa.tenant_id, sa.business_location_id, sa.ref_no, sa.transaction_date, sa.adjustment_type, sa.final_total, l.name as location_name
			  FROM stock_adjustments sa
			  JOIN business_locations l ON sa.business_location_id = l.id
			  WHERE sa.tenant_id = ?`
	
	args := []interface{}{tenantID}
	if locationID != 0 {
		query += " AND sa.business_location_id = ?"
		args = append(args, locationID)
	}
	query += " ORDER BY sa.transaction_date DESC"
	
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adjustments []*StockAdjustment
	for rows.Next() {
		var sa StockAdjustment
		err := rows.Scan(&sa.ID, &sa.TenantID, &sa.BusinessLocationID, &sa.RefNo, &sa.TransactionDate, &sa.AdjustmentType, &sa.FinalTotal, &sa.LocationName)
		if err != nil {
			return nil, err
		}
		adjustments = append(adjustments, &sa)
	}

	return adjustments, nil
}

func (m *Models) InsertStockAdjustment(sa *StockAdjustment) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `INSERT INTO stock_adjustments (tenant_id, business_location_id, ref_no, transaction_date, adjustment_type, final_total, created_by)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	res, err := tx.Exec(query, sa.TenantID, sa.BusinessLocationID, sa.RefNo, sa.TransactionDate, sa.AdjustmentType, sa.FinalTotal, sa.CreatedBy)
	if err != nil {
		return 0, err
	}

	adjustmentID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range sa.Items {
		itemQuery := "INSERT INTO stock_adjustment_items (stock_adjustment_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)"
		_, err = tx.Exec(itemQuery, adjustmentID, item.ProductID, item.Quantity, item.UnitPrice)
		if err != nil {
			return 0, err
		}

		// Update stock (subtract because adjustment is usually for loss/damage)
		// If user wants to add stock, they can use a purchase or I could add logic for positive adjustment
		_, err = tx.Exec("UPDATE product_locations SET qty_available = qty_available - ? WHERE product_id = ? AND location_id = ?", item.Quantity, item.ProductID, sa.BusinessLocationID)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	return adjustmentID, err
}
