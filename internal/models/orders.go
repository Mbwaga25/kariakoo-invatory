package models

import (
	"fmt"
	"time"
)

type StoreOrder struct {
	ID              int
	TenantID        int
	OrderType       string // 'StoreOrder' or 'BulkOrder'
	RefNo           string
	PlacedBy        int
	OrderFrom       string
	FromShopID      *int
	FromStoreID     *int
	ToLocationID    int
	Status          string // pending, accepted, rejected, completed
	PaymentStatus   string // paid, unpaid, incomplete
	TotalAmount     float64
	AmountPaid      float64
	RemainingAmount float64
	ProcessedBy     *int
	ProcessedAt     *time.Time
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Joined fields
	Items            []*OrderItem
	PlacedByName     string
	ProcessedByName  string
	ToLocationName   string
	FromLocationName string
}

type OrderItem struct {
	ID           int
	OrderID      int
	ProductID    int
	Quantity     float64
	FromShopQty  float64 // Just recorded, no stock deduction
	FromStoreQty float64 // Deducted from store on accept
	UnitPrice    float64
	Subtotal     float64
	ProductName  string
	ProductSKU   string
	CategoryName string
	BrandName    string
	ProductType  string
}

type OrderPayment struct {
	ID            int
	OrderID       int
	Amount        float64
	PaymentMethod string
	Notes         string
	PaidBy        *int
	PaidByName    string
	CreatedAt     time.Time
}

type OrderSummary struct {
	TotalOrders     int
	PendingOrders   int
	AcceptedOrders  int
	CompletedOrders int
	RejectedOrders  int
	PaidCount       int
	UnpaidCount     int
	IncompleteCount int
	TotalRevenue    float64
}

// GetOrdersByTenant returns all orders for a tenant, optionally filtered by role and location
func (m *Models) GetOrdersByTenant(tenantID int, status string, orderType string, role string, locationID int, userID int) ([]*StoreOrder, error) {
	query := `SELECT o.id, o.tenant_id, o.order_type, o.ref_no, o.placed_by, 
			  COALESCE(o.order_from, ''), o.to_location_id, o.status, o.payment_status,
			  o.total_amount, o.amount_paid, o.remaining_amount, 
			  COALESCE(o.notes, ''), o.created_at,
			  u.name as placed_by_name,
			  COALESCE(bl_to.name, '') as to_location_name,
			  COALESCE(bl_from.name, '') as from_location_name
			  FROM orders o
			  JOIN users u ON o.placed_by = u.id
			  LEFT JOIN business_locations bl_to ON o.to_location_id = bl_to.id
			  LEFT JOIN business_locations bl_from ON o.from_store_id = bl_from.id
			  WHERE o.tenant_id = ?`
	
	args := []interface{}{tenantID}

	if role == "ShopKeeper" {
		query += " AND o.placed_by = ?"
		args = append(args, userID)
	} else if role == "StoreKeeper" {
		query += " AND o.from_store_id = ?"
		args = append(args, locationID)
	}

	if status != "" {
		query += " AND o.status = ?"
		args = append(args, status)
	}
	if orderType != "" {
		query += " AND o.order_type = ?"
		args = append(args, orderType)
	}

	query += " ORDER BY o.created_at DESC"
	
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("GetOrdersByTenant: %v", err)
	}
	defer rows.Close()

	var orders []*StoreOrder
	for rows.Next() {
		var o StoreOrder
		err := rows.Scan(&o.ID, &o.TenantID, &o.OrderType, &o.RefNo, &o.PlacedBy,
			&o.OrderFrom, &o.ToLocationID, &o.Status, &o.PaymentStatus,
			&o.TotalAmount, &o.AmountPaid, &o.RemainingAmount,
			&o.Notes, &o.CreatedAt,
			&o.PlacedByName, &o.ToLocationName, &o.FromLocationName)
		if err != nil {
			return nil, fmt.Errorf("scan order: %v", err)
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

// GetOrderByID returns a single order with items
func (m *Models) GetOrderByID(id int, tenantID int) (*StoreOrder, error) {
	query := `SELECT o.id, o.tenant_id, o.order_type, o.ref_no, o.placed_by,
			  COALESCE(o.order_from, ''), o.to_location_id, o.status, o.payment_status,
			  o.total_amount, o.amount_paid, o.remaining_amount,
			  o.processed_by, o.processed_at,
			  COALESCE(o.notes, ''), o.created_at,
			  u.name as placed_by_name,
			  COALESCE(pu.name, '') as processed_by_name,
			  COALESCE(bl_to.name, '') as to_location_name,
			  COALESCE(bl_from.name, '') as from_location_name
			  FROM orders o
			  JOIN users u ON o.placed_by = u.id
			  LEFT JOIN users pu ON o.processed_by = pu.id
			  LEFT JOIN business_locations bl_to ON o.to_location_id = bl_to.id
			  LEFT JOIN business_locations bl_from ON o.from_store_id = bl_from.id
			  WHERE o.id = ? AND o.tenant_id = ?`

	var o StoreOrder
	err := m.DB.QueryRow(query, id, tenantID).Scan(
		&o.ID, &o.TenantID, &o.OrderType, &o.RefNo, &o.PlacedBy,
		&o.OrderFrom, &o.ToLocationID, &o.Status, &o.PaymentStatus,
		&o.TotalAmount, &o.AmountPaid, &o.RemainingAmount,
		&o.ProcessedBy, &o.ProcessedAt,
		&o.Notes, &o.CreatedAt,
		&o.PlacedByName, &o.ProcessedByName, &o.ToLocationName, &o.FromLocationName)
	if err != nil {
		return nil, err
	}

	// Get items
	itemRows, err := m.DB.Query(`
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, 
		       COALESCE(oi.from_shop_qty, 0), COALESCE(oi.from_store_qty, 0),
		       oi.unit_price, oi.subtotal,
		       p.name, p.sku, COALESCE(c.name, ''), COALESCE(b.name, ''), COALESCE(p.product_type, 'Protector')
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE oi.order_id = ?`, id)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var item OrderItem
			itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, 
				&item.FromShopQty, &item.FromStoreQty,
				&item.UnitPrice, &item.Subtotal, &item.ProductName, &item.ProductSKU,
				&item.CategoryName, &item.BrandName, &item.ProductType)
			o.Items = append(o.Items, &item)
		}
	}

	return &o, nil
}

