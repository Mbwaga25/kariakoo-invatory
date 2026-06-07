package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) ProductCreate(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin", "StoreKeeper") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())

	categories, _ := app.Models.GetCategoriesByTenant(tenantID)
	brands, _ := app.Models.GetBrandsByTenant(tenantID)
	units, _ := app.Models.GetUnitsByTenant(tenantID)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)
	if user != nil && user.Role == "StoreKeeper" {
		locations = storeLocationsOnly(locations)
	}

	groups, _ := app.Models.GetSellingPriceGroupsByTenant(tenantID)

	app.RenderPage(w, r, "products/create", struct {
		Categories []*models.Category
		Brands     []*models.Brand
		Units      []*models.Unit
		Locations  []*models.BusinessLocation
<<<<<<< HEAD
		Groups     []*models.SellingPriceGroup
	}{
		Categories: categories,
		Brands:     brands,
		Units:      units,
		Locations:  locations,
		Groups:     groups,
	})
}

func (app *Application) ProductList(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin", "StoreKeeper") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())
	
	// Default location to the session location, unless overridden by filter
	locationID := middleware.GetLocationID(r.Context())
	if filterLocID, _ := strconv.Atoi(r.URL.Query().Get("location_id")); filterLocID > 0 {
		locationID = filterLocID
	}

	categoryID, _ := strconv.Atoi(r.URL.Query().Get("category_id"))
	brandID, _ := strconv.Atoi(r.URL.Query().Get("brand_id"))
	productType := r.URL.Query().Get("product_type")
	
	products, err := app.Models.GetProductsByTenantFiltered(tenantID, locationID, productType, categoryID, brandID)
	if err != nil {
		log.Printf("ERROR GetProductsByTenant: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	categories, _ := app.Models.GetCategoriesByTenant(tenantID)
	brands, _ := app.Models.GetBrandsByTenant(tenantID)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)
	if user != nil && user.Role == "StoreKeeper" {
		locations = storeLocationsOnly(locations)
	}

	app.RenderPage(w, r, "products/index", struct {
		Products    []*models.Product
		Categories  []*models.Category
		Brands      []*models.Brand
		Locations   []*models.BusinessLocation
		Filters     map[string]interface{}
	}{
		Products:   products,
		Categories: categories,
		Brands:     brands,
		Locations:  locations,
		Filters: map[string]interface{}{
			"CategoryID":  categoryID,
			"BrandID":     brandID,
			"LocationID":  locationID,
			"ProductType": productType,
		},
	})
}

func (app *Application) CategoryList(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	categories, err := app.Models.GetCategoriesByTenant(tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "categories/index", struct {
		Categories []*models.Category
	}{
		Categories: categories,
	})
}

func (app *Application) BrandList(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	brands, err := app.Models.GetBrandsByTenant(tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "brands/index", struct {
		Brands []*models.Brand
	}{
		Brands: brands,
	})
}

func (app *Application) UnitList(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	units, err := app.Models.GetUnitsByTenant(tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "units/index", struct {
		Units []*models.Unit
	}{
		Units: units,
	})
}

func (app *Application) ProductStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/products/create", http.StatusSeeOther)
		return
	}

	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin", "StoreKeeper") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	user := middleware.GetUser(r.Context())
	alertQty, _ := strconv.ParseFloat(r.FormValue("alert_quantity"), 64)
	
	var unitID, categoryID, brandID *int
	if id, _ := strconv.Atoi(r.FormValue("unit_id")); id > 0 { unitID = &id }
	if id, _ := strconv.Atoi(r.FormValue("category_id")); id > 0 { categoryID = &id }
	if id, _ := strconv.Atoi(r.FormValue("brand_id")); id > 0 { brandID = &id }

	productType := r.FormValue("product_type")
	if productType == "" {
		productType = "Protector"
	}

	// Auto-generate product name from type + category + brand
	name := r.FormValue("name")
	if name == "" {
		name = productType
		if categoryID != nil {
			cats, _ := app.Models.GetCategoriesByTenant(tenantID)
			for _, c := range cats {
				if c.ID == *categoryID {
					name += " - " + c.Name
					break
				}
			}
		}
		if brandID != nil {
			brands, _ := app.Models.GetBrandsByTenant(tenantID)
			for _, b := range brands {
				if b.ID == *brandID {
					name += " - " + b.Name
					break
				}
			}
		}
	}

	sku := r.FormValue("sku")
	if sku == "" {
		sku = "SKU-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	// Location Stocks
	r.ParseForm()
	locationType := r.FormValue("location_type")
	locationStocks := make(map[int]float64)
	locations, err := app.Models.GetLocationsByTenant(tenantID)
	if err != nil {
		http.Error(w, "Unable to load locations", http.StatusInternalServerError)
		return
	}
	if user != nil && user.Role == "StoreKeeper" {
		locations = storeLocationsOnly(locations)
	}
	allowedLocations := make(map[int]struct{}, len(locations))
	for _, loc := range locations {
		allowedLocations[loc.ID] = struct{}{}
	}

	if locationType == "all" || locationType == "" { // Fallback if missing
		for _, loc := range locations {
			locationStocks[loc.ID] = 0.0
		}
	} else {
		locationIDStrs := r.Form["location_ids"]
		for _, s := range locationIDStrs {
			id, _ := strconv.Atoi(s)
			if _, ok := allowedLocations[id]; !ok {
				continue
			}
			stockStr := r.FormValue("opening_stock_" + s)
			stock, _ := strconv.ParseFloat(stockStr, 64)
			locationStocks[id] = stock
		}
		if len(locationStocks) == 0 {
			http.Error(w, "Please select at least one location", http.StatusBadRequest)
			return
		}
	}

	p := &models.Product{
		TenantID:      tenantID,
		ProductType:   productType,
		Name:          name,
		SKU:           sku,
		AlertQuantity: app.Float64Ptr(alertQty),
		UnitID:        unitID,
		CategoryID:    categoryID,
		BrandID:       brandID,
		Description:   app.StringPtr(r.FormValue("description")),
	}

	productID, err := app.Models.InsertProduct(p, locationStocks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save selling price group values
	groups, _ := app.Models.GetSellingPriceGroupsByTenant(tenantID)
	for _, g := range groups {
		priceStr := r.FormValue("group_price_" + strconv.Itoa(g.ID))
		if priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err == nil {
				_ = app.Models.SetProductGroupPrice(tenantID, int(productID), g.ID, price)
			}
		}
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func (app *Application) CategoryStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/categories", http.StatusSeeOther)
		return
	}

	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	var parentID *int
	if pid, _ := strconv.Atoi(r.FormValue("parent_id")); pid > 0 {
		parentID = &pid
	}

	c := &models.Category{
		TenantID:    tenantID,
		ParentID:    parentID,
		Name:        r.FormValue("name"),
		Description: app.StringPtr(r.FormValue("description")),
	}

	id, err := app.Models.InsertCategory(c)
	_ = id
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (app *Application) BrandStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/brands", http.StatusSeeOther)
		return
	}

	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	b := &models.Brand{
		TenantID:    tenantID,
		Name:        r.FormValue("name"),
		Description: app.StringPtr(r.FormValue("description")),
	}

	_, err := app.Models.InsertBrand(b)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/brands", http.StatusSeeOther)
}

