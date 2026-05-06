package handlers

import (
	"net/http"
	"strconv"
	"time"

	"kariakoo/inventory/internal/middleware"
	"kariakoo/inventory/internal/models"
)

func (app *Application) ProductCreate(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	categories, _ := app.Models.GetCategoriesByTenant(tenantID)
	brands, _ := app.Models.GetBrandsByTenant(tenantID)
	units, _ := app.Models.GetUnitsByTenant(tenantID)
	locations, _ := app.Models.GetLocationsByTenant(tenantID)

	app.RenderPage(w, r, "products/create", struct {
		Categories []*models.Category
		Brands     []*models.Brand
		Units      []*models.Unit
		Locations  []*models.BusinessLocation
	}{
		Categories: categories,
		Brands:     brands,
		Units:      units,
		Locations:  locations,
	})
}

func (app *Application) ProductList(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	locationID := middleware.GetLocationID(r.Context())
	
	products, err := app.Models.GetProductsByTenant(tenantID, locationID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.RenderPage(w, r, "products/index", struct {
		Products []*models.Product
	}{
		Products: products,
	})
}

func (app *Application) CategoryList(w http.ResponseWriter, r *http.Request) {
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

	tenantID := middleware.GetTenantID(r.Context())
	purchasePrice, _ := strconv.ParseFloat(r.FormValue("purchase_price"), 64)
	sellingPrice, _ := strconv.ParseFloat(r.FormValue("selling_price"), 64)
	alertQty, _ := strconv.ParseFloat(r.FormValue("alert_quantity"), 64)
	
	var unitID, categoryID, brandID *int
	if id, _ := strconv.Atoi(r.FormValue("unit_id")); id > 0 { unitID = &id }
	if id, _ := strconv.Atoi(r.FormValue("category_id")); id > 0 { categoryID = &id }
	if id, _ := strconv.Atoi(r.FormValue("brand_id")); id > 0 { brandID = &id }

	sku := r.FormValue("sku")
	if sku == "" {
		sku = "SKU-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	// Location Stocks
	r.ParseForm()
	locationIDStrs := r.Form["location_ids"]
	locationStocks := make(map[int]float64)
	for _, s := range locationIDStrs {
		id, _ := strconv.Atoi(s)
		stockStr := r.FormValue("opening_stock_" + s)
		stock, _ := strconv.ParseFloat(stockStr, 64)
		locationStocks[id] = stock
	}

	p := &models.Product{
		TenantID:      tenantID,
		Name:          r.FormValue("name"),
		SKU:           sku,
		PurchasePrice: purchasePrice,
		SellingPrice:  sellingPrice,
		AlertQuantity: alertQty,
		UnitID:        unitID,
		CategoryID:    categoryID,
		BrandID:       brandID,
		Description:   r.FormValue("description"),
	}

	_, err := app.Models.InsertProduct(p, locationStocks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func (app *Application) CategoryStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/categories", http.StatusSeeOther)
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
		Description: r.FormValue("description"),
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

	tenantID := middleware.GetTenantID(r.Context())
	b := &models.Brand{
		TenantID:    tenantID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
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

	app.RenderPage(w, r, "products/view", struct {
		Product *models.Product
		Stock   []struct {
			LocationName string
			Qty          float64
			City         string
		}
	}{
		Product: product,
		Stock:   stock,
	})
}

func (app *Application) ProductEdit(w http.ResponseWriter, r *http.Request) {
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

	app.RenderPage(w, r, "products/edit", struct {
		Product    *models.Product
		Categories []*models.Category
		Brands     []*models.Brand
		Units      []*models.Unit
	}{
		Product:    product,
		Categories: categories,
		Brands:     brands,
		Units:      units,
	})
}

func (app *Application) ProductUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	tenantID := middleware.GetTenantID(r.Context())
	id, _ := strconv.Atoi(r.FormValue("id"))

	purchasePrice, _ := strconv.ParseFloat(r.FormValue("purchase_price"), 64)
	sellingPrice, _ := strconv.ParseFloat(r.FormValue("selling_price"), 64)
	alertQty, _ := strconv.ParseFloat(r.FormValue("alert_quantity"), 64)
	
	var unitID, categoryID, brandID *int
	if uid, _ := strconv.Atoi(r.FormValue("unit_id")); uid > 0 { unitID = &uid }
	if cid, _ := strconv.Atoi(r.FormValue("category_id")); cid > 0 { categoryID = &cid }
	if bid, _ := strconv.Atoi(r.FormValue("brand_id")); bid > 0 { brandID = &bid }

	p := &models.Product{
		ID:            id,
		TenantID:      tenantID,
		Name:          r.FormValue("name"),
		SKU:           r.FormValue("sku"),
		PurchasePrice: purchasePrice,
		SellingPrice:  sellingPrice,
		AlertQuantity: alertQty,
		UnitID:        unitID,
		CategoryID:    categoryID,
		BrandID:       brandID,
		Description:   r.FormValue("description"),
	}

	err := app.Models.UpdateProduct(p)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/products/view?id="+strconv.Itoa(id), http.StatusSeeOther)
}

func (app *Application) ApiCategoryStore(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var parentID *int
	if pid, _ := strconv.Atoi(r.FormValue("parent_id")); pid > 0 {
		parentID = &pid
	}

	c := &models.Category{
		TenantID:    tenantID,
		ParentID:    parentID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	id, err := app.Models.InsertCategory(c)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true, "id": id})
}

func (app *Application) ApiBrandStore(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	b := &models.Brand{
		TenantID:    tenantID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	id, err := app.Models.InsertBrand(b)
	if err != nil {
		app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	app.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true, "id": id})
}

func (app *Application) ApiUnitStore(w http.ResponseWriter, r *http.Request) {
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
	query := r.URL.Query().Get("q")
	
	if len(query) < 1 {
		app.jsonResponse(w, http.StatusOK, []interface{}{})
		return
	}

	sqlQuery := "SELECT id, name, sku, selling_price FROM products WHERE tenant_id = ? AND (name LIKE ? OR sku LIKE ?) LIMIT 10"
	rows, err := app.DB.Query(sqlQuery, tenantID, "%"+query+"%", "%"+query+"%")
	if err != nil {
		app.jsonResponse(w, http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id int
		var name, sku string
		var price float64
		rows.Scan(&id, &name, &sku, &price)
		products = append(products, map[string]interface{}{
			"id": id,
			"name": name,
			"sku": sku,
			"price": price,
		})
	}

	app.jsonResponse(w, http.StatusOK, products)
}
