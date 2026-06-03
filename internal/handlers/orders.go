package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

// OrderList shows all orders (filtered by role)
func (app *Application) OrderList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())
	locationID := middleware.GetLocationID(r.Context())
	
	statusFilter := r.URL.Query().Get("status")
	typeFilter := r.URL.Query().Get("type")

	start, end := app.ParseDateRange(r)
	preset := r.URL.Query().Get("preset")
	if preset == "" {
		if r.URL.Query().Get("start_date") != "" && r.URL.Query().Get("end_date") != "" {
			preset = "custom"
		} else {
			preset = "today"
		}
	}

	orders, err := app.Models.GetOrdersByTenant(tenantID, statusFilter, typeFilter, user.Role, locationID, user.ID, start, end)
	if err != nil {
		log.Printf("ERROR OrderList: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	summary, _ := app.Models.GetOrderSummary(tenantID, user.Role, locationID, user.ID, start, end)

	app.RenderPage(w, r, "orders/index", struct {
		Orders       []*models.StoreOrder
		Summary      *models.OrderSummary
		StatusFilter string
		TypeFilter   string
		UserRole     string
		Preset       string
		StartDate    string
		EndDate      string
	}{
		Orders:       orders,
		Summary:      summary,
		StatusFilter: statusFilter,
		TypeFilter:   typeFilter,
		UserRole:     user.Role,
		Preset:       preset,
		StartDate:    start.Format("2006-01-02"),
		EndDate:      end.Format("2006-01-02"),
	})
}

// OrderCreate shows the create order form (for ShopKeeper)
func (app *Application) OrderCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())

	user := middleware.GetUser(r.Context())
	
	if locationID == 0 && user.LocationID != nil {
		locationID = *user.LocationID
	}
	
	// Load the full tenant catalog so the searchable picker can find every product.
	products, _ := app.Models.GetProductsByTenantFiltered(tenantID, 0, "", 0, 0)
	categories, _ := app.Models.GetCategoriesByTenant(tenantID)
	brands, _ := app.Models.GetBrandsByTenant(tenantID)
	customers, _ := app.Models.GetContactsByTenant(tenantID, "customer")
	locations, _ := app.Models.GetLocationsByTenant(tenantID)

	orderType := r.URL.Query().Get("type")
	if orderType == "" {
		orderType = "StoreOrder"
	}

	sellingGroups, _ := app.Models.GetSellingPriceGroupsByTenant(tenantID)

	rows, err := app.DB.Query(`SELECT product_id, selling_price_group_id, price FROM product_group_prices WHERE tenant_id = ?`, tenantID)
	groupPrices := make(map[int]map[int]float64)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var pid, gid int
			var price float64
			rows.Scan(&pid, &gid, &price)
			if _, ok := groupPrices[pid]; !ok {
				groupPrices[pid] = make(map[int]float64)
			}
			groupPrices[pid][gid] = price
		}
	}

	groupPricesJSON, _ := json.Marshal(groupPrices)

	app.RenderPage(w, r, "orders/create", struct {
		Products        []*models.Product
		Categories      []*models.Category
		Brands          []*models.Brand
		Customers       []*models.Contact
		Locations       []*models.BusinessLocation
		OrderType       string
		SellingGroups   []*models.SellingPriceGroup
		GroupPricesJSON string
	}{
		Products:        products,
		Categories:      categories,
		Brands:          brands,
		Customers:       customers,
		Locations:       locations,
		OrderType:       orderType,
		SellingGroups:   sellingGroups,
		GroupPricesJSON: string(groupPricesJSON),
	})
}

