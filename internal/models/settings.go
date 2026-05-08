package models

import (
	"time"
)

type BusinessSetting struct {
	ID                 int
	TenantID           int
	BusinessName       string
	StartDate          *time.Time
	Currency           string
	CurrencySymbol     string
	TimeZone           string
	TaxNumber          string
	TaxName            string
	FinancialYearStart string
	StockExpirySetting string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (m *Models) GetBusinessSettings(tenantID int) (*BusinessSetting, error) {
	query := `SELECT id, tenant_id, business_name, start_date, 
			  COALESCE(currency, 'USD'), 
			  COALESCE(currency_symbol, '$'), 
			  COALESCE(time_zone, 'UTC'), 
			  COALESCE(tax_number, ''), 
			  COALESCE(tax_name, ''), 
			  COALESCE(financial_year_start, 'January'), 
			  COALESCE(stock_expiry_setting, 'keep_stock')
			  FROM business_settings WHERE tenant_id = ? LIMIT 1`
	
	var s BusinessSetting
	err := m.DB.QueryRow(query, tenantID).Scan(&s.ID, &s.TenantID, &s.BusinessName, &s.StartDate, &s.Currency, &s.CurrencySymbol, &s.TimeZone, &s.TaxNumber, &s.TaxName, &s.FinancialYearStart, &s.StockExpirySetting)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (m *Models) UpdateBusinessSettings(s *BusinessSetting) error {
	query := `REPLACE INTO business_settings (tenant_id, business_name, start_date, currency, currency_symbol, time_zone, tax_number, tax_name, financial_year_start, stock_expiry_setting)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := m.DB.Exec(query, s.TenantID, s.BusinessName, s.StartDate, s.Currency, s.CurrencySymbol, s.TimeZone, s.TaxNumber, s.TaxName, s.FinancialYearStart, s.StockExpirySetting)
	return err
}

type Module struct {
	Key         string
	Name        string
	IsInstalled bool
}

func (m *Models) GetTenantModules(tenantID int) ([]Module, error) {
	rows, err := m.DB.Query("SELECT module_key, is_installed FROM tenant_modules WHERE tenant_id = ?", tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Define all possible modules
	allModules := map[string]string{
		"sales":             "Sales Management",
		"purchases":         "Purchase Management",
		"stock_adjustments": "Stock Adjustments",
		"stock_transfers":   "Stock Transfers",
		"expenses":          "Expense Tracking",
		"reports":           "Advanced Reporting",
	}

	installed := make(map[string]bool)
	for rows.Next() {
		var key string
		var status bool
		rows.Scan(&key, &status)
		installed[key] = status
	}

	var result []Module
	for key, name := range allModules {
		result = append(result, Module{
			Key:         key,
			Name:        name,
			IsInstalled: installed[key],
		})
	}
	return result, nil
}

func (m *Models) ToggleModule(tenantID int, moduleKey string, status bool) error {
	_, err := m.DB.Exec(`
		INSERT INTO tenant_modules (tenant_id, module_key, is_installed) 
		VALUES (?, ?, ?) 
		ON DUPLICATE KEY UPDATE is_installed = ?`, 
		tenantID, moduleKey, status, status)
	return err
}
