package handlers

import (
	"log"
	"net/http"
	"strconv"

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

	orders, err := app.Models.GetOrdersByTenant(tenantID, statusFilter, typeFilter, user.Role, locationID, user.ID)
	if err != nil {
		log.Printf("ERROR OrderList: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	summary, _ := app.Models.GetOrderSummary(tenantID, user.Role, locationID, user.ID)

	app.RenderPage(w, r, "orders/index", struct {
		Orders       []*models.StoreOrder
		Summary      *models.OrderSummary
		StatusFilter string
		TypeFilter   string
		UserRole     string
	}{
		Orders:       orders,
		Summary:      summary,
		StatusFilter: statusFilter,
		TypeFilter:   typeFilter,
		UserRole:     user.Role,
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
	
	products, _ := app.Models.GetProductsByTenant(tenantID, locationID, 0, 0)
	allLocations, _ := app.Models.GetLocationsByTenant(tenantID)
	
	var locations []*models.BusinessLocation
	if user.Role == "ShopKeeper" || user.Role == "StoreKeeper" {
		for _, loc := range allLocations {
			if loc.ID == locationID {
				locations = append(locations, loc)
				break
			}
		}
	} else {
		locations = allLocations
	}

	customers, _ := app.Models.GetContactsByTenant(tenantID, "customer")

	orderType := r.URL.Query().Get("type")
	if orderType == "" {
		orderType = "StoreOrder"
	}

	app.RenderPage(w, r, "orders/create", struct {
		Products  []*models.Product
		Locations []*models.BusinessLocation
		Customers []*models.Contact
		OrderType string
	}{
		Products:  products,
		Locations: locations,
		Customers: customers,
		OrderType: orderType,
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
	toLocationStr := r.FormValue("to_location_id")
	
	toLocationID := locationID
	if toLocationStr != "" {
		if id, err := strconv.Atoi(toLocationStr); err == nil {
			toLocationID = id
		}
	}

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

	paymentStatus := r.FormValue("payment_status")
	amountPaid, _ := strconv.ParseFloat(r.FormValue("amount_paid"), 64)

	order := &models.StoreOrder{
		TenantID:      tenantID,
		OrderType:     orderType,
		RefNo:         refNo,
		PlacedBy:      user.ID,
		OrderFrom:     orderFrom,
		ToLocationID:  toLocationID,
		Status:        "pending",
		PaymentStatus: paymentStatus,
		AmountPaid:    amountPaid,
		Notes:         notes,
	}

	var totalAmount float64
	for i, pidStr := range productIDs {
		pid, _ := strconv.Atoi(pidStr)
		qty := 0.0
		if i < len(quantities) {
			qty, _ = strconv.ParseFloat(quantities[i], 64)
		}

		// Get product price
		product, err := app.Models.GetProductByID(pid, tenantID)
		if err != nil {
			continue
		}

		subtotal := qty * product.SellingPrice
		totalAmount += subtotal

		order.Items = append(order.Items, &models.OrderItem{
			ProductID: pid,
			Quantity:  qty,
			UnitPrice: product.SellingPrice,
			Subtotal:  subtotal,
		})
	}

	order.TotalAmount = totalAmount
	order.RemainingAmount = totalAmount - amountPaid

	// Final check on payment status if it was set to "paid" but amounts don't match
	if order.PaymentStatus == "paid" && order.RemainingAmount > 0 {
		order.PaymentStatus = "incomplete"
	} else if order.PaymentStatus == "unpaid" && order.AmountPaid > 0 {
		order.PaymentStatus = "incomplete"
	}

	orderID64, err := app.Models.InsertOrder(order)
	if err != nil {
		log.Printf("ERROR OrderStore: %v", err)
		http.Error(w, "Error creating order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Record initial payment in history if amount > 0
	if order.AmountPaid > 0 {
		err = app.Models.UpdateOrderPayment(int(orderID64), tenantID, order.AmountPaid, user.ID)
		if err != nil {
			log.Printf("ERROR OrderStore Payment History: %v", err)
		}
	}

	http.Redirect(w, r, "/orders", http.StatusSeeOther)
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
	
	amountPaidStr := r.FormValue("amount_paid")
	amountPaid, _ := strconv.ParseFloat(amountPaidStr, 64)

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
		http.Error(w, "Error accepting order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if amountPaid > 0 {
		err = app.Models.UpdateOrderPayment(id, tenantID, amountPaid, user.ID)
		if err != nil {
			log.Printf("ERROR OrderAccept Payment: %v", err)
		}
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

	err := app.Models.RejectOrder(id, tenantID, user.ID, reason)
	if err != nil {
		log.Printf("ERROR OrderReject: %v", err)
		http.Error(w, "Error rejecting order", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/orders/view?id="+idStr, http.StatusSeeOther)
}

// OrderPaymentUpdate handles payment status updates
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

	err := app.Models.UpdateOrderPayment(id, tenantID, amount, user.ID)
	if err != nil {
		log.Printf("ERROR OrderPaymentUpdate: %v", err)
		http.Error(w, "Error updating payment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/orders/view?id="+idStr, http.StatusSeeOther)
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
