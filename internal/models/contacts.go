package models

import (
	"database/sql"
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
	CreditLimit    *float64
	InvoiceDueDays *int
	CreatedAt      time.Time
}

func (m *Models) GetContactsByTenant(tenantID int, contactType string) ([]*Contact, error) {
	query := `SELECT id, tenant_id, type, name, COALESCE(business_name, ''), 
			  COALESCE(email, ''), mobile, COALESCE(tax_number, ''), 
			  COALESCE(opening_balance, 0), COALESCE(address, ''), 
			  COALESCE(city, ''), COALESCE(state, ''), COALESCE(country, ''), 
			  COALESCE(zip_code, ''), credit_limit, invoice_due_days, created_at 
			  FROM contacts WHERE tenant_id = ?`
	
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
		err := rows.Scan(&c.ID, &c.TenantID, &c.Type, &c.Name, &c.BusinessName, &c.Email, &c.Mobile, &c.TaxNumber, &c.OpeningBalance, &c.Address, &c.City, &c.State, &c.Country, &c.ZipCode, &c.CreditLimit, &c.InvoiceDueDays, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, &c)
	}
	return contacts, nil
}

func (m *Models) GetContactByID(id int, tenantID int) (*Contact, error) {
	query := `SELECT id, tenant_id, type, name, COALESCE(business_name, ''), 
			  COALESCE(email, ''), mobile, COALESCE(tax_number, ''), 
			  COALESCE(opening_balance, 0), COALESCE(address, ''), 
			  COALESCE(city, ''), COALESCE(state, ''), COALESCE(country, ''), 
			  COALESCE(zip_code, ''), credit_limit, invoice_due_days, created_at 
			  FROM contacts WHERE id = ? AND tenant_id = ?`
	
	var c Contact
	err := m.DB.QueryRow(query, id, tenantID).Scan(
		&c.ID, &c.TenantID, &c.Type, &c.Name, &c.BusinessName,
		&c.Email, &c.Mobile, &c.TaxNumber, &c.OpeningBalance, &c.Address,
		&c.City, &c.State, &c.Country, &c.ZipCode, &c.CreditLimit, &c.InvoiceDueDays, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m *Models) GetContactByName(tenantID int, name string) (*Contact, error) {
	query := `SELECT id, name, credit_limit, invoice_due_days FROM contacts WHERE tenant_id = ? AND name = ? LIMIT 1`
	var c Contact
	err := m.DB.QueryRow(query, tenantID, name).Scan(&c.ID, &c.Name, &c.CreditLimit, &c.InvoiceDueDays)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m *Models) InsertContact(c *Contact) (int64, error) {
	// Auto-fill defaults from business settings if not specified
	if c.CreditLimit == nil || c.InvoiceDueDays == nil {
		var defLimit float64
		var defDays int
		err := m.DB.QueryRow("SELECT COALESCE(default_credit_limit, 0.0), COALESCE(default_invoice_due_days, 30) FROM business_settings WHERE tenant_id = ? LIMIT 1", c.TenantID).Scan(&defLimit, &defDays)
		if err == nil {
			if c.CreditLimit == nil {
				c.CreditLimit = &defLimit
			}
			if c.InvoiceDueDays == nil {
				c.InvoiceDueDays = &defDays
			}
		}
	}

	query := `INSERT INTO contacts (tenant_id, type, name, business_name, email, mobile, tax_number, opening_balance, address, city, state, country, zip_code, created_by, credit_limit, invoice_due_days)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	res, err := m.DB.Exec(query, c.TenantID, c.Type, c.Name, c.BusinessName, c.Email, c.Mobile, c.TaxNumber, c.OpeningBalance, c.Address, c.City, c.State, c.Country, c.ZipCode, c.CreatedBy, c.CreditLimit, c.InvoiceDueDays)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *Models) UpdateContact(c *Contact) error {
	query := `UPDATE contacts 
			  SET name = ?, business_name = ?, email = ?, mobile = ?, tax_number = ?, 
			      opening_balance = ?, address = ?, city = ?, state = ?, country = ?, 
			      zip_code = ?, credit_limit = ?, invoice_due_days = ?
			  WHERE id = ? AND tenant_id = ?`
	_, err := m.DB.Exec(query, c.Name, c.BusinessName, c.Email, c.Mobile, c.TaxNumber,
		c.OpeningBalance, c.Address, c.City, c.State, c.Country,
		c.ZipCode, c.CreditLimit, c.InvoiceDueDays, c.ID, c.TenantID)
	return err
}
