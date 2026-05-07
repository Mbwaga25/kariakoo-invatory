package models

import (
	"time"
)

type ExpenseCategory struct {
	ID       int
	TenantID int
	Name     string
	Code     string
}

type Expense struct {
	ID                 int
	TenantID           int
	BusinessLocationID int
	ExpenseCategoryID  *int
	RefNo              string
	TransactionDate    time.Time
	FinalTotal         float64
	ExpenseFor         *int
	AdditionalNotes    string
	CreatedBy          *int
	CreatedAt          time.Time
	CategoryName       string
	LocationName       string
}

func (m *Models) GetExpensesByTenant(tenantID int, locationID int) ([]*Expense, error) {
	query := `SELECT e.id, e.tenant_id, e.business_location_id, e.expense_category_id, e.ref_no, e.transaction_date, e.final_total, ec.name as category_name, l.name as location_name
			  FROM expenses e
			  LEFT JOIN expense_categories ec ON e.expense_category_id = ec.id
			  JOIN business_locations l ON e.business_location_id = l.id
			  WHERE e.tenant_id = ?`
	
	args := []interface{}{tenantID}
	if locationID != 0 {
		query += " AND e.business_location_id = ?"
		args = append(args, locationID)
	}
	query += " ORDER BY e.transaction_date DESC"
	
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*Expense
	for rows.Next() {
		var e Expense
		err := rows.Scan(&e.ID, &e.TenantID, &e.BusinessLocationID, &e.ExpenseCategoryID, &e.RefNo, &e.TransactionDate, &e.FinalTotal, &e.CategoryName, &e.LocationName)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, &e)
	}

	return expenses, nil
}

func (m *Models) InsertExpense(e *Expense) (int64, error) {
	query := `INSERT INTO expenses (tenant_id, business_location_id, expense_category_id, ref_no, transaction_date, final_total, created_by)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	res, err := m.DB.Exec(query, e.TenantID, e.BusinessLocationID, e.ExpenseCategoryID, e.RefNo, e.TransactionDate, e.FinalTotal, e.CreatedBy)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (m *Models) GetExpenseCategoriesByTenant(tenantID int) ([]*ExpenseCategory, error) {
	query := "SELECT id, tenant_id, name, code FROM expense_categories WHERE tenant_id = ?"
	rows, err := m.DB.Query(query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*ExpenseCategory
	for rows.Next() {
		var ec ExpenseCategory
		err := rows.Scan(&ec.ID, &ec.TenantID, &ec.Name, &ec.Code)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &ec)
	}
	return categories, nil
}
