package handlers

import (
	"net/http"
	"strconv"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) ExpenseList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())
	
	expenses, err := app.Models.GetExpensesByTenant(tenantID, locationID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "expenses/index", struct {
		Expenses []*models.Expense
	}{
		Expenses: expenses,
	})
}

func (app *Application) ExpenseCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	
	categories, _ := app.Models.GetExpenseCategoriesByTenant(tenantID)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)

	app.RenderPage(w, r, "expenses/create", struct {
		Categories []*models.ExpenseCategory
		Locations  []*models.BusinessLocation
	}{
		Categories: categories,
		Locations:  locations,
	})
}

func (app *Application) ExpenseStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/expenses/create", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	r.ParseForm()

	locationID, _ := strconv.Atoi(r.FormValue("location_id"))
	categoryID, _ := strconv.Atoi(r.FormValue("expense_category_id"))
	refNo := r.FormValue("ref_no")
	amount, _ := strconv.ParseFloat(r.FormValue("final_total"), 64)
	
	e := &models.Expense{
		TenantID:           tenantID,
		BusinessLocationID: locationID,
		RefNo:              refNo,
		TransactionDate:    time.Now(),
		FinalTotal:         amount,
		CreatedBy:          &user.ID,
	}

	if categoryID > 0 {
		e.ExpenseCategoryID = &categoryID
	}

	_, err := app.Models.InsertExpense(e)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}