// InsertOrder creates a new order with items
func (m *Models) InsertOrder(o *StoreOrder) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `INSERT INTO orders (tenant_id, order_type, ref_no, placed_by, order_from, from_shop_id, from_store_id, to_location_id, status, payment_status, total_amount, amount_paid, remaining_amount, notes)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := tx.Exec(query, o.TenantID, o.OrderType, o.RefNo, o.PlacedBy,
		o.OrderFrom, o.FromShopID, o.FromStoreID, o.ToLocationID,
		o.Status, o.PaymentStatus, o.TotalAmount, o.AmountPaid, o.RemainingAmount, o.Notes)
	if err != nil {
		return 0, fmt.Errorf("insert order: %v", err)
	}

	orderID, _ := res.LastInsertId()

	for _, item := range o.Items {
		_, err = tx.Exec(`INSERT INTO order_items (order_id, product_id, quantity, from_shop_qty, from_store_qty, unit_price, subtotal) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			orderID, item.ProductID, item.Quantity, item.FromShopQty, item.FromStoreQty, item.UnitPrice, item.Subtotal)
		if err != nil {
			return 0, fmt.Errorf("insert order item: %v", err)
		}
	}

	return orderID, tx.Commit()
}

// AcceptOrder accepts an order and reduces stock from the processing location
func (m *Models) AcceptOrder(orderID int, tenantID int, processedBy int, locationID int) error {
	// Ping to ensure connection is alive before starting transaction
	if err := m.DB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	if locationID == 0 {
		return fmt.Errorf("no active location is assigned to this user")
	}

	var sourceLocationID int
	err := m.DB.QueryRow(`SELECT COALESCE(from_store_id, 0) FROM orders WHERE id = ? AND tenant_id = ? AND status = 'pending'`,
		orderID, tenantID).Scan(&sourceLocationID)
	if err != nil {
		return fmt.Errorf("order not found or not in pending status: %v", err)
	}
	if sourceLocationID == 0 {
		return fmt.Errorf("this order does not have a source store assigned")
	}
	if sourceLocationID != locationID {
		return fmt.Errorf("this order must be processed from the assigned source store")
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Update order status
	res, err := tx.Exec(`UPDATE orders SET status = 'accepted', processed_by = ?, processed_at = NOW() WHERE id = ? AND tenant_id = ? AND status = 'pending'`,
		processedBy, orderID, tenantID)
	if err != nil {
		return fmt.Errorf("update order status SQL: %v", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("order not found or not in pending status")
	}

	// Get order items - only deduct from_store_qty (from shop items are just numbers)
	rows, err := tx.Query(`SELECT product_id, COALESCE(from_store_qty, quantity) FROM order_items WHERE order_id = ?`, orderID)
	if err != nil {
		return fmt.Errorf("get order items: %v", err)
	}
	defer rows.Close()

	type item struct {
		pid int
		qty float64
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.pid, &it.qty); err != nil {
			return err
		}
		if it.qty > 0 {
			items = append(items, it)
		}
	}
	rows.Close() // Close early to free up connection for Execs

	for _, it := range items {
		// Reduce stock from the processing store location
		// Note: We use COALESCE and a check to ensure we don't fail if row missing (though it shouldn't be for a store)
		// Or we can check if it exists first.
		_, err = tx.Exec(`UPDATE product_locations SET qty_available = qty_available - ? WHERE product_id = ? AND location_id = ?`,
			it.qty, it.pid, locationID)
		if err != nil {
			return fmt.Errorf("reduce stock for product %d: %v", it.pid, err)
		}
	}

	return tx.Commit()
}

// RejectOrder rejects an order and logs the reason
func (m *Models) RejectOrder(orderID int, tenantID int, processedBy int, reason string) error {
	_, err := m.DB.Exec(`UPDATE orders SET status = 'rejected', processed_by = ?, processed_at = NOW(), 
						 notes = CONCAT(COALESCE(notes, ''), '\n[Rejected]: ', ?) 
						 WHERE id = ? AND tenant_id = ? AND status = 'pending'`,
		processedBy, reason, orderID, tenantID)
	return err
}

// CompleteOrder marks the order as completed
func (m *Models) CompleteOrder(orderID int, tenantID int) error {
	_, err := m.DB.Exec(`UPDATE orders SET status = 'completed' WHERE id = ? AND tenant_id = ?`, orderID, tenantID)
	return err
}

// UpdateOrderPayment updates the payment status and amounts
func (m *Models) UpdateOrderPayment(orderID int, tenantID int, amountPaid float64, paidBy int) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert payment record
	_, err = tx.Exec(`INSERT INTO order_payments (order_id, amount, paid_by) VALUES (?, ?, ?)`,
		orderID, amountPaid, paidBy)
	if err != nil {
		return err
	}

	// Update order totals
	// Note: MySQL evaluates expressions left to right. amount_paid is updated first, 
	// so subsequent uses of amount_paid refer to the NEW value.
	_, err = tx.Exec(`UPDATE orders SET 
		amount_paid = amount_paid + ?,
		remaining_amount = total_amount - amount_paid,
		payment_status = CASE 
			WHEN amount_paid >= total_amount THEN 'paid'
			WHEN amount_paid > 0 THEN 'incomplete'
			ELSE 'unpaid'
		END
		WHERE id = ? AND tenant_id = ?`,
		amountPaid, orderID, tenantID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetOrderPayments returns all payments for an order
func (m *Models) GetOrderPayments(orderID int) ([]*OrderPayment, error) {
	rows, err := m.DB.Query(`
		SELECT op.id, op.order_id, op.amount, COALESCE(op.payment_method, 'cash'), 
		       COALESCE(op.notes, ''), op.created_at, COALESCE(u.name, 'System')
		FROM order_payments op
		LEFT JOIN users u ON op.paid_by = u.id
		WHERE op.order_id = ?
		ORDER BY op.created_at DESC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*OrderPayment
	for rows.Next() {
		var p OrderPayment
		rows.Scan(&p.ID, &p.OrderID, &p.Amount, &p.PaymentMethod, &p.Notes, &p.CreatedAt, &p.PaidByName)
		payments = append(payments, &p)
	}
	return payments, nil
}

// GetOrderSummary returns aggregated order stats for the dashboard
func (m *Models) GetOrderSummary(tenantID int, role string, locationID int, userID int) (*OrderSummary, error) {
	summary := &OrderSummary{}

	whereClause := "tenant_id = ?"
	args := []interface{}{tenantID}

	if role == "ShopKeeper" {
		whereClause += " AND placed_by = ?"
		args = append(args, userID)
	} else if role == "StoreKeeper" {
		whereClause += " AND from_store_id = ?"
		args = append(args, locationID)
	}

	buildQuery := func(status, payment string) (string, []interface{}) {
		q := fmt.Sprintf("SELECT COUNT(*) FROM orders WHERE %s", whereClause)
		a := append([]interface{}{}, args...)
		if status != "" {
			q += " AND status = ?"
			a = append(a, status)
		}
		if payment != "" {
			q += " AND payment_status = ?"
			a = append(a, payment)
		}
		return q, a
	}

	q, a := buildQuery("", "")
	m.DB.QueryRow(q, a...).Scan(&summary.TotalOrders)
	
	q, a = buildQuery("pending", "")
	m.DB.QueryRow(q, a...).Scan(&summary.PendingOrders)
	
	q, a = buildQuery("accepted", "")
	m.DB.QueryRow(q, a...).Scan(&summary.AcceptedOrders)
	
	q, a = buildQuery("completed", "")
	m.DB.QueryRow(q, a...).Scan(&summary.CompletedOrders)
	
	q, a = buildQuery("rejected", "")
	m.DB.QueryRow(q, a...).Scan(&summary.RejectedOrders)
	
	q, a = buildQuery("", "paid")
	m.DB.QueryRow(q, a...).Scan(&summary.PaidCount)
	
	q, a = buildQuery("", "unpaid")
	m.DB.QueryRow(q, a...).Scan(&summary.UnpaidCount)
	
	q, a = buildQuery("", "incomplete")
	m.DB.QueryRow(q, a...).Scan(&summary.IncompleteCount)

	// Revenue
	revQuery := fmt.Sprintf("SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE %s AND status != 'rejected'", whereClause)
	m.DB.QueryRow(revQuery, args...).Scan(&summary.TotalRevenue)

	return summary, nil
}

// GetPendingOrdersForStoreKeeper returns pending orders for the store keeper to process
func (m *Models) GetPendingOrdersForStoreKeeper(tenantID int, locationID int) ([]*StoreOrder, error) {
	query := `SELECT o.id, o.tenant_id, o.order_type, o.ref_no, o.placed_by,
			  COALESCE(o.order_from, ''), o.to_location_id, o.status, o.payment_status,
			  o.total_amount, o.amount_paid, o.remaining_amount,
			  COALESCE(o.notes, ''), o.created_at,
			  u.name as placed_by_name,
			  COALESCE(bl.name, '') as to_location_name
			  FROM orders o
			  JOIN users u ON o.placed_by = u.id
			  LEFT JOIN business_locations bl ON o.to_location_id = bl.id
			  WHERE o.tenant_id = ? AND o.status = 'pending' AND o.from_store_id = ?
			  ORDER BY o.created_at DESC`

	rows, err := m.DB.Query(query, tenantID, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*StoreOrder
	for rows.Next() {
		var o StoreOrder
		err := rows.Scan(&o.ID, &o.TenantID, &o.OrderType, &o.RefNo, &o.PlacedBy,
			&o.OrderFrom, &o.ToLocationID, &o.Status, &o.PaymentStatus,
			&o.TotalAmount, &o.AmountPaid, &o.RemainingAmount,
			&o.Notes, &o.CreatedAt,
			&o.PlacedByName, &o.ToLocationName)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

// GetUsersByTenant returns all users for a specific tenant
func (m *Models) GetUsersByTenant(tenantID int) ([]*User, error) {
	query := `SELECT id, tenant_id, location_id, name, email, role, created_at FROM users WHERE tenant_id = ? ORDER BY created_at DESC`
	
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.TenantID, &u.LocationID, &u.Name, &u.Email, &u.Role, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

// GetAllUsers returns all users (for SuperAdmin)
func (m *Models) GetAllUsers() ([]*User, error) {
	query := `SELECT id, tenant_id, location_id, name, email, role, created_at FROM users ORDER BY created_at DESC`
	
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.TenantID, &u.LocationID, &u.Name, &u.Email, &u.Role, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

// GenerateOrderRefNo creates a unique reference number
func (m *Models) GenerateOrderRefNo(tenantID int) string {
	var count int
	m.DB.QueryRow("SELECT COUNT(*) + 1 FROM orders WHERE tenant_id = ?", tenantID).Scan(&count)
	return fmt.Sprintf("ORD-%04d", count)
}

// GetBestSellingProducts returns top selling products by order quantity
func (m *Models) GetBestSellingProducts(tenantID int, limit int) ([]*Product, error) {
	query := `SELECT p.id, p.name, p.sku, COALESCE(SUM(oi.quantity), 0) as total_qty
			  FROM order_items oi
			  JOIN orders o ON oi.order_id = o.id
			  JOIN products p ON oi.product_id = p.id
			  WHERE o.tenant_id = ? AND o.status != 'rejected'
			  GROUP BY p.id
			  ORDER BY total_qty DESC
			  LIMIT ?`

	rows, err := m.DB.Query(query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var p Product
		rows.Scan(&p.ID, &p.Name, &p.SKU, &p.LocationQty)
		products = append(products, &p)
	}
	return products, nil
}
