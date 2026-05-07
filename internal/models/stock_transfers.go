package models

import (
	"time"
)

type StockTransfer struct {
	ID                int
	TenantID          int
	FromLocationID    int
	ToLocationID      int
	RefNo             string
	TransactionDate   time.Time
	Status            string
	ShippingCharges   float64
	FinalTotal        float64
	AdditionalNotes   string
	CreatedBy         int
	CreatedAt         time.Time
	Items             []*StockTransferItem
	FromLocationName  string
	ToLocationName    string
}

type StockTransferItem struct {
	ID              int
	StockTransferID int
	ProductID       int
	Quantity        float64
	UnitPurchasePrice float64
	ProductName     string
	SKU             string
}

func (m *Models) GetStockTransfersByTenant(tenantID int) ([]*StockTransfer, error) {
	query := `SELECT st.id, st.tenant_id, st.from_location_id, st.to_location_id, st.ref_no, st.transaction_date, st.status, st.final_total, fl.name as from_location_name, tl.name as to_location_name
			  FROM stock_transfers st
			  JOIN business_locations fl ON st.from_location_id = fl.id
			  JOIN business_locations tl ON st.to_location_id = tl.id
			  WHERE st.tenant_id = ?
			  ORDER BY st.transaction_date DESC`
	
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []*StockTransfer
	for rows.Next() {
		var st StockTransfer
		err := rows.Scan(&st.ID, &st.TenantID, &st.FromLocationID, &st.ToLocationID, &st.RefNo, &st.TransactionDate, &st.Status, &st.FinalTotal, &st.FromLocationName, &st.ToLocationName)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, &st)
	}

	return transfers, nil
}

func (m *Models) InsertStockTransfer(st *StockTransfer) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `INSERT INTO stock_transfers (tenant_id, from_location_id, to_location_id, ref_no, transaction_date, status, final_total, created_by)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	res, err := tx.Exec(query, st.TenantID, st.FromLocationID, st.ToLocationID, st.RefNo, st.TransactionDate, st.Status, st.FinalTotal, st.CreatedBy)
	if err != nil {
		return 0, err
	}

	transferID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range st.Items {
		itemQuery := "INSERT INTO stock_transfer_items (stock_transfer_id, product_id, quantity) VALUES (?, ?, ?)"
		_, err = tx.Exec(itemQuery, transferID, item.ProductID, item.Quantity)
		if err != nil {
			return 0, err
		}

		// Deduct from FROM location
		_, err = tx.Exec("UPDATE product_locations SET qty_available = qty_available - ? WHERE product_id = ? AND location_id = ?", item.Quantity, item.ProductID, st.FromLocationID)
		if err != nil {
			return 0, err
		}

		// Add to TO location
		_, err = tx.Exec(`INSERT INTO product_locations (product_id, location_id, qty_available) 
						  VALUES (?, ?, ?) 
						  ON DUPLICATE KEY UPDATE qty_available = qty_available + ?`, 
						  item.ProductID, st.ToLocationID, item.Quantity, item.Quantity)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	return transferID, err
}
