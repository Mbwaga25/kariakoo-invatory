package models

import (
	"time"
)

type Contact struct {
	ID             int
	TenantID       int
	Type           string // 'supplier', 'customer', 'both'
	Name           string
	BusinessName   string
	Email          string
	Mobile         string
	TaxNumber      string
	OpeningBalance float64
	Address        string
	City           string
	State          string
	Country        string
	ZipCode        string
	CreatedBy      *int
	CreatedAt      time.Time
}

func (m *Models) GetContactsByTenant(tenantID int, contactType string) ([]*Contact, error) {
	query := "SELECT id, tenant_id, type, name, business_name, email, mobile, tax_number, opening_balance, address, city, state, country, zip_code, created_at FROM contacts WHERE tenant_id = ?"
	
	args := []interface{}{tenantID}
	if contactType != "" {
		query += " AND type = ?"
		args = append(args, contactType)
	}
	query += " ORDER BY name ASC"
	
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		var c Contact
		err := rows.Scan(&c.ID, &c.TenantID, &c.Type, &c.Name, &c.BusinessName, &c.Email, &c.Mobile, &c.TaxNumber, &c.OpeningBalance, &c.Address, &c.City, &c.State, &c.Country, &c.ZipCode, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, &c)
	}
	return contacts, nil
}

func (m *Models) GetContactByName(tenantID int, name string) (*Contact, error) {
	query := "SELECT id, name FROM contacts WHERE tenant_id = ? AND name = ? LIMIT 1"
	var c Contact
	err := m.DB.QueryRow(query, tenantID, name).Scan(&c.ID, &c.Name)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m *Models) InsertContact(c *Contact) (int64, error) {
	query := `INSERT INTO contacts (tenant_id, type, name, business_name, email, mobile, tax_number, opening_balance, address, city, state, country, zip_code, created_by)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	res, err := m.DB.Exec(query, c.TenantID, c.Type, c.Name, c.BusinessName, c.Email, c.Mobile, c.TaxNumber, c.OpeningBalance, c.Address, c.City, c.State, c.Country, c.ZipCode, c.CreatedBy)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