// OrderStore processes a new order submission
func (app *Application) OrderStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}

	user := middleware.GetUser(r.Context())
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())

	r.ParseForm()

	orderType := r.FormValue("order_type")
	orderFrom := r.FormValue("order_from")
	notes := r.FormValue("notes")

	// Auto-create customer if it doesn't exist (Bulk Orders)
	if orderType == "BulkOrder" && orderFrom != "" {
		_, err := app.Models.GetContactByName(tenantID, orderFrom)
		if err != nil {
			// Doesn't exist, create it
			newContact := &models.Contact{
				TenantID:  tenantID,
				Type:      "customer",
				Name:      orderFrom,
				CreatedBy: &user.ID,
			}
			_, _ = app.Models.InsertContact(newContact)
		}
	}

	// Parse items
	productIDs := r.Form["product_id[]"]
	quantities := r.Form["quantity[]"]

	if len(productIDs) == 0 {
		http.Error(w, "No items in order", http.StatusBadRequest)
		return
	}

	refNo := app.Models.GenerateOrderRefNo(tenantID)

	// For Store Orders: no payment info needed
	// For Bulk Orders: payment tracking enabled
	paymentStatus := "unpaid"
	amountPaid := 0.0
	
	if orderType == "BulkOrder" {
		paymentStatus = r.FormValue("payment_status")
		if paymentStatus == "" {
			paymentStatus = "unpaid"
		}
		amountPaid, _ = strconv.ParseFloat(r.FormValue("amount_paid"), 64)
	}

	toLocID, _ := strconv.Atoi(r.FormValue("to_location_id"))
	if toLocID == 0 {
		toLocID = locationID
	}

	fromLocID, _ := strconv.Atoi(r.FormValue("from_location_id"))
	if orderType == "StoreOrder" {
		if fromLocID == 0 {
			http.Error(w, "Please select a source store", http.StatusBadRequest)
			return
		}
		if toLocID == 0 {
			http.Error(w, "Please select a destination shop", http.StatusBadRequest)
			return
		}
	}

	if orderType == "BulkOrder" && toLocID == 0 {
		http.Error(w, "Unable to determine the destination location", http.StatusBadRequest)
		return
	}
	
	order := &models.StoreOrder{
		TenantID:      tenantID,
		OrderType:     orderType,
		RefNo:         refNo,
		PlacedBy:      user.ID,
		OrderFrom:     orderFrom,
		ToLocationID:  toLocID,
		Status:        "pending",
		PaymentStatus: paymentStatus,
		AmountPaid:    amountPaid,
		Notes:         notes,
	}

	if orderType == "StoreOrder" {
		order.FromStoreID = &fromLocID
	} else if orderType == "BulkOrder" {
		locations, err := app.Models.GetLocationsByTenant(tenantID)
		if err != nil {
			http.Error(w, "Unable to load tenant locations", http.StatusInternalServerError)
			return
		}

		sourceStore := locationByID(locations, fromLocID)
		if sourceStore != nil && !strings.EqualFold(sourceStore.LocationType, "store") {
			http.Error(w, "Bulk orders must be assigned to a store location", http.StatusBadRequest)
			return
		}
		if sourceStore == nil {
			sourceStore = firstStoreLocation(locations)
		}
		if sourceStore == nil {
			http.Error(w, "No store location is available to process this order", http.StatusBadRequest)
			return
		}

		order.FromStoreID = &sourceStore.ID
	}

	unitPrices := r.Form["unit_price[]"]
	var totalAmount float64
	
	if orderType == "BulkOrder" {
		// Bulk Order: parse from_shop_qty and from_store_qty
		fromShopQtys := r.Form["from_shop_qty[]"]
		fromStoreQtys := r.Form["from_store_qty[]"]
		
		for i, pidStr := range productIDs {
			pid, _ := strconv.Atoi(pidStr)
			totalQty := 0.0
			if i < len(quantities) {
				totalQty, _ = strconv.ParseFloat(quantities[i], 64)
			}

			fromShop := 0.0
			fromStore := 0.0
			if i < len(fromShopQtys) {
				fromShop, _ = strconv.ParseFloat(fromShopQtys[i], 64)
			}
			if i < len(fromStoreQtys) {
				fromStore, _ = strconv.ParseFloat(fromStoreQtys[i], 64)
			}

			// If individual splits aren't provided, use total as from_store
			if fromShop == 0 && fromStore == 0 && totalQty > 0 {
				fromStore = totalQty
			}

			// Fetch product selling price (locations first, then products default, fallback to 0)
			var unitPrice float64
			if i < len(unitPrices) {
				unitPrice, _ = strconv.ParseFloat(unitPrices[i], 64)
			}
			if unitPrice == 0 {
				err := app.DB.QueryRow(`
					SELECT COALESCE(pl.selling_price, p.selling_price, 0)
					FROM products p
					LEFT JOIN product_locations pl ON p.id = pl.product_id AND pl.location_id = ?
					WHERE p.id = ? AND p.tenant_id = ?`, toLocID, pid, tenantID).Scan(&unitPrice)
				if err != nil {
					log.Printf("Error fetching price for product %d: %v", pid, err)
				}
			}

			subtotal := unitPrice * totalQty
			totalAmount += subtotal

			order.Items = append(order.Items, &models.OrderItem{
				ProductID:    pid,
				Quantity:     totalQty,
				FromShopQty:  fromShop,
				FromStoreQty: fromStore,
				UnitPrice:    unitPrice,
				Subtotal:     subtotal,
			})
		}
	} else {
		// Store Order: simple product + quantity
		for i, pidStr := range productIDs {
			pid, _ := strconv.Atoi(pidStr)
			qty := 0.0
			if i < len(quantities) {
				qty, _ = strconv.ParseFloat(quantities[i], 64)
			}

			// Fetch product selling price (locations first, then products default, fallback to 0)
			var unitPrice float64
			if i < len(unitPrices) {
				unitPrice, _ = strconv.ParseFloat(unitPrices[i], 64)
			}
			if unitPrice == 0 {
				err := app.DB.QueryRow(`
					SELECT COALESCE(pl.selling_price, p.selling_price, 0)
					FROM products p
					LEFT JOIN product_locations pl ON p.id = pl.product_id AND pl.location_id = ?
					WHERE p.id = ? AND p.tenant_id = ?`, toLocID, pid, tenantID).Scan(&unitPrice)
				if err != nil {
					log.Printf("Error fetching price for product %d: %v", pid, err)
				}
			}

			subtotal := unitPrice * qty
			totalAmount += subtotal

			order.Items = append(order.Items, &models.OrderItem{
				ProductID:    pid,
				Quantity:     qty,
				FromShopQty:  0,
				FromStoreQty: qty, // All from store for store orders
				UnitPrice:    unitPrice,
				Subtotal:     subtotal,
			})
		}
	}

	order.TotalAmount = totalAmount
	order.RemainingAmount = totalAmount - amountPaid

	if orderType == "BulkOrder" {
		if contact, err := app.Models.GetContactByName(tenantID, orderFrom); err == nil && contact != nil {
			if contact.CreditLimit != nil {
				limit := *contact.CreditLimit
				balance, err := app.Models.GetCustomerBalance(tenantID, orderFrom)
				if err == nil {
					remaining := totalAmount - amountPaid
					if balance + remaining > limit {
						repaymentNeeded := (balance + remaining) - limit
						currencySymbol := "TSh"
						if settings, err := app.Models.GetBusinessSettings(tenantID); err == nil {
							currencySymbol = settings.CurrencySymbol
						}
						http.Error(w, fmt.Sprintf("Order rejected: Customer credit limit exceeded. Current Limit: %s %.2f, Current Balance: %s %.2f. Repayment of at least %s %.2f is required.", 
							currencySymbol, limit, currencySymbol, balance, currencySymbol, repaymentNeeded), http.StatusBadRequest)
						return
					}
				}
			}
		}
	}

	orderID64, err := app.Models.InsertOrder(order)
	if err != nil {
		log.Printf("ERROR OrderStore: %v", err)
		http.Error(w, "Error creating order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Record initial payment in history if amount > 0 (Bulk Orders only)
	if order.AmountPaid > 0 {
		err = app.Models.UpdateOrderPayment(int(orderID64), tenantID, order.AmountPaid, "cash", "Initial order payment", user.ID)
		if err != nil {
			log.Printf("ERROR OrderStore Payment History: %v", err)
		}
	}

	http.Redirect(w, r, "/orders/invoice?id="+strconv.FormatInt(orderID64, 10), http.StatusSeeOther)
}

// OrderView shows order details
func (app *Application) OrderView(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	order, err := app.Models.GetOrderByID(id, tenantID)
	if err != nil {
		log.Printf("ERROR OrderView: %v", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	payments, _ := app.Models.GetOrderPayments(id)

	app.RenderPage(w, r, "orders/view", struct {
		Order    *models.StoreOrder
		Payments []*models.OrderPayment
		UserRole string
	}{
		Order:    order,
		Payments: payments,
		UserRole: user.Role,
	})
}

// OrderAccept handles order acceptance by Store Keeper
func (app *Application) OrderAccept(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())
	
	idStr := r.FormValue("order_id")
	id, _ := strconv.Atoi(idStr)

	locID := 0
	if user.LocationID != nil {
		locID = *user.LocationID
	}

	log.Printf("DEBUG: OrderAccept - ID: %v, UserID: %v, LocID: %v, TenantID: %v", id, user.ID, locID, tenantID)

	if id == 0 {
		http.Error(w, "Invalid Order ID", http.StatusBadRequest)
		return
	}

	err := app.Models.AcceptOrder(id, tenantID, user.ID, locID)
	if err != nil {
		log.Printf("ERROR OrderAccept Model: %v", err)
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			status = http.StatusNotFound
		}
		http.Error(w, "Error accepting order: "+err.Error(), status)
		return
	}

	http.Redirect(w, r, "/orders/view?id="+idStr, http.StatusSeeOther)
}

// OrderReject handles order rejection by Store Keeper
func (app *Application) OrderReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	idStr := r.FormValue("order_id")
	id, _ := strconv.Atoi(idStr)
	reason := r.FormValue("reason")
	if reason == "" {
		http.Error(w, "Please provide a rejection reason", http.StatusBadRequest)
		return
	}

	err := app.Models.RejectOrder(id, tenantID, user.ID, reason)
	if err != nil {
		log.Printf("ERROR OrderReject: %v", err)
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			status = http.StatusNotFound
		}
		http.Error(w, "Error rejecting order: "+err.Error(), status)
		return
	}

	http.Redirect(w, r, "/orders/view?id="+idStr, http.StatusSeeOther)
}

