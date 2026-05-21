package models

import (
	"database/sql"
	"time"
)

type Product struct {
	ID            int
	TenantID      int
	ProductType   string // 'Protector' or 'Cover'
	Name          string
	SKU           string
	AlertQuantity *float64
	UnitID        *int
	CategoryID    *int
	BrandID       *int
	Description   *string
	CreatedAt     time.Time
	
	// Location-specific details
	LocationQty   float64         `json:"location_qty"`
	LocationPrice sql.NullFloat64  `json:"location_price"`
	
	// Joined fields for display
	CategoryName string
	BrandName    string
}

type Category struct {
	ID          int
	TenantID    int
	ParentID    *int
	Name        string
	Description *string
	CreatedAt   time.Time
}

type Brand struct {
	ID          int
	TenantID    int
	Name        string
	Description *string
	CreatedAt   time.Time
}

type Unit struct {
	ID           int
	TenantID     int
	ActualName   string
	ShortName    string
	AllowDecimal bool
	CreatedAt    time.Time
}

type ProductLocation struct {
	ProductID    int
	LocationID   int
	QtyAvailable float64
	SellingPrice sql.NullFloat64
	IsActive     bool
}

func (m *Models) GetProductsByTenant(tenantID int, locationID int, categoryID int, brandID int) ([]*Product, error) {
	query := `SELECT p.id, p.tenant_id, COALESCE(p.product_type, 'Protector'), p.name, p.sku, p.alert_quantity, p.unit_id, p.category_id, p.brand_id, p.description, p.created_at,
			  COALESCE(pl.qty_available, 0) as location_qty,
			  pl.selling_price as location_price,
			  COALESCE(c.name, '') as category_name,
			  COALESCE(b.name, '') as brand_name
			  FROM products p
			  LEFT JOIN product_locations pl ON p.id = pl.product_id AND pl.location_id = ?
			  LEFT JOIN categories c ON p.category_id = c.id
			  LEFT JOIN brands b ON p.brand_id = b.id
			  WHERE p.tenant_id = ?`
	
	args := []interface{}{locationID, tenantID}

	if categoryID > 0 {
		query += " AND p.category_id = ?"
		args = append(args, categoryID)
	}
	if brandID > 0 {
		query += " AND p.brand_id = ?"
		args = append(args, brandID)
	}

	query += " ORDER BY p.id DESC"
	
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.TenantID, &p.ProductType, &p.Name, &p.SKU, &p.AlertQuantity, &p.UnitID, &p.CategoryID, &p.BrandID, &p.Description, &p.CreatedAt, &p.LocationQty, &p.LocationPrice, &p.CategoryName, &p.BrandName)
		if err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, nil
}

// GetProductsByTenantFiltered returns products filtered by product type as well
func (m *Models) GetProductsByTenantFiltered(tenantID int, locationID int, productType string, categoryID int, brandID int) ([]*Product, error) {
	query := `SELECT p.id, p.tenant_id, COALESCE(p.product_type, 'Protector'), p.name, p.sku, p.alert_quantity, p.unit_id, p.category_id, p.brand_id, p.description, p.created_at,
			  COALESCE(pl.qty_available, 0) as location_qty,
			  pl.selling_price as location_price,
			  COALESCE(c.name, '') as category_name,
			  COALESCE(b.name, '') as brand_name
			  FROM products p
			  LEFT JOIN product_locations pl ON p.id = pl.product_id AND pl.location_id = ?
			  LEFT JOIN categories c ON p.category_id = c.id
			  LEFT JOIN brands b ON p.brand_id = b.id
			  WHERE p.tenant_id = ?`
	
	args := []interface{}{locationID, tenantID}

	if productType != "" {
		query += " AND p.product_type = ?"
		args = append(args, productType)
	}
	if categoryID > 0 {
		query += " AND p.category_id = ?"
		args = append(args, categoryID)
	}
	if brandID > 0 {
		query += " AND p.brand_id = ?"
		args = append(args, brandID)
	}

	query += " ORDER BY p.product_type, c.name, b.name, p.name"
	
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.TenantID, &p.ProductType, &p.Name, &p.SKU, &p.AlertQuantity, &p.UnitID, &p.CategoryID, &p.BrandID, &p.Description, &p.CreatedAt, &p.LocationQty, &p.LocationPrice, &p.CategoryName, &p.BrandName)
		if err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, nil
}

