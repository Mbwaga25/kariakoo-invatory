package models

import (
	"time"
)

type ProfitLossReport struct {
	TotalSales          float64
	TotalPurchases      float64
	TotalExpenses       float64
	GrossProfit         float64
	NetProfit           float64
}

type StockReport struct {
	ProductName      string
	SKU              string
	CurrentStock     float64
	StockValue       float64 // Current Stock * Unit Purchase Price
}

type RegisterReport struct {
	UserName     string
	LocationName string
	Status       string
	OpenedAt     time.Time
	ClosedAt     *time.Time
	ClosingAmount float64
}

func (m *Models) GetProfitLossReport(tenantID int, start, end time.Time) (*ProfitLossReport, error) {
	var report ProfitLossReport

	// Get Total Sales
	err := m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM sales WHERE tenant_id = ? AND transaction_date BETWEEN ? AND ?", tenantID, start, end).Scan(&report.TotalSales)
	if err != nil {
		return nil, err
	}

	// Get Total Purchases
	err = m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM purchases WHERE tenant_id = ? AND purchase_date BETWEEN ? AND ?", tenantID, start, end).Scan(&report.TotalPurchases)
	if err != nil {
		return nil, err
	}

	// Get Total Expenses
	err = m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM expenses WHERE tenant_id = ? AND transaction_date BETWEEN ? AND ?", tenantID, start, end).Scan(&report.TotalExpenses)
	if err != nil {
		return nil, err
	}

	report.GrossProfit = report.TotalSales - report.TotalPurchases
	report.NetProfit = report.GrossProfit - report.TotalExpenses

	return &report, nil
}

func (m *Models) GetStockReport(tenantID int) ([]*StockReport, error) {
	query := `SELECT p.name, p.sku, COALESCE(SUM(pl.qty_available), 0) as current_stock, p.purchase_price
			  FROM products p
			  LEFT JOIN product_locations pl ON p.id = pl.product_id
			  WHERE p.tenant_id = ?
			  GROUP BY p.id`
	
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*StockReport
	for rows.Next() {
		var sr StockReport
		var purchasePrice float64
		err := rows.Scan(&sr.ProductName, &sr.SKU, &sr.CurrentStock, &purchasePrice)
		if err != nil {
			return nil, err
		}
		sr.StockValue = sr.CurrentStock * purchasePrice
		reports = append(reports, &sr)
	}

	return reports, nil
}

func (m *Models) GetRegisterReport(tenantID int) ([]*RegisterReport, error) {
	query := `SELECT u.name, l.name, cr.status, cr.created_at, cr.closed_at, cr.closing_amount
			  FROM cash_registers cr
			  JOIN users u ON cr.user_id = u.id
			  JOIN business_locations l ON cr.business_location_id = l.id
			  WHERE cr.tenant_id = ?
			  ORDER BY cr.created_at DESC`
	
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*RegisterReport
	for rows.Next() {
		var rr RegisterReport
		err := rows.Scan(&rr.UserName, &rr.LocationName, &rr.Status, &rr.OpenedAt, &rr.ClosedAt, &rr.ClosingAmount)
		if err != nil {
			return nil, err
		}
		reports = append(reports, &rr)
	}
	return reports, nil
}

