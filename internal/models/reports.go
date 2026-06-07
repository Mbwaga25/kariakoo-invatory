package models

import (
	"database/sql"
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
	RecentOrders        []*StoreOrder
	PendingOrders       []*Sale
	PendingInvoices     []*Sale
	DailyOrderFlow      []DailyOrderFlow
}

type DailyOrderFlow struct {
	Date  string
	Count int
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

type StockMovement struct {
	Date         time.Time
	ProductName  string
	LocationName string
	Type         string // Purchase, Sale, StoreOrder, BulkOrder
	RefNo        string
	Quantity     float64
	Flow         string // IN / OUT
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
	var rows *sql.Rows

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

	// 7. Stock Alerts (Filtered by location if provided)
	var alertQuery string
	var alertArgs []interface{}
	if locationID != nil && *locationID > 0 {
		alertQuery = `
			SELECT COUNT(*) FROM (
				SELECT p.id,
					   COALESCE(pl.qty_available, 0) as total_qty,
					   COALESCE(p.alert_quantity, 50) as alert_qty
				FROM products p
				LEFT JOIN product_locations pl ON p.id = pl.product_id AND pl.location_id = ?
				WHERE p.tenant_id = ?
			) t WHERE t.total_qty < t.alert_qty`
		alertArgs = []interface{}{*locationID, tenantID}
	} else {
		alertQuery = `
			SELECT COUNT(*) FROM (
				SELECT p.id,
					   COALESCE(SUM(pl.qty_available), 0) as total_qty,
					   COALESCE(p.alert_quantity, 50) as alert_qty
				FROM products p
				LEFT JOIN product_locations pl ON p.id = pl.product_id
				WHERE p.tenant_id = ?
				GROUP BY p.id, p.alert_quantity
			) t WHERE t.total_qty < t.alert_qty`
		alertArgs = []interface{}{tenantID}
	}

	err = m.DB.QueryRow(alertQuery, alertArgs...).Scan(&data.StockAlertsCount)
	if err != nil {
		return nil, err
	}


	// 7b. Stock Alerts Details
	var detailsQuery string
	var detailsArgs []interface{}
	if locationID != nil && *locationID > 0 {
		detailsQuery = `
			SELECT t.id, t.name, t.sku, t.total_qty, t.alert_qty
			FROM (
				SELECT p.id, p.name, p.sku,
					   COALESCE(pl.qty_available, 0) as total_qty,
					   COALESCE(p.alert_quantity, 50) as alert_qty
				FROM products p
				LEFT JOIN product_locations pl ON p.id = pl.product_id AND pl.location_id = ?
				WHERE p.tenant_id = ?
			) t
			WHERE t.total_qty < t.alert_qty
			ORDER BY t.total_qty ASC
			LIMIT 5`
		detailsArgs = []interface{}{*locationID, tenantID}
	} else {
		detailsQuery = `
			SELECT t.id, t.name, t.sku, t.total_qty, t.alert_qty
			FROM (
				SELECT p.id, p.name, p.sku,
					   COALESCE(SUM(pl.qty_available), 0) as total_qty,
					   COALESCE(p.alert_quantity, 50) as alert_qty
				FROM products p
				LEFT JOIN product_locations pl ON p.id = pl.product_id
				WHERE p.tenant_id = ?
				GROUP BY p.id, p.name, p.sku, p.alert_quantity
			) t
			WHERE t.total_qty < t.alert_qty
			ORDER BY t.total_qty ASC
			LIMIT 5`
		detailsArgs = []interface{}{tenantID}
	}

	rows, err = m.DB.Query(detailsQuery, detailsArgs...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p Product
			var alertQty float64
			rows.Scan(&p.ID, &p.Name, &p.SKU, &p.LocationQty, &alertQty)
			p.AlertQuantity = &alertQty
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

	// 10b. Recent Store/Bulk Orders
	ordersQuery := `
		SELECT o.id, o.ref_no, o.order_type, o.total_amount, o.status, o.created_at, 
		       u.name as placed_by_name, 
		       COALESCE(bl_from.name, '') as from_location_name,
		       COALESCE(bl_to.name, '') as to_location_name
		FROM orders o
		JOIN users u ON o.placed_by = u.id
		LEFT JOIN business_locations bl_from ON o.from_store_id = bl_from.id
		LEFT JOIN business_locations bl_to ON o.to_location_id = bl_to.id
		WHERE o.tenant_id = ?
	`
	ordersArgs := []interface{}{tenantID}
	if locationID != nil && *locationID > 0 {
		ordersQuery += " AND (o.from_store_id = ? OR o.to_location_id = ?)"
		ordersArgs = append(ordersArgs, *locationID, *locationID)
	}
	ordersQuery += " ORDER BY o.created_at DESC LIMIT 10"

	rows, err = m.DB.Query(ordersQuery, ordersArgs...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var o StoreOrder
			rows.Scan(&o.ID, &o.RefNo, &o.OrderType, &o.TotalAmount, &o.Status, &o.CreatedAt, 
				&o.PlacedByName, &o.FromLocationName, &o.ToLocationName)
			data.RecentOrders = append(data.RecentOrders, &o)
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

	// 12. Daily Order Flow (Last 30 days)
	rows, err = m.DB.Query(`
		SELECT DATE(created_at) as d, COUNT(*) 
		FROM orders 
		WHERE tenant_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
		GROUP BY d ORDER BY d ASC
	`, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dof DailyOrderFlow
			rows.Scan(&dof.Date, &dof.Count)
			data.DailyOrderFlow = append(data.DailyOrderFlow, dof)
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

func (m *Models) GetLocationStockReport(tenantID int) (interface{}, []string, error) {
	locations, _ := m.GetLocationsByTenant(tenantID)
	var locNames []string
	locMap := make(map[int]string)
	for _, l := range locations {
		locNames = append(locNames, l.Name)
		locMap[l.ID] = l.Name
	}

	query := `SELECT p.id, p.name, bl.id as loc_id, COALESCE(pl.qty_available, 0)
			  FROM products p
			  CROSS JOIN business_locations bl
			  LEFT JOIN product_locations pl ON p.id = pl.product_id AND bl.id = pl.location_id
			  WHERE p.tenant_id = ? AND bl.tenant_id = ?
			  ORDER BY p.name, bl.name`
	
	rows, err := m.DB.Query(query, tenantID, tenantID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	type ProductStockRow struct {
		Product *Product
		Stock   map[string]float64
	}
	var report []ProductStockRow
	prodMap := make(map[int]int) // Maps pID to index in report slice

	for rows.Next() {
		var pID int
		var pName string
		var locID int
		var qty float64
		rows.Scan(&pID, &pName, &locID, &qty)
		
		idx, ok := prodMap[pID]
		if !ok {
			p := &Product{ID: pID, Name: pName}
			idx = len(report)
			prodMap[pID] = idx
			report = append(report, ProductStockRow{
				Product: p,
				Stock:   make(map[string]float64),
			})
		}
		report[idx].Stock[locMap[locID]] = qty
	}

	return report, locNames, nil
}



func (m *Models) GetStockHistory(tenantID int, start, end time.Time, productID int) ([]*StockMovement, error) {
	var movements []*StockMovement

	// Helper to add product filter
	prodFilter := ""
	if productID > 0 {
		prodFilter = fmt.Sprintf(" AND prod.id = %d ", productID)
	}

	// 1. Purchases (Stock IN)
	queryPurchases := fmt.Sprintf(`SELECT p.purchase_date, prod.name, bl.name, 'Purchase', p.ref_no, pi.quantity, 'IN'
					   FROM purchase_items pi
					   JOIN purchases p ON pi.purchase_id = p.id
					   JOIN products prod ON pi.product_id = prod.id
					   JOIN business_locations bl ON p.business_location_id = bl.id
					   WHERE p.tenant_id = ? %s AND p.purchase_date BETWEEN ? AND ?`, prodFilter)
	
	rows, err := m.DB.Query(queryPurchases, tenantID, start, end)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sm StockMovement
			rows.Scan(&sm.Date, &sm.ProductName, &sm.LocationName, &sm.Type, &sm.RefNo, &sm.Quantity, &sm.Flow)
			movements = append(movements, &sm)
		}
	}

	// 2. Sales (Stock OUT)
	querySales := fmt.Sprintf(`SELECT s.transaction_date, prod.name, bl.name, 'Sale', s.invoice_no, si.quantity, 'OUT'
				   FROM sale_items si
				   JOIN sales s ON si.sale_id = s.id
				   JOIN products prod ON si.product_id = prod.id
				   JOIN business_locations bl ON s.business_location_id = bl.id
				   WHERE s.tenant_id = ? %s AND s.status = 'final' AND s.transaction_date BETWEEN ? AND ?`, prodFilter)
	
	rows, err = m.DB.Query(querySales, tenantID, start, end)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sm StockMovement
			rows.Scan(&sm.Date, &sm.ProductName, &sm.LocationName, &sm.Type, &sm.RefNo, &sm.Quantity, &sm.Flow)
			movements = append(movements, &sm)
		}
	}

	// 3. Orders (Transfers)
	queryOrdersOut := fmt.Sprintf(`SELECT o.created_at, prod.name, bl.name, 'StoreOrder (Source)', o.ref_no, oi.quantity, 'OUT'
					   FROM order_items oi
					   JOIN orders o ON oi.order_id = o.id
					   JOIN products prod ON oi.product_id = prod.id
					   JOIN business_locations bl ON o.from_store_id = bl.id
					   WHERE o.tenant_id = ? %s AND o.status = 'accepted' AND o.created_at BETWEEN ? AND ?`, prodFilter)
	
	rows, err = m.DB.Query(queryOrdersOut, tenantID, start, end)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sm StockMovement
			rows.Scan(&sm.Date, &sm.ProductName, &sm.LocationName, &sm.Type, &sm.RefNo, &sm.Quantity, &sm.Flow)
			movements = append(movements, &sm)
		}
	}

	queryOrdersIn := fmt.Sprintf(`SELECT o.created_at, prod.name, bl.name, 'StoreOrder (Dest)', o.ref_no, oi.quantity, 'IN'
					  FROM order_items oi
					  JOIN orders o ON oi.order_id = o.id
					  JOIN products prod ON oi.product_id = prod.id
					  JOIN business_locations bl ON o.to_location_id = bl.id
					  WHERE o.tenant_id = ? %s AND o.status = 'accepted' AND o.created_at BETWEEN ? AND ?`, prodFilter)
	
	rows, err = m.DB.Query(queryOrdersIn, tenantID, start, end)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sm StockMovement
			rows.Scan(&sm.Date, &sm.ProductName, &sm.LocationName, &sm.Type, &sm.RefNo, &sm.Quantity, &sm.Flow)
			movements = append(movements, &sm)
		}
	}

	return movements, nil
}