func (app *Application) UnitStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/units", http.StatusSeeOther)
		return
	}

	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	allowDecimal, _ := strconv.ParseBool(r.FormValue("allow_decimal"))
	
	u := &models.Unit{
		TenantID:     tenantID,
		ActualName:   r.FormValue("actual_name"),
		ShortName:    r.FormValue("short_name"),
		AllowDecimal: allowDecimal,
	}

	unitID, err := app.Models.InsertUnit(u)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Smart Unit Conversion
	baseUnitID, _ := strconv.Atoi(r.FormValue("base_unit_id"))
	convValue, _ := strconv.ParseFloat(r.FormValue("operation_value"), 64)
	if baseUnitID > 0 && convValue > 0 {
		query := "INSERT INTO unit_conversions (unit_id, base_unit_id, operator, operation_value) VALUES (?, ?, ?, ?)"
		app.DB.Exec(query, unitID, baseUnitID, "multiply", convValue)
	}

	http.Redirect(w, r, "/units", http.StatusSeeOther)
}

func (app *Application) ProductView(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	product, err := app.Models.GetProductByID(id, tenantID)
	if err != nil {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	// Fetch stock across ALL locations for this product
	query := `SELECT l.name, COALESCE(pl.qty_available, 0) as qty, l.city
			  FROM business_locations l
			  LEFT JOIN product_locations pl ON l.id = pl.location_id AND pl.product_id = ?
			  WHERE l.tenant_id = ?`
	
	rows, err := app.DB.Query(query, id, tenantID)
	var stock []struct {
		LocationName string
		Qty          float64
		City         string
	}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s struct {
				LocationName string
				Qty          float64
				City         string
			}
			rows.Scan(&s.LocationName, &s.Qty, &s.City)
			stock = append(stock, s)
		}
	}

	// Fetch recent stock history for this product
	start := time.Now().AddDate(0, 0, -30) // Last 30 days by default
	end := time.Now()
	history, _ := app.Models.GetStockHistory(tenantID, start, end, id)

	app.RenderPage(w, r, "products/view", struct {
		Product *models.Product
		Stock   []struct {
			LocationName string
			Qty          float64
			City         string
		}
		History []*models.StockMovement
	}{
		Product: product,
		Stock:   stock,
		History: history,
	})
}

