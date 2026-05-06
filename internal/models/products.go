package models

import (
	"database/sql"
	"time"
)

type Product struct {
	ID            int
	TenantID      int
	Name          string
	SKU           string
	PurchasePrice float64
	SellingPrice  float64
	AlertQuantity float64
	UnitID        *int
	CategoryID    *int
	BrandID       *int
	Description   string
	CreatedAt     time.Time
	
	// Location-specific details
	LocationQty   float64         `json:"location_qty"`
	LocationPrice sql.NullFloat64  `json:"location_price"`
}

type Category struct {
	ID          int
	TenantID    int
	ParentID    *int
	Name        string
	Description string
	CreatedAt   time.Time
}

type Brand struct {
	ID          int
	TenantID    int
	Name        string
	Description string
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

func (m *Models) GetProductsByTenant(tenantID int, locationID int) ([]*Product, error) {
	query := `SELECT p.id, p.tenant_id, p.name, p.sku, p.purchase_price, p.selling_price, p.alert_quantity, p.unit_id, p.category_id, p.brand_id, p.description, p.created_at,
			  COALESCE(pl.qty_available, 0) as location_qty,
			  pl.selling_price as location_price
			  FROM products p
			  LEFT JOIN product_locations pl ON p.id = pl.product_id AND pl.location_id = ?
			  WHERE p.tenant_id = ? ORDER BY p.id DESC`
	
	rows, err := m.DB.Query(query, locationID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.SKU, &p.PurchasePrice, &p.SellingPrice, &p.AlertQuantity, &p.UnitID, &p.CategoryID, &p.BrandID, &p.Description, &p.CreatedAt, &p.LocationQty, &p.LocationPrice)
		if err != nil {
			return nil, err
		}
		products = append(products, &p)
	}
	return products, nil
}

func (m *Models) GetProductByID(id int, tenantID int) (*Product, error) {
	query := `SELECT id, tenant_id, name, sku, purchase_price, selling_price, alert_quantity, unit_id, category_id, brand_id, description, created_at 
			  FROM products WHERE id = ? AND tenant_id = ?`
	var p Product
	err := m.DB.QueryRow(query, id, tenantID).Scan(&p.ID, &p.TenantID, &p.Name, &p.SKU, &p.PurchasePrice, &p.SellingPrice, &p.AlertQuantity, &p.UnitID, &p.CategoryID, &p.BrandID, &p.Description, &p.CreatedAt)
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

	query := `INSERT INTO products (tenant_id, name, sku, purchase_price, selling_price, alert_quantity, unit_id, category_id, brand_id, description)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(query, p.TenantID, p.Name, p.SKU, p.PurchasePrice, p.SellingPrice, p.AlertQuantity, p.UnitID, p.CategoryID, p.BrandID, p.Description)
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
	query := `UPDATE products SET name = ?, sku = ?, purchase_price = ?, selling_price = ?, alert_quantity = ?, unit_id = ?, category_id = ?, brand_id = ?, description = ?
			  WHERE id = ? AND tenant_id = ?`
	_, err := m.DB.Exec(query, p.Name, p.SKU, p.PurchasePrice, p.SellingPrice, p.AlertQuantity, p.UnitID, p.CategoryID, p.BrandID, p.Description, p.ID, p.TenantID)
	return err
}

func (m *Models) GetCategoriesByTenant(tenantID int) ([]*Category, error) {
	rows, err := m.DB.Query("SELECT id, tenant_id, name, description FROM categories WHERE tenant_id = ?", tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []*Category
	for rows.Next() {
		var c Category
		rows.Scan(&c.ID, &c.TenantID, &c.Name, &c.Description)
		cats = append(cats, &c)
	}
	return cats, nil
}

func (m *Models) InsertCategory(c *Category) (int64, error) {
	res, err := m.DB.Exec("INSERT INTO categories (tenant_id, name, description) VALUES (?, ?, ?)", c.TenantID, c.Name, c.Description)
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