type OrderReportData struct {
	Summary   OrderReportSummary
	Orders    []*StoreOrder
	Customers []*CustomerReportAnalysis
}

type OrderReportSummary struct {
	TotalOrders      int
	TotalAmount      float64
	TotalPaid        float64
	TotalRemaining   float64
	PaidFullCount    int
	PaidPartialCount int
	PendingCount     int
}

type CustomerReportAnalysis struct {
	CustomerName     string
	CreditLimit      float64
	InvoiceDueDays   int
	RangeTotal       float64
	RangePaid        float64
	RangeRemaining   float64
	TotalDebt        float64
	OldestUnpaidDays int
	LimitExceeded    bool
	Overdue          bool
}

func (m *Models) GetOrderReport(tenantID int, start, end time.Time, paymentStatus string) (*OrderReportData, error) {
	data := &OrderReportData{}

	// 1. Fetch Summary
	summaryQuery := `
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0), COALESCE(SUM(amount_paid), 0), COALESCE(SUM(remaining_amount), 0),
		       COUNT(CASE WHEN payment_status = 'paid' THEN 1 END),
		       COUNT(CASE WHEN payment_status = 'incomplete' THEN 1 END),
		       COUNT(CASE WHEN payment_status = 'unpaid' THEN 1 END)
		FROM orders
		WHERE tenant_id = ? AND created_at BETWEEN ? AND ?`
	summaryArgs := []interface{}{tenantID, start, end}
	if paymentStatus != "" {
		summaryQuery += " AND payment_status = ?"
		summaryArgs = append(summaryArgs, paymentStatus)
	}

	err := m.DB.QueryRow(summaryQuery, summaryArgs...).Scan(
		&data.Summary.TotalOrders, &data.Summary.TotalAmount, &data.Summary.TotalPaid, &data.Summary.TotalRemaining,
		&data.Summary.PaidFullCount, &data.Summary.PaidPartialCount, &data.Summary.PendingCount)
	if err != nil {
		return nil, err
	}

	// 2. Fetch Orders List
	ordersQuery := `
		SELECT o.id, o.tenant_id, o.order_type, o.ref_no, o.placed_by,
		       COALESCE(o.order_from, ''), o.to_location_id, o.status, o.payment_status,
		       o.total_amount, o.amount_paid, o.remaining_amount, o.created_at,
		       u.name as placed_by_name,
		       COALESCE(bl_to.name, '') as to_location_name,
		       COALESCE(bl_from.name, '') as from_location_name
		FROM orders o
		JOIN users u ON o.placed_by = u.id
		LEFT JOIN business_locations bl_to ON o.to_location_id = bl_to.id
		LEFT JOIN business_locations bl_from ON o.from_store_id = bl_from.id
		WHERE o.tenant_id = ? AND o.created_at BETWEEN ? AND ?`
	ordersArgs := []interface{}{tenantID, start, end}
	if paymentStatus != "" {
		ordersQuery += " AND o.payment_status = ?"
		ordersArgs = append(ordersArgs, paymentStatus)
	}
	ordersQuery += " ORDER BY o.created_at DESC"

	rows, err := m.DB.Query(ordersQuery, ordersArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o StoreOrder
		err := rows.Scan(&o.ID, &o.TenantID, &o.OrderType, &o.RefNo, &o.PlacedBy,
			&o.OrderFrom, &o.ToLocationID, &o.Status, &o.PaymentStatus,
			&o.TotalAmount, &o.AmountPaid, &o.RemainingAmount, &o.CreatedAt,
			&o.PlacedByName, &o.ToLocationName, &o.FromLocationName)
		if err != nil {
			return nil, err
		}
		data.Orders = append(data.Orders, &o)
	}

	// 3. Fetch Customer analysis
	custRows, err := m.DB.Query(`
		SELECT name, COALESCE(credit_limit, 0), COALESCE(invoice_due_days, 0)
		FROM contacts
		WHERE tenant_id = ? AND type = 'customer'
		ORDER BY name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer custRows.Close()

	for custRows.Next() {
		var ca CustomerReportAnalysis
		err := custRows.Scan(&ca.CustomerName, &ca.CreditLimit, &ca.InvoiceDueDays)
		if err != nil {
			return nil, err
		}

		// Calculate range stats for this specific customer name
		err = m.DB.QueryRow(`
			SELECT COALESCE(SUM(total_amount), 0), COALESCE(SUM(amount_paid), 0), COALESCE(SUM(remaining_amount), 0)
			FROM orders
			WHERE tenant_id = ? AND order_type = 'BulkOrder' AND order_from = ? AND created_at BETWEEN ? AND ?`,
			tenantID, ca.CustomerName, start, end).Scan(&ca.RangeTotal, &ca.RangePaid, &ca.RangeRemaining)
		if err != nil {
			return nil, err
		}

		// Calculate total outstanding debt (lifetime)
		err = m.DB.QueryRow(`
			SELECT COALESCE(SUM(remaining_amount), 0)
			FROM orders
			WHERE tenant_id = ? AND order_type = 'BulkOrder' AND order_from = ? AND payment_status != 'paid' AND status != 'rejected'`,
			tenantID, ca.CustomerName).Scan(&ca.TotalDebt)
		if err != nil {
			return nil, err
		}

		// Calculate oldest unpaid invoice days
		err = m.DB.QueryRow(`
			SELECT COALESCE(TIMESTAMPDIFF(DAY, MIN(created_at), NOW()), 0)
			FROM orders
			WHERE tenant_id = ? AND order_type = 'BulkOrder' AND order_from = ? AND payment_status != 'paid' AND status != 'rejected'`,
			tenantID, ca.CustomerName).Scan(&ca.OldestUnpaidDays)
		if err != nil {
			return nil, err
		}

		// Calculate warning flags
		if ca.CreditLimit > 0 && ca.TotalDebt > ca.CreditLimit {
			ca.LimitExceeded = true
		}
		if ca.InvoiceDueDays > 0 && ca.OldestUnpaidDays > ca.InvoiceDueDays {
			ca.Overdue = true
		}

		// Only include in analysis report if they placed an order in the range or have outstanding debt
		if ca.RangeTotal > 0 || ca.TotalDebt > 0 {
			data.Customers = append(data.Customers, &ca)
		}
	}

	return data, nil
}