func (app *Application) ProductEdit(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	product, err := app.Models.GetProductByID(id, tenantID)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	categories, _ := app.Models.GetCategoriesByTenant(tenantID)
	brands, _ := app.Models.GetBrandsByTenant(tenantID)
	units, _ := app.Models.GetUnitsByTenant(tenantID)

	groups, _ := app.Models.GetSellingPriceGroupsByTenant(tenantID)
	groupPrices, _ := app.Models.GetProductGroupPrices(tenantID, id)

	app.RenderPage(w, r, "products/edit", struct {
		Product     *models.Product
		Categories  []*models.Category
		Brands      []*models.Brand
		Units       []*models.Unit
		Groups      []*models.SellingPriceGroup
		GroupPrices map[int]float64
	}{
		Product:     product,
		Categories:  categories,
		Brands:      brands,
		Units:       units,
		Groups:      groups,
		GroupPrices: groupPrices,
	})
}

func (app *Application) ProductUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))

	alertQty, _ := strconv.ParseFloat(r.FormValue("alert_quantity"), 64)
	
	var unitID, categoryID, brandID *int
	if uid, _ := strconv.Atoi(r.FormValue("unit_id")); uid > 0 { unitID = &uid }
	if cid, _ := strconv.Atoi(r.FormValue("category_id")); cid > 0 { categoryID = &cid }
	if bid, _ := strconv.Atoi(r.FormValue("brand_id")); bid > 0 { brandID = &bid }

	productType := r.FormValue("product_type")
	if productType == "" {
		productType = "Protector"
	}

	p := &models.Product{
		ID:            id,
		TenantID:      tenantID,
		ProductType:   productType,
		Name:          r.FormValue("name"),
		SKU:           r.FormValue("sku"),
		AlertQuantity: app.Float64Ptr(alertQty),
		UnitID:        unitID,
		CategoryID:    categoryID,
		BrandID:       brandID,
		Description:   app.StringPtr(r.FormValue("description")),
	}

	err := app.Models.UpdateProduct(p)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Update selling price group values
	groups, _ := app.Models.GetSellingPriceGroupsByTenant(tenantID)
	for _, g := range groups {
		priceStr := r.FormValue("group_price_" + strconv.Itoa(g.ID))
		if priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err == nil {
				_ = app.Models.SetProductGroupPrice(tenantID, id, g.ID, price)
			}
		}
	}

	http.Redirect(w, r, "/products/view?id="+strconv.Itoa(id), http.StatusSeeOther)
}

func (app *Application) ApiCategoryStore(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	var parentID *int
	if pid, _ := strconv.Atoi(r.FormValue("parent_id")); pid > 0 {
		parentID = &pid
	}

	c := &models.Category{
		TenantID:    tenantID,
		ParentID:    parentID,
		Name:        r.FormValue("name"),
		Description: app.StringPtr(r.FormValue("description")),
	}

	id, err := app.Models.InsertCategory(c)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true, "id": id})
}

