package models

import (
	"time"
)

type BusinessLocation struct {
	ID         int
	TenantID   int
	Name       string
	LocationID string
	City       string
	State      string
	Country    string
	ZipCode    string
	CreatedAt  time.Time
}

func (m *Models) GetLocationsByTenant(tenantID int) ([]*BusinessLocation, error) {
	query := `SELECT id, tenant_id, name, 
			  COALESCE(location_id, ''), 
			  COALESCE(city, ''), 
			  COALESCE(state, ''), 
			  COALESCE(country, ''), 
			  COALESCE(zip_code, ''), 
			  created_at 
			  FROM business_locations WHERE tenant_id = ? ORDER BY name ASC`
	
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*BusinessLocation
	for rows.Next() {
		var l BusinessLocation
		err := rows.Scan(&l.ID, &l.TenantID, &l.Name, &l.LocationID, &l.City, &l.State, &l.Country, &l.ZipCode, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		locations = append(locations, &l)
	}
	return locations, nil
}

func (m *Models) InsertLocation(l *BusinessLocation) (int64, error) {
	query := `INSERT INTO business_locations (tenant_id, name, location_id, city, state, country, zip_code)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := m.DB.Exec(query, l.TenantID, l.Name, l.LocationID, l.City, l.State, l.Country, l.ZipCode)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *Models) UpdateLocation(l *BusinessLocation) error {
	query := `UPDATE business_locations 
			  SET name = ?, location_id = ?, city = ?, state = ?, country = ?, zip_code = ?
			  WHERE id = ? AND tenant_id = ?`
	_, err := m.DB.Exec(query, l.Name, l.LocationID, l.City, l.State, l.Country, l.ZipCode, l.ID, l.TenantID)
	return err
}

func (m *Models) DeleteLocation(id int, tenantID int) error {
	query := "DELETE FROM business_locations WHERE id = ? AND tenant_id = ?"
	_, err := m.DB.Exec(query, id, tenantID)
	return err
}
