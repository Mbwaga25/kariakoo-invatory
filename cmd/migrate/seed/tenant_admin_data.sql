-- Seed data for Tenant 1 (Main Business)
-- Admin: shopadmin@test.com / 123456

-- Products for Tenant 1
INSERT IGNORE INTO products (id, tenant_id, name, sku, purchase_price, selling_price) VALUES (20, 1, 'Bread', 'BRD01', 1.00, 2.50);
INSERT IGNORE INTO products (id, tenant_id, name, sku, purchase_price, selling_price) VALUES (21, 1, 'Milk', 'MLK01', 0.80, 1.50);
INSERT IGNORE INTO products (id, tenant_id, name, sku, purchase_price, selling_price) VALUES (22, 1, 'Eggs', 'EGG01', 2.00, 4.00);

-- Product Locations for Tenant 1 (Location 1: Headquarters)
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (20, 1, 100.00);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (21, 1, 50.00);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (22, 1, 20.00);

-- Sales for Tenant 1 (Today)
INSERT INTO sales (tenant_id, business_location_id, invoice_no, transaction_date, status, payment_status, final_total, created_by) 
VALUES (1, 1, 'INV-MAIN-001', NOW(), 'final', 'paid', 250.00, 2);
SET @sale1 = LAST_INSERT_ID();
INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, line_total) VALUES (@sale1, 20, 100, 2.50, 250.00);
INSERT INTO transaction_payments (tenant_id, sale_id, amount, method, paid_on, created_by) VALUES (1, @sale1, 250.00, 'cash', NOW(), 2);

-- Purchases for Tenant 1 (Today)
INSERT INTO purchases (tenant_id, business_location_id, ref_no, purchase_date, status, payment_status, final_total, created_by)
VALUES (1, 1, 'PUR-MAIN-001', NOW(), 'received', 'paid', 100.00, 2);
SET @pur1 = LAST_INSERT_ID();
INSERT INTO purchase_items (purchase_id, product_id, quantity, purchase_price, line_total) VALUES (@pur1, 20, 100, 1.00, 100.00);

-- Expenses for Tenant 1 (Today)
INSERT IGNORE INTO expense_categories (id, tenant_id, name) VALUES (20, 1, 'Packaging');
INSERT INTO expenses (tenant_id, business_location_id, expense_category_id, ref_no, transaction_date, final_total, created_by)
VALUES (1, 1, 20, 'EXP-MAIN-001', NOW(), 50.00, 2);