func (m *Models) GetProductByID(id int, tenantID int) (*Product, error) {
	query := `SELECT p.id, p.tenant_id, COALESCE(p.product_type, 'Protector'), p.name, p.sku, p.alert_quantity, p.unit_id, p.category_id, p.brand_id, p.description, p.created_at,
			  COALESCE(c.name, '') as category_name, COALESCE(b.name, '') as brand_name
			  FROM products p
			  LEFT JOIN categories c ON p.category_id = c.id
			  LEFT JOIN brands b ON p.brand_id = b.id
			  WHERE p.id = ? AND p.tenant_id = ?`
	var p Product
	err := m.DB.QueryRow(query, id, tenantID).Scan(&p.ID, &p.TenantID, &p.ProductType, &p.Name, &p.SKU, &p.AlertQuantity, &p.UnitID, &p.CategoryID, &p.BrandID, &p.Description, &p.CreatedAt, &p.CategoryName, &p.BrandName)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *Models) InsertProduct(p *Product, locationStocks map[int]float64) (int64, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `INSERT INTO products (tenant_id, product_type, name, sku, alert_quantity, unit_id, category_id, brand_id, description)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(query, p.TenantID, p.ProductType, p.Name, p.SKU, p.AlertQuantity, p.UnitID, p.CategoryID, p.BrandID, p.Description)
	if err != nil {
		return 0, err
	}
	productID, _ := res.LastInsertId()

	for locID, stock := range locationStocks {
		_, err = tx.Exec("INSERT INTO product_locations (product_id, location_id, qty_available) VALUES (?, ?, ?)", productID, locID, stock)
		if err != nil {
			return 0, err
		}
	}

	return productID, tx.Commit()
}

func (m *Models) UpdateProduct(p *Product) error {
	query := `UPDATE products SET product_type = ?, name = ?, sku = ?, alert_quantity = ?, unit_id = ?, category_id = ?, brand_id = ?, description = ?
			  WHERE id = ? AND tenant_id = ?`
	_, err := m.DB.Exec(query, p.ProductType, p.Name, p.SKU, p.AlertQuantity, p.UnitID, p.CategoryID, p.BrandID, p.Description, p.ID, p.TenantID)
	return err
}

func (m *Models) GetCategoriesByTenant(tenantID int) ([]*Category, error) {
	rows, err := m.DB.Query("SELECT id, tenant_id, parent_id, name, description FROM categories WHERE tenant_id = ? ORDER BY parent_id, name", tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []*Category
	for rows.Next() {
		var c Category
		rows.Scan(&c.ID, &c.TenantID, &c.ParentID, &c.Name, &c.Description)
		cats = append(cats, &c)
	}
	return cats, nil
}

func (m *Models) InsertCategory(c *Category) (int64, error) {
	res, err := m.DB.Exec("INSERT INTO categories (tenant_id, parent_id, name, description) VALUES (?, ?, ?, ?)", c.TenantID, c.ParentID, c.Name, c.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *Models) GetBrandsByTenant(tenantID int) ([]*Brand, error) {
	rows, err := m.DB.Query("SELECT id, tenant_id, name, description FROM brands WHERE tenant_id = ?", tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var brands []*Brand
	for rows.Next() {
		var b Brand
		rows.Scan(&b.ID, &b.TenantID, &b.Name, &b.Description)
		brands = append(brands, &b)
	}
	return brands, nil
}

func (m *Models) InsertBrand(b *Brand) (int64, error) {
	res, err := m.DB.Exec("INSERT INTO brands (tenant_id, name, description) VALUES (?, ?, ?)", b.TenantID, b.Name, b.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *Models) GetUnitsByTenant(tenantID int) ([]*Unit, error) {
	rows, err := m.DB.Query("SELECT id, tenant_id, actual_name, short_name, allow_decimal FROM units WHERE tenant_id = ?", tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var units []*Unit
	for rows.Next() {
		var u Unit
		rows.Scan(&u.ID, &u.TenantID, &u.ActualName, &u.ShortName, &u.AllowDecimal)
		units = append(units, &u)
	}
	return units, nil
}

func (m *Models) InsertUnit(u *Unit) (int64, error) {
	res, err := m.DB.Exec("INSERT INTO units (tenant_id, actual_name, short_name, allow_decimal) VALUES (?, ?, ?, ?)", u.TenantID, u.ActualName, u.ShortName, u.AllowDecimal)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetLowStockProducts returns products below their alert_quantity threshold
func (m *Models) GetLowStockProductsByLocation(tenantID int, locationID int) ([]*Product, error) {
	query := `SELECT p.id, p.name, p.sku, COALESCE(p.product_type, 'Protector'), 
			  COALESCE(SUM(pl.qty_available), 0) as total_qty, COALESCE(p.alert_quantity, 50)
			  FROM products p
			  LEFT JOIN product_locations pl ON p.id = pl.product_id
			  WHERE p.tenant_id = ?`
	
	params := []interface{}{tenantID}
	if locationID > 0 {
		query += " AND pl.location_id = ? "
		params = append(params, locationID)
	}

	query += ` GROUP BY p.id
			  HAVING total_qty < COALESCE(p.alert_quantity, 50)
			  ORDER BY total_qty ASC`
	
	rows, err := m.DB.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var p Product
		var alertQty float64
		rows.Scan(&p.ID, &p.Name, &p.SKU, &p.ProductType, &p.LocationQty, &alertQty)
		p.AlertQuantity = &alertQty
		products = append(products, &p)
	}
	return products, nil
}

func (m *Models) GetLowStockProducts(tenantID int) ([]*Product, error) {
	return m.GetLowStockProductsByLocation(tenantID, 0)
}

func (m *Models) UpdateCategory(c *Category) error {
	_, err := m.DB.Exec("UPDATE categories SET name=?, description=?, parent_id=? WHERE id=? AND tenant_id=?", c.Name, c.Description, c.ParentID, c.ID, c.TenantID)
	return err
}

func (m *Models) DeleteCategory(id int, tenantID int) error {
	_, err := m.DB.Exec("DELETE FROM categories WHERE id=? AND tenant_id=?", id, tenantID)
	return err
}

func (m *Models) UpdateBrand(b *Brand) error {
	_, err := m.DB.Exec("UPDATE brands SET name=?, description=? WHERE id=? AND tenant_id=?", b.Name, b.Description, b.ID, b.TenantID)
	return err
}

func (m *Models) DeleteBrand(id int, tenantID int) error {
	_, err := m.DB.Exec("DELETE FROM brands WHERE id=? AND tenant_id=?", id, tenantID)
	return err
}

func (m *Models) UpdateUnit(u *Unit) error {
	_, err := m.DB.Exec("UPDATE units SET actual_name=?, short_name=?, allow_decimal=? WHERE id=? AND tenant_id=?", u.ActualName, u.ShortName, u.AllowDecimal, u.ID, u.TenantID)
	return err
}

func (m *Models) DeleteUnit(id int, tenantID int) error {
	_, err := m.DB.Exec("DELETE FROM units WHERE id=? AND tenant_id=?", id, tenantID)
	return err
}

func (m *Models) DeleteProduct(id int, tenantID int) error {
	_, err := m.DB.Exec("DELETE FROM products WHERE id=? AND tenant_id=?", id, tenantID)
	return err
}