// OrderPaymentUpdate handles payment status updates (Bulk Orders only)
func (app *Application) OrderPaymentUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	idStr := r.FormValue("order_id")
	id, _ := strconv.Atoi(idStr)
	amountStr := r.FormValue("amount")
	amount, _ := strconv.ParseFloat(amountStr, 64)
	paymentMethod := r.FormValue("payment_method")
	if paymentMethod == "" {
		paymentMethod = "cash"
	}
	notes := r.FormValue("notes")

	err := app.Models.UpdateOrderPayment(id, tenantID, amount, paymentMethod, notes, user.ID)
	if err != nil {
		log.Printf("ERROR OrderPaymentUpdate: %v", err)
		http.Error(w, "Error updating payment", http.StatusInternalServerError)
		return
	}

	referer := r.Header.Get("Referer")
	if strings.Contains(referer, "invoice") {
		http.Redirect(w, r, "/orders/invoice?id="+idStr, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/orders/view?id="+idStr, http.StatusSeeOther)
	}
}

// PendingOrders shows pending orders for Store Keeper
func (app *Application) PendingOrdersList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())

	orders, err := app.Models.GetPendingOrdersForStoreKeeper(tenantID, locationID)
	if err != nil {
		log.Printf("ERROR PendingOrders: %v", err)
	}

	app.RenderPage(w, r, "orders/pending", struct {
		Orders []*models.StoreOrder
	}{
		Orders: orders,
	})
}

