package models

import (
	"time"
)

type Sale struct {
	ID                 int
	TenantID           int
	BusinessLocationID *int
	CustomerID         *int
	InvoiceNo          string
	TransactionDate    time.Time
	DueDate            *time.Time
	Status             string
	PaymentStatus      string
	TotalBeforeTax     float64
	TaxAmount          float64
	DiscountAmount     float64
	FinalTotal         float64
	CreatedBy          *int
	CreatedAt          time.Time
	Items              []*SaleItem
	Payments           []*TransactionPayment
}

type SaleItem struct {
	ID        int
	SaleID    int
	ProductID int
	Quantity  float64
	UnitPrice float64
	LineTotal float64
	// Extended info for UI
	ProductName string
	SKU         string
}

type TransactionPayment struct {
	ID            int
	TenantID      int
	SaleID        int
	Amount        float64
	Method        string
	TransactionNo string
	Note          string
	PaidOn        time.Time
	CreatedBy     int
}

type CashRegister struct {
	ID                 int
	TenantID           int
	BusinessLocationID int
	UserID             int
	Status             string
	ClosedAt           *time.Time
	ClosingAmount      float64
	TotalCardSlips     int
	TotalCheques       int
	ClosingNote        string
	CreatedAt          time.Time
}

func (m *Models) GetSalesByTenant(tenantID int, locationID int) ([]*Sale, error) {
	query := `SELECT id, tenant_id, business_location_id, customer_id, invoice_no, transaction_date, due_date, status, payment_status, total_before_tax, tax_amount, discount_amount, final_total, created_by, created_at 
			  FROM sales WHERE tenant_id = ?`
	
	args := []interface{}{tenantID}
	if locationID != 0 {
		query += " AND business_location_id = ?"
		args = append(args, locationID)
	}
	query += " ORDER BY transaction_date DESC"
	
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []*Sale
	for rows.Next() {
		var s Sale
		err := rows.Scan(&s.ID, &s.TenantID, &s.BusinessLocationID, &s.CustomerID, &s.InvoiceNo, &s.TransactionDate, &s.DueDate, &s.Status, &s.PaymentStatus, &s.TotalBeforeTax, &s.TaxAmount, &s.DiscountAmount, &s.FinalTotal, &s.CreatedBy, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		sales = append(sales, &s)
	}

	return sales, nil
}

func (m *Models) InsertSale(s *Sale) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `INSERT INTO sales (tenant_id, business_location_id, customer_id, invoice_no, transaction_date, due_date, status, payment_status, total_before_tax, tax_amount, discount_amount, final_total, created_by)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	res, err := tx.Exec(query, s.TenantID, s.BusinessLocationID, s.CustomerID, s.InvoiceNo, s.TransactionDate, s.DueDate, s.Status, s.PaymentStatus, s.TotalBeforeTax, s.TaxAmount, s.DiscountAmount, s.FinalTotal, s.CreatedBy)
	if err != nil {
		return 0, err
	}

	saleID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert Items
	for _, item := range s.Items {
		itemQuery := "INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, line_total) VALUES (?, ?, ?, ?, ?)"
		_, err = tx.Exec(itemQuery, saleID, item.ProductID, item.Quantity, item.UnitPrice, item.LineTotal)
		if err != nil {
			return 0, err
		}
	}

	// Insert Payments
	for _, p := range s.Payments {
		pQuery := "INSERT INTO transaction_payments (tenant_id, sale_id, amount, method, transaction_no, note, paid_on, created_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
		_, err = tx.Exec(pQuery, s.TenantID, saleID, p.Amount, p.Method, p.TransactionNo, p.Note, time.Now(), s.CreatedBy)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	return saleID, err
}

func (m *Models) GetOpenRegister(userID int, locationID int) (*CashRegister, error) {
	query := "SELECT id, tenant_id, business_location_id, user_id, status, created_at FROM cash_registers WHERE user_id = ? AND business_location_id = ? AND status = 'open' LIMIT 1"
	
	var r CashRegister
	err := m.DB.QueryRow(query, userID, locationID).Scan(&r.ID, &r.TenantID, &r.BusinessLocationID, &r.UserID, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (m *Models) OpenRegister(r *CashRegister) error {
	query := "INSERT INTO cash_registers (tenant_id, business_location_id, user_id, status, created_at) VALUES (?, ?, ?, 'open', ?)"
	_, err := m.DB.Exec(query, r.TenantID, r.BusinessLocationID, r.UserID, time.Now())
	return err
}
