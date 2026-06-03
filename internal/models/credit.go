package models

import (
	"database/sql"
)

// GetCustomerBalance calculates the outstanding balance for a customer (opening balance + unpaid bulk orders)
func (m *Models) GetCustomerBalance(tenantID int, customerName string) (float64, error) {
	var opening float64
	err := m.DB.QueryRow("SELECT COALESCE(opening_balance, 0) FROM contacts WHERE tenant_id = ? AND name = ? LIMIT 1", tenantID, customerName).Scan(&opening)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	var unpaidOrders float64
	query := `SELECT COALESCE(SUM(total_amount - amount_paid), 0) FROM orders WHERE tenant_id = ? AND order_from = ? AND order_type = 'BulkOrder' AND status != 'rejected'`
	err = m.DB.QueryRow(query, tenantID, customerName).Scan(&unpaidOrders)
	if err != nil {
		return opening, nil // Default to opening balance if error or no orders
	}

	return opening + unpaidOrders, nil
}
