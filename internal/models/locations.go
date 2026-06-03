package models

import (
	"time"
)

type BusinessLocation struct {
	ID                 int
	TenantID           int
	TenantName         string
	Name               string
	LocationType       string
	LocationID         string
	City               string
	State              string
	Country            string
	ZipCode            string
	InvoiceDescription string
	CreatedAt          time.Time
}

func (m *Models) GetLocationsByTenant(tenantID int) ([]*BusinessLocation, error) {
	query := `SELECT id, tenant_id, name, 
			  COALESCE(location_type, 'shop'), 
			  COALESCE(location_id, ''), 
			  COALESCE(city, ''), 
			  COALESCE(state, ''), 
			  COALESCE(country, ''), 
			  COALESCE(zip_code, ''), 
			  COALESCE(invoice_description, ''),
			  created_at 
			  FROM business_locations WHERE tenant_id = ? ORDER BY location_type DESC, name ASC`
	
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*BusinessLocation
	for rows.Next() {
		var l BusinessLocation
		err := rows.Scan(&l.ID, &l.TenantID, &l.Name, &l.LocationType, &l.LocationID, &l.City, &l.State, &l.Country, &l.ZipCode, &l.InvoiceDescription, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		locations = append(locations, &l)
	}
	return locations, nil
}

func (m *Models) GetAllLocations() ([]*BusinessLocation, error) {
	query := `SELECT bl.id, bl.tenant_id, COALESCE(t.name, ''), bl.name,
			  COALESCE(bl.location_type, 'shop'),
			  COALESCE(bl.location_id, ''),
			  COALESCE(bl.city, ''),
			  COALESCE(bl.state, ''),
			  COALESCE(bl.country, ''),
			  COALESCE(bl.zip_code, ''),
			  COALESCE(bl.invoice_description, ''),
			  bl.created_at
			  FROM business_locations bl
			  LEFT JOIN tenants t ON bl.tenant_id = t.id
			  ORDER BY t.name, bl.location_type DESC, bl.name ASC`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*BusinessLocation
	for rows.Next() {
		var l BusinessLocation
		err := rows.Scan(&l.ID, &l.TenantID, &l.TenantName, &l.Name, &l.LocationType, &l.LocationID, &l.City, &l.State, &l.Country, &l.ZipCode, &l.InvoiceDescription, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		locations = append(locations, &l)
	}
	return locations, nil
}

func (m *Models) InsertLocation(l *BusinessLocation) (int64, error) {
	query := `INSERT INTO business_locations (tenant_id, name, location_type, location_id, city, state, country, zip_code, invoice_description)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := m.DB.Exec(query, l.TenantID, l.Name, l.LocationType, l.LocationID, l.City, l.State, l.Country, l.ZipCode, l.InvoiceDescription)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *Models) UpdateLocation(l *BusinessLocation) error {
	query := `UPDATE business_locations 
			  SET name = ?, location_type = ?, location_id = ?, city = ?, state = ?, country = ?, zip_code = ?, invoice_description = ?
			  WHERE id = ? AND tenant_id = ?`
	_, err := m.DB.Exec(query, l.Name, l.LocationType, l.LocationID, l.City, l.State, l.Country, l.ZipCode, l.InvoiceDescription, l.ID, l.TenantID)
	return err
}

func (m *Models) DeleteLocation(id int, tenantID int) error {
	query := "DELETE FROM business_locations WHERE id = ? AND tenant_id = ?"
	_, err := m.DB.Exec(query, id, tenantID)
	return err
}