// OrderInvoice shows the invoice preview and print/download options
func (app *Application) OrderInvoice(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())
	
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	order, err := app.Models.GetOrderByID(id, tenantID)
	if err != nil {
		log.Printf("ERROR OrderInvoice: %v", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	payments, _ := app.Models.GetOrderPayments(id)
	business, _ := app.Models.GetBusinessSettings(tenantID)

	var invoiceDescription string
	_ = app.DB.QueryRow("SELECT COALESCE(invoice_description, '') FROM business_locations WHERE id = ?", order.ToLocationID).Scan(&invoiceDescription)
	if invoiceDescription == "" && order.FromStoreID != nil {
		_ = app.DB.QueryRow("SELECT COALESCE(invoice_description, '') FROM business_locations WHERE id = ?", *order.FromStoreID).Scan(&invoiceDescription)
	}
	if invoiceDescription == "" && business != nil {
		invoiceDescription = business.DefaultInvoiceDescription
	}

	app.RenderPage(w, r, "orders/invoice", struct {
		Order              *models.StoreOrder
		Payments           []*models.OrderPayment
		UserRole           string
		Business           *models.BusinessSetting
		InvoiceDescription string
	}{
		Order:              order,
		Payments:           payments,
		UserRole:           user.Role,
		Business:           business,
		InvoiceDescription: invoiceDescription,
	})
}
