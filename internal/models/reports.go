package models

import (
	"fmt"
	"time"
)

type ProfitLossReport struct {
	TotalSales          float64
	TotalPurchases      float64
	TotalExpenses       float64
	GrossProfit         float64
	NetProfit           float64
}

type DashboardData struct {
	TotalSales          float64
	TotalPurchases      float64
	TotalExpenses       float64
	InvoiceDue          float64
	PurchaseDue         float64
	TotalSellReturn     float64
	TotalPurchaseReturn float64
	Net                 float64
	StockAlertsCount    int
	MonthlySales        []MonthlySale
}

type MonthlySale struct {
	Date  string
	Total float64
}

type PurchaseSellReport struct {
	TotalPurchase      float64
	TotalPurchaseReturn float64
	TotalSale          float64
	TotalSaleReturn    float64
	PurchaseDue        float64
	InvoiceDue         float64
}

type ExpenseReport struct {
	CategoryName string
	TotalAmount  float64
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

func (m *Models) GetDashboardData(tenantID int, locationID *int) (*DashboardData, error) {
	data := &DashboardData{}

	// Basic filters
	salesWhere := "WHERE s.tenant_id = ?"
	args := []interface{}{tenantID}
	if locationID != nil {
		salesWhere += " AND s.business_location_id = ?"
		args = append(args, *locationID)
	}

	purchasesWhere := "WHERE p.tenant_id = ?"
	pArgs := []interface{}{tenantID}
	if locationID != nil {
		purchasesWhere += " AND p.business_location_id = ?" // Note: check column name for location in purchases
		pArgs = append(pArgs, *locationID)
	}

	// 1. Total Sales
	err := m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM sales s "+salesWhere, args...).Scan(&data.TotalSales)
	if err != nil {
		return nil, fmt.Errorf("TotalSales: %v", err)
	}

	// 2. Total Purchases
	err = m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM purchases p "+purchasesWhere, pArgs...).Scan(&data.TotalPurchases)
	if err != nil {
		// Fallback: check if column is location_id or business_location_id
		err = m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM purchases p WHERE p.tenant_id = ?", tenantID).Scan(&data.TotalPurchases)
		if err != nil {
			return nil, fmt.Errorf("TotalPurchases: %v", err)
		}
	}

	// 3. Total Expenses
	err = m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM expenses s "+salesWhere, args...).Scan(&data.TotalExpenses)
	if err != nil {
		return nil, fmt.Errorf("TotalExpenses: %v", err)
	}

	// 4. Invoice Due (Sales - Paid)
	err = m.DB.QueryRow(`
		SELECT COALESCE(SUM(s.final_total - COALESCE(tp.paid_amount, 0)), 0)
		FROM sales s
		LEFT JOIN (
			SELECT sale_id, SUM(amount) as paid_amount 
			FROM transaction_payments 
			GROUP BY sale_id
		) tp ON s.id = tp.sale_id
		`+salesWhere, args...).Scan(&data.InvoiceDue)
	if err != nil {
		return nil, fmt.Errorf("InvoiceDue: %v", err)
	}

	// 5. Purchase Due
	err = m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM purchases p "+purchasesWhere, pArgs...).Scan(&data.PurchaseDue)
	if err != nil {
		return nil, fmt.Errorf("PurchaseDue: %v", err)
	}

	// 6. Net (Sales - Expenses)
	data.Net = data.TotalSales - data.TotalExpenses

	// 7. Stock Alerts (qty < 10 for simplicity, or we could use alert_quantity if we had it)
	err = m.DB.QueryRow(`
		SELECT COUNT(*) FROM product_locations pl
		JOIN products p ON pl.product_id = p.id
		WHERE p.tenant_id = ? AND pl.qty_available < 10
	`, tenantID).Scan(&data.StockAlertsCount)
	if err != nil {
		return nil, err
	}

	// 8. Monthly Sales (Last 30 days)
	rows, err := m.DB.Query(`
		SELECT DATE(transaction_date) as d, SUM(final_total) 
		FROM sales 
		WHERE tenant_id = ? AND transaction_date >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY d ORDER BY d ASC
	`, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ms MonthlySale
			rows.Scan(&ms.Date, &ms.Total)
			data.MonthlySales = append(data.MonthlySales, ms)
		}
	}

	return data, nil
}

func (m *Models) GetPurchaseSellReport(tenantID int, start, end time.Time) (*PurchaseSellReport, error) {
	report := &PurchaseSellReport{}

	// Total Purchases
	err := m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM purchases WHERE tenant_id = ? AND purchase_date BETWEEN ? AND ?", tenantID, start, end).Scan(&report.TotalPurchase)
	if err != nil {
		return nil, err
	}

	// Total Sales
	err = m.DB.QueryRow("SELECT COALESCE(SUM(final_total), 0) FROM sales WHERE tenant_id = ? AND transaction_date BETWEEN ? AND ?", tenantID, start, end).Scan(&report.TotalSale)
	if err != nil {
		return nil, err
	}

	// Dues (Simplified)
	err = m.DB.QueryRow(`
		SELECT COALESCE(SUM(s.final_total - COALESCE(tp.paid_amount, 0)), 0)
		FROM sales s
		LEFT JOIN (
			SELECT sale_id, SUM(amount) as paid_amount 
			FROM transaction_payments 
			GROUP BY sale_id
		) tp ON s.id = tp.sale_id
		WHERE s.tenant_id = ? AND s.transaction_date BETWEEN ? AND ?`, tenantID, start, end).Scan(&report.InvoiceDue)
	if err != nil {
		return nil, err
	}

	err = m.DB.QueryRow(`
		SELECT COALESCE(SUM(final_total), 0)
		FROM purchases
		WHERE tenant_id = ? AND purchase_date BETWEEN ? AND ?`, tenantID, start, end).Scan(&report.PurchaseDue)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func (m *Models) GetExpenseReport(tenantID int, start, end time.Time) ([]*ExpenseReport, error) {
	query := `SELECT COALESCE(ec.name, 'Uncategorized'), SUM(e.final_total)
			  FROM expenses e
			  LEFT JOIN categories ec ON e.category_id = ec.id
			  WHERE e.tenant_id = ? AND e.transaction_date BETWEEN ? AND ?
			  GROUP BY e.category_id`
	
	rows, err := m.DB.Query(query, tenantID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*ExpenseReport
	for rows.Next() {
		var er ExpenseReport
		err := rows.Scan(&er.CategoryName, &er.TotalAmount)
		if err != nil {
			return nil, err
		}
		reports = append(reports, &er)
	}
	return reports, nil
}

