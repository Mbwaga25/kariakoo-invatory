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
	StockAlerts         []*Product
	RecentSales         []*Sale
	RecentTransfers     []*StockTransfer
	PendingOrders       []*Sale
	PendingInvoices     []*Sale
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
	query := `SELECT p.name, p.sku, COALESCE(SUM(pl.qty_available), 0) as current_stock
			  FROM products p
			  LEFT JOIN product_locations pl ON p.id = pl.product_id
			  WHERE p.tenant_id = ?
			  GROUP BY p.id
			  ORDER BY p.product_type, p.name`
	
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*StockReport
	for rows.Next() {
		var sr StockReport
		err := rows.Scan(&sr.ProductName, &sr.SKU, &sr.CurrentStock)
		if err != nil {
			return nil, err
		}
		sr.StockValue = 0
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

func (m *Models) GetDashboardData(tenantID int, locationID *int, start, end time.Time) (*DashboardData, error) {
	data := &DashboardData{}

	// Basic filters for Sales
	salesWhere := "WHERE s.tenant_id = ? AND s.transaction_date BETWEEN ? AND ?"
	args := []interface{}{tenantID, start, end}
	if locationID != nil {
		salesWhere += " AND s.business_location_id = ?"
		args = append(args, *locationID)
	}

	// Basic filters for Purchases
	purchasesWhere := "WHERE p.tenant_id = ? AND p.purchase_date BETWEEN ? AND ?"
	pArgs := []interface{}{tenantID, start, end}
	if locationID != nil {
		purchasesWhere += " AND p.business_location_id = ?"
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
		return nil, fmt.Errorf("TotalPurchases: %v", err)
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

	// 7. Stock Alerts (All time, usually not filtered by date)
	err = m.DB.QueryRow(`
		SELECT COUNT(*) FROM product_locations pl
		JOIN products p ON pl.product_id = p.id
		WHERE p.tenant_id = ? AND pl.qty_available < 10
	`, tenantID).Scan(&data.StockAlertsCount)
	if err != nil {
		return nil, err
	}

	// 7b. Stock Alerts Details
	rows, err := m.DB.Query(`
		SELECT p.id, p.name, p.sku, pl.qty_available 
		FROM product_locations pl
		JOIN products p ON pl.product_id = p.id
		WHERE p.tenant_id = ? AND pl.qty_available < 10
		LIMIT 5
	`, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p Product
			rows.Scan(&p.ID, &p.Name, &p.SKU, &p.LocationQty)
			data.StockAlerts = append(data.StockAlerts, &p)
		}
	}

	// 8. Monthly Sales (Last 30 days for the chart, independent of filter)
	rows, err = m.DB.Query(`
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

	// 9. Recent Sales
	rows, err = m.DB.Query(`
		SELECT id, invoice_no, final_total, transaction_date, status
		FROM sales WHERE tenant_id = ? ORDER BY transaction_date DESC LIMIT 5
	`, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s Sale
			rows.Scan(&s.ID, &s.InvoiceNo, &s.FinalTotal, &s.TransactionDate, &s.Status)
			data.RecentSales = append(data.RecentSales, &s)
		}
	}

	// 10. Recent Transfers
	rows, err = m.DB.Query(`
		SELECT id, ref_no, final_total, transaction_date, status
		FROM stock_transfers WHERE tenant_id = ? ORDER BY transaction_date DESC LIMIT 5
	`, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t StockTransfer
			rows.Scan(&t.ID, &t.RefNo, &t.FinalTotal, &t.TransactionDate, &t.Status)
			data.RecentTransfers = append(data.RecentTransfers, &t)
		}
	}

	// 11. Pending Orders (status not final)
	rows, err = m.DB.Query(`
		SELECT id, invoice_no, final_total, transaction_date, status
		FROM sales WHERE tenant_id = ? AND status != 'final' ORDER BY transaction_date DESC LIMIT 5
	`, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s Sale
			rows.Scan(&s.ID, &s.InvoiceNo, &s.FinalTotal, &s.TransactionDate, &s.Status)
			data.PendingOrders = append(data.PendingOrders, &s)
		}
	}

	// 12. Pending Invoices (payment_status not paid)
	rows, err = m.DB.Query(`
		SELECT id, invoice_no, final_total, transaction_date, payment_status
		FROM sales WHERE tenant_id = ? AND payment_status != 'paid' ORDER BY transaction_date DESC LIMIT 5
	`, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s Sale
			rows.Scan(&s.ID, &s.InvoiceNo, &s.FinalTotal, &s.TransactionDate, &s.PaymentStatus)
			data.PendingInvoices = append(data.PendingInvoices, &s)
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
			  LEFT JOIN expense_categories ec ON e.expense_category_id = ec.id
			  WHERE e.tenant_id = ? AND e.transaction_date BETWEEN ? AND ?
			  GROUP BY e.expense_category_id`
	
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

