package main

import (
	"database/sql"
	"log"
	"net/http"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"kariakoo/inventory/internal/handlers"
	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func main() {
	// Connect to MySQL
	dsn := "root:@tcp(127.0.0.1:3306)/invatory?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Could not connect to database: ", err)
	}
	log.Println("Successfully connected to MySQL database 'invatory'")

	app := &handlers.Application{
		DB:     db,
		Models: models.NewModels(db),
	}

	mux := http.NewServeMux()

	// Serve static assets from ui/static
	staticDir := filepath.Join("ui", "static")
	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Public Routes
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			app.LoginPost(w, r)
		} else {
			app.Login(w, r)
		}
	})
	mux.HandleFunc("/logout", app.Logout)

	// Protected Routes (SaaS isolation)
	// We wrap the Home handler with both RequireAuthentication and TenantContext
	dashboardChain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.Home)))
	mux.Handle("/", dashboardChain)

	productCreateChain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ProductCreate)))
	mux.Handle("/products/create", productCreateChain)

	productListChain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ProductList)))
	mux.Handle("/products", productListChain)

	categoryListChain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.CategoryList)))
	mux.Handle("/categories", categoryListChain)

	brandListChain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.BrandList)))
	mux.Handle("/brands", brandListChain)

	unitListChain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.UnitList)))
	mux.Handle("/units", unitListChain)

	settingsChain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationSettings)))
	mux.Handle("/settings", settingsChain)

	// Inventory Store Routes
	mux.Handle("/products/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ProductStore))))
	mux.Handle("/products/view", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ProductView))))
	mux.Handle("/products/edit", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ProductEdit))))
	mux.Handle("/products/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ProductUpdate))))
	mux.Handle("/categories/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.CategoryStore))))
	mux.Handle("/brands/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.BrandStore))))
	mux.Handle("/units/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.UnitStore))))

	// Quick Add APIs
	mux.Handle("/api/categories/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiCategoryStore))))
	mux.Handle("/api/brands/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiBrandStore))))
	mux.Handle("/api/units/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiUnitStore))))
	mux.Handle("/api/products/search", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiProductSearch))))

	// Sales Routes
	mux.Handle("/sales", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SalesList))))
	mux.Handle("/pos/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SalesCreate))))
	mux.Handle("/pos/open-register", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.RegisterOpen))))

	// Purchase Routes
	mux.Handle("/purchases", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PurchaseList))))
	mux.Handle("/purchases/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PurchaseCreate))))
	mux.Handle("/purchases/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PurchaseStore))))

	// Stock Transfer Routes
	mux.Handle("/stock-transfers", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockTransferList))))
	mux.Handle("/stock-transfers/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockTransferCreate))))
	mux.Handle("/stock-transfers/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockTransferStore))))

	// Stock Adjustment Routes
	mux.Handle("/stock-adjustments", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockAdjustmentList))))
	mux.Handle("/stock-adjustments/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockAdjustmentCreate))))
	mux.Handle("/stock-adjustments/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockAdjustmentStore))))

	// Expense Routes
	mux.Handle("/expenses", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ExpenseList))))
	mux.Handle("/expenses/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ExpenseCreate))))
	mux.Handle("/expenses/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ExpenseStore))))

	// Contact Routes
	mux.Handle("/contacts", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ContactList))))
	mux.Handle("/contacts/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ContactCreate))))
	mux.Handle("/contacts/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ContactStore))))

	// Report Routes
	mux.Handle("/reports/profit-loss", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ProfitLossReport))))
	mux.Handle("/reports/stock", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockReport))))
	mux.Handle("/reports/register", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.RegisterReport))))

	// Sidebar Routes (Placeholder pages only)
	routes := map[string]string{
		"/variations":           "products/variations",
		"/import-products":      "products/import",
		"/import-opening-stock": "products/opening_stock",
		"/warranties":           "products/warranties",
		"/purchase-return":      "purchases/return",
		"/sell-return":          "sales/return",
		"/shipments":            "sales/shipments",
		"/discount":             "sales/discount",
		"/subscriptions":        "admin/subscriptions",
		"/tenants":              "admin/tenants",
		"/users":                "admin/users",
	}

	// User Management Routes
	mux.Handle("/users/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.UserStore))))
	mux.Handle("/users-list", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.UserList))))

	// Business Settings Routes
	mux.Handle("/business-settings", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.BusinessSettings))))
	mux.Handle("/business-settings/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.BusinessSettingsUpdate))))
	mux.Handle("/business-location", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationSettings))))
	mux.Handle("/business-location/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationStore))))
	mux.Handle("/business-location/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationUpdate))))
	mux.Handle("/business-location/delete", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationDelete))))
	mux.Handle("/business-location/switch", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationSwitch))))

	for route, tmpl := range routes {
		tName := tmpl // capture for closure
		chain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.RenderPage(w, r, tName, tName)
		})))
		mux.Handle(route, chain)
	}

	log.Println("Starting Inventory server on :8081...")
	err = http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}
