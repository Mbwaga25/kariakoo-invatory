package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
	"kariakoo/inventory/internal/handlers"
	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables from system")
	}

	// Build DSN from environment variables
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Connect to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	// Ensure the invoice_description column exists to avoid query errors
	if _, err := db.Exec(`ALTER TABLE business_locations ADD COLUMN IF NOT EXISTS invoice_description TEXT`); err != nil {
		log.Printf("WARNING: could not add invoice_description column: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Could not connect to database: ", err)
	}
	log.Println("Successfully connected to MySQL database:", os.Getenv("DB_NAME"))

	app := &handlers.Application{
		DB:     db,
		Models: models.NewModels(db),
	}

	mux := http.NewServeMux()

	// Serve static assets from ui/static
	wd, _ := os.Getwd()
	log.Printf("Current working directory: %s", wd)

	staticDir := "ui/static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		staticDir = "../ui/static"
		if _, err := os.Stat(staticDir); os.IsNotExist(err) {
			staticDir = "invatory/ui/static"
		}
	}
	
	absStaticDir, _ := filepath.Abs(staticDir)
	log.Printf("Serving static files from: %s", absStaticDir)

	
	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	
	mux.HandleFunc("/test-static", func(w http.ResponseWriter, r *http.Request) {
		exists := "NO"
		if _, err := os.Stat(staticDir); err == nil {
			exists = "YES"
		}
		fmt.Fprintf(w, "WD: %s\nStaticDir: %s\nAbsStaticDir: %s\nExists: %s", wd, staticDir, absStaticDir, exists)
	})




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
	dashboardChain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.Home)))
	mux.Handle("/", dashboardChain)
	mux.Handle("/dashboard", dashboardChain)

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
	mux.Handle("/api/categories/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiCategoryUpdate))))
	mux.Handle("/api/categories/delete", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiCategoryDelete))))
	
	mux.Handle("/api/brands/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiBrandStore))))
	mux.Handle("/api/brands/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiBrandUpdate))))
	mux.Handle("/api/brands/delete", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiBrandDelete))))
	
	mux.Handle("/api/units/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiUnitStore))))
	mux.Handle("/api/units/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiUnitUpdate))))
	mux.Handle("/api/units/delete", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiUnitDelete))))
	
	mux.Handle("/api/products/search", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiProductSearch))))
	mux.Handle("/api/products/delete", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ApiProductDelete))))
	mux.Handle("/api/contacts/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ContactStoreQuick))))
	mux.Handle("/api/contacts/credit-check", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ContactCreditCheck))))

	// Sales Routes
	mux.Handle("/sales", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SalesList))))
	mux.Handle("/pos/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SalesCreate))))
	mux.Handle("/pos/open-register", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.RegisterOpen))))

	// Purchase Routes
	mux.Handle("/purchases", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PurchaseList))))
	mux.Handle("/purchases/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PurchaseCreate))))
	mux.Handle("/purchases/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PurchaseStore))))
	mux.Handle("/purchases/view", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PurchaseView))))

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
	mux.Handle("/contacts/edit", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ContactEdit))))
	mux.Handle("/contacts/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ContactUpdate))))

	// Report Routes
	mux.Handle("/reports/profit-loss", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ProfitLossReport))))
	mux.Handle("/reports/stock", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockReport))))
	mux.Handle("/reports/stock-history", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.StockHistoryReport))))
	mux.Handle("/reports/register", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.RegisterReport))))
	mux.Handle("/reports/purchase-sell", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PurchaseSellReport))))
	mux.Handle("/reports/expense", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.ExpenseReport))))
	mux.Handle("/reports/orders", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderReport))))

	// Store Order Management Routes
	mux.Handle("/orders", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderList))))
	mux.Handle("/orders/create", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderCreate))))
	mux.Handle("/orders/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderStore))))
	mux.Handle("/orders/view", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderView))))
	mux.Handle("/orders/invoice", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderInvoice))))
	mux.Handle("/orders/accept", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderAccept))))
	mux.Handle("/orders/reject", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderReject))))
	mux.Handle("/orders/payment", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.OrderPaymentUpdate))))
	mux.Handle("/orders/pending", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.PendingOrdersList))))

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
	}

	// User Management Routes
	mux.Handle("/users/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.UserStore))))
	mux.Handle("/users", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.UserList))))
	mux.Handle("/users-list", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.UserList))))

	// Business Settings Routes
	mux.Handle("/business-settings", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.BusinessSettings))))
	mux.Handle("/business-settings/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.BusinessSettingsUpdate))))
	mux.Handle("/invoice-settings", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.InvoiceSettings))))
	mux.Handle("/invoice-settings/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.InvoiceSettingsUpdate))))
	mux.Handle("/business-location", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationSettings))))
	mux.Handle("/business-location/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationStore))))
	mux.Handle("/business-location/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationUpdate))))
	mux.Handle("/business-location/delete", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationDelete))))
	mux.Handle("/business-location/switch", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.LocationSwitch))))
	mux.Handle("/settings/modules", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SettingsModules))))
	
	mux.Handle("/selling-price-groups", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SellingPriceGroupList))))
	mux.Handle("/selling-price-groups/store", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SellingPriceGroupStore))))
	mux.Handle("/selling-price-groups/update", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SellingPriceGroupUpdate))))
	mux.Handle("/selling-price-groups/delete", middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(app.SellingPriceGroupDelete))))

	for route, tmpl := range routes {
		tName := tmpl // capture for closure
		chain := middleware.RequireAuthentication(middleware.TenantContext(&app.Models)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.RenderPage(w, r, tName, tName)
		})))
		mux.Handle(route, chain)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8081"
	}

	log.Println("Starting Inventory server on :" + port)
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