func (app *Application) ApiBrandStore(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	b := &models.Brand{
		TenantID:    tenantID,
		Name:        r.FormValue("name"),
		Description: app.StringPtr(r.FormValue("description")),
	}

	id, err := app.Models.InsertBrand(b)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true, "id": id})
}

func (app *Application) ApiUnitStore(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	allowDecimal, _ := strconv.ParseBool(r.FormValue("allow_decimal"))
	
	u := &models.Unit{
		TenantID:     tenantID,
		ActualName:   r.FormValue("actual_name"),
		ShortName:    r.FormValue("short_name"),
		AllowDecimal: allowDecimal,
	}

	id, err := app.Models.InsertUnit(u)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true, "id": id})
}

func (app *Application) ApiProductSearch(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())
	query := r.URL.Query().Get("q")
	
	if len(query) < 1 {
		app.jsonResponse(w, http.StatusOK, []interface{}{})
		return
	}

	sqlQuery := `SELECT p.id, p.name, p.sku, COALESCE(p.product_type, 'Protector'),
				 COALESCE(c.name, ''), COALESCE(b.name, ''),
				 COALESCE(pl.selling_price, p.selling_price, 0)
				 FROM products p
				 LEFT JOIN categories c ON p.category_id = c.id
				 LEFT JOIN brands b ON p.brand_id = b.id
				 LEFT JOIN product_locations pl ON p.id = pl.product_id AND pl.location_id = ?
				 WHERE p.tenant_id = ? AND (p.name LIKE ? OR p.sku LIKE ? OR c.name LIKE ? OR b.name LIKE ?) LIMIT 10`
	rows, err := app.DB.Query(sqlQuery, locationID, tenantID, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		app.jsonResponse(w, http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var (
			id         int
			name       string
			sku        string
			pType      string
			catName    string
			brandName  string
			price      float64
		)
		rows.Scan(&id, &name, &sku, &pType, &catName, &brandName, &price)
		products = append(products, map[string]interface{}{
			"id":       id,
			"name":     name,
			"sku":      sku,
			"type":     pType,
			"category": catName,
			"brand":    brandName,
			"price":    price,
		})
	}

	app.jsonResponse(w, http.StatusOK, products)
}

func (app *Application) ApiCategoryUpdate(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))
	
	var parentID *int
	if pid, _ := strconv.Atoi(r.FormValue("parent_id")); pid > 0 {
		parentID = &pid
	}

	c := &models.Category{
		ID:          id,
		TenantID:    tenantID,
		ParentID:    parentID,
		Name:        r.FormValue("name"),
		Description: app.StringPtr(r.FormValue("description")),
	}

	err := app.Models.UpdateCategory(c)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (app *Application) ApiCategoryDelete(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))
	
	err := app.Models.DeleteCategory(id, tenantID)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (app *Application) ApiBrandUpdate(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))
	
	b := &models.Brand{
		ID:          id,
		TenantID:    tenantID,
		Name:        r.FormValue("name"),
		Description: app.StringPtr(r.FormValue("description")),
	}

	err := app.Models.UpdateBrand(b)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (app *Application) ApiBrandDelete(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))
	
	err := app.Models.DeleteBrand(id, tenantID)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (app *Application) ApiUnitUpdate(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))
	allowDecimal, _ := strconv.ParseBool(r.FormValue("allow_decimal"))
	
	u := &models.Unit{
		ID:           id,
		TenantID:     tenantID,
		ActualName:   r.FormValue("actual_name"),
		ShortName:    r.FormValue("short_name"),
		AllowDecimal: allowDecimal,
	}

	err := app.Models.UpdateUnit(u)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (app *Application) ApiUnitDelete(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))
	
	err := app.Models.DeleteUnit(id, tenantID)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (app *Application) ApiProductDelete(w http.ResponseWriter, r *http.Request) {
	if !app.requireRoles(w, r, "SuperAdmin", "ShopAdmin") {
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))
	
	err := app.Models.DeleteProduct(id, tenantID)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true})
}
