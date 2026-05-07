package models

import (
	"time"
)

type Purchase struct {
	ID                 int
	TenantID           int
	BusinessLocationID int
	SupplierID         *int
	RefNo              string
	PurchaseDate       time.Time
	Status             string
	PaymentStatus      string
	TotalBeforeTax     float64
	TaxAmount          float64
	DiscountAmount     float64
	FinalTotal         float64
	CreatedBy          *int
	CreatedAt          time.Time
	Items              []*PurchaseItem
}

type PurchaseItem struct {
	ID            int
	PurchaseID    int
	ProductID     int
	Quantity      float64
	PurchasePrice float64
	LineTotal     float64
	// Extended info for UI
	ProductName string
	SKU         string
}

func (m *Models) GetPurchasesByTenant(tenantID int, locationID int) ([]*Purchase, error) {
	query := `SELECT id, tenant_id, business_location_id, supplier_id, ref_no, purchase_date, status, payment_status, total_before_tax, tax_amount, discount_amount, final_total, created_by, created_at 
			  FROM purchases WHERE tenant_id = ?`
	
	args := []interface{}{tenantID}
	if locationID != 0 {
		query += " AND business_location_id = ?"
		args = append(args, locationID)
	}
	query += " ORDER BY purchase_date DESC"
	
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var purchases []*Purchase
	for rows.Next() {
		var p Purchase
		err := rows.Scan(&p.ID, &p.TenantID, &p.BusinessLocationID, &p.SupplierID, &p.RefNo, &p.PurchaseDate, &p.Status, &p.PaymentStatus, &p.TotalBeforeTax, &p.TaxAmount, &p.DiscountAmount, &p.FinalTotal, &p.CreatedBy, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		purchases = append(purchases, &p)
	}

	return purchases, nil
}

func (m *Models) InsertPurchase(p *Purchase) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `INSERT INTO purchases (tenant_id, business_location_id, supplier_id, ref_no, purchase_date, status, payment_status, total_before_tax, tax_amount, discount_amount, final_total, created_by)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	res, err := tx.Exec(query, p.TenantID, p.BusinessLocationID, p.SupplierID, p.RefNo, p.PurchaseDate, p.Status, p.PaymentStatus, p.TotalBeforeTax, p.TaxAmount, p.DiscountAmount, p.FinalTotal, p.CreatedBy)
	if err != nil {
		return 0, err
	}

	purchaseID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert Items and update stock
	for _, item := range p.Items {
		itemQuery := "INSERT INTO purchase_items (purchase_id, product_id, quantity, purchase_price, line_total) VALUES (?, ?, ?, ?, ?)"
		_, err = tx.Exec(itemQuery, purchaseID, item.ProductID, item.Quantity, item.PurchasePrice, item.LineTotal)
		if err != nil {
			return 0, err
		}

		// Update stock in product_locations
		stockQuery := `INSERT INTO product_locations (product_id, location_id, qty_available) 
					   VALUES (?, ?, ?) 
					   ON DUPLICATE KEY UPDATE qty_available = qty_available + ?`
		_, err = tx.Exec(stockQuery, item.ProductID, p.BusinessLocationID, item.Quantity, item.Quantity)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	return purchaseID, err
}
