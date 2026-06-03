package models

import (
	"database/sql"
)

type SellingPriceGroup struct {
	ID          int
	TenantID    int
	Name        string
	Description string
}

type ProductGroupPrice struct {
	ID                  int
	TenantID            int
	ProductID           int
	SellingPriceGroupID int
	Price               float64
}

func (m *Models) GetSellingPriceGroupsByTenant(tenantID int) ([]*SellingPriceGroup, error) {
	query := "SELECT id, tenant_id, name, COALESCE(description, '') FROM selling_price_groups WHERE tenant_id = ? ORDER BY name ASC"
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*SellingPriceGroup
	for rows.Next() {
		var g SellingPriceGroup
		err := rows.Scan(&g.ID, &g.TenantID, &g.Name, &g.Description)
		if err != nil {
			return nil, err
		}
		groups = append(groups, &g)
	}
	return groups, nil
}

func (m *Models) GetSellingPriceGroupByID(id int, tenantID int) (*SellingPriceGroup, error) {
	query := "SELECT id, tenant_id, name, COALESCE(description, '') FROM selling_price_groups WHERE id = ? AND tenant_id = ?"
	var g SellingPriceGroup
	err := m.DB.QueryRow(query, id, tenantID).Scan(&g.ID, &g.TenantID, &g.Name, &g.Description)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (m *Models) InsertSellingPriceGroup(g *SellingPriceGroup) (int64, error) {
	query := "INSERT INTO selling_price_groups (tenant_id, name, description) VALUES (?, ?, ?)"
	res, err := m.DB.Exec(query, g.TenantID, g.Name, g.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *Models) UpdateSellingPriceGroup(g *SellingPriceGroup) error {
	query := "UPDATE selling_price_groups SET name = ?, description = ? WHERE id = ? AND tenant_id = ?"
	_, err := m.DB.Exec(query, g.Name, g.Description, g.ID, g.TenantID)
	return err
}

func (m *Models) DeleteSellingPriceGroup(id int, tenantID int) error {
	query := "DELETE FROM selling_price_groups WHERE id = ? AND tenant_id = ?"
	_, err := m.DB.Exec(query, id, tenantID)
	return err
}

// GetProductGroupPrices returns a map of group_id -> price for a given product
func (m *Models) GetProductGroupPrices(tenantID int, productID int) (map[int]float64, error) {
	query := "SELECT selling_price_group_id, price FROM product_group_prices WHERE tenant_id = ? AND product_id = ?"
	rows, err := m.DB.Query(query, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make(map[int]float64)
	for rows.Next() {
		var groupID int
		var price float64
		if err := rows.Scan(&groupID, &price); err != nil {
			return nil, err
		}
		prices[groupID] = price
	}
	return prices, nil
}

// SetProductGroupPrice updates or inserts a price for a product in a selling price group
func (m *Models) SetProductGroupPrice(tenantID int, productID int, groupID int, price float64) error {
	// Check if already exists
	var id int
	err := m.DB.QueryRow("SELECT id FROM product_group_prices WHERE tenant_id = ? AND product_id = ? AND selling_price_group_id = ?", tenantID, productID, groupID).Scan(&id)
	if err == sql.ErrNoRows {
		// Insert
		query := "INSERT INTO product_group_prices (tenant_id, product_id, selling_price_group_id, price) VALUES (?, ?, ?, ?)"
		_, err = m.DB.Exec(query, tenantID, productID, groupID, price)
		return err
	} else if err != nil {
		return err
	}

	// Update
	query := "UPDATE product_group_prices SET price = ? WHERE id = ?"
	_, err = m.DB.Exec(query, price, id)
	return err
}
