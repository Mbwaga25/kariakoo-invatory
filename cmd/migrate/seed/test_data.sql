-- Test Data for Tenant 2 (Test Admin)
-- Password for all users: 123456

-- Ensure Tenant 2 exists
INSERT IGNORE INTO tenants (id, name) VALUES (2, 'Test Business Admin');

-- Add Locations for Tenant 2
INSERT IGNORE INTO business_locations (id, tenant_id, name) VALUES (3, 2, 'Main Shop');
INSERT IGNORE INTO business_locations (id, tenant_id, name) VALUES (4, 2, 'Warehouse');

-- Add Stores for Tenant 2 (needed for stock transfers FK)
INSERT IGNORE INTO stores (id, location_id, name) VALUES (3, 3, 'Main Shop Store');
INSERT IGNORE INTO stores (id, location_id, name) VALUES (4, 4, 'Warehouse Store');

-- Add User for Tenant 2 (Shop Admin)
-- Password hash for '123456'
INSERT IGNORE INTO users (id, tenant_id, location_id, name, email, password_hash, role) 
VALUES (4, 2, 3, 'Test Admin', 'testadmin@test.com', '$2a$10$f38iMscz6l/J1rC6R3n0m.Tf8p9kZ3W5K.L5U7h8/Qk/T9yZ.1yO.', 'ShopAdmin');

-- Business Settings for Tenant 2
INSERT IGNORE INTO business_settings (tenant_id, business_name, currency, currency_symbol) 
VALUES (2, 'Test Business', 'TZS', 'TSh');

-- Products for Tenant 2
INSERT IGNORE INTO products (id, tenant_id, name, sku, purchase_price, selling_price) VALUES (10, 2, 'iPhone 15', 'IP15', 800.00, 1200.00);
INSERT IGNORE INTO products (id, tenant_id, name, sku, purchase_price, selling_price) VALUES (11, 2, 'Samsung S24', 'S24', 700.00, 1100.00);
INSERT IGNORE INTO products (id, tenant_id, name, sku, purchase_price, selling_price) VALUES (12, 2, 'MacBook Air', 'MBA', 900.00, 1500.00);

-- Product Locations Initial Stock
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (10, 3, 50.00);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (11, 3, 30.00);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (12, 3, 5.00); -- Low stock for alert test

-- Sales (Sells)
-- Create a sale from yesterday
INSERT INTO sales (tenant_id, business_location_id, invoice_no, transaction_date, status, payment_status, final_total, created_by) 
VALUES (2, 3, 'INV-TEST-001', DATE_SUB(NOW(), INTERVAL 1 DAY), 'final', 'paid', 2400.00, 4);
SET @sale1 = LAST_INSERT_ID();
INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, line_total) VALUES (@sale1, 10, 2, 1200.00, 2400.00);
INSERT INTO transaction_payments (tenant_id, sale_id, amount, method, paid_on, created_by) VALUES (2, @sale1, 2400.00, 'cash', NOW(), 4);

-- Purchases
-- Create a purchase from 2 days ago
INSERT INTO purchases (tenant_id, business_location_id, ref_no, purchase_date, status, payment_status, final_total, created_by)
VALUES (2, 3, 'PUR-TEST-001', DATE_SUB(NOW(), INTERVAL 2 DAY), 'received', 'paid', 8000.00, 4);
SET @pur1 = LAST_INSERT_ID();
INSERT INTO purchase_items (purchase_id, product_id, quantity, purchase_price, line_total) VALUES (@pur1, 10, 10, 800.00, 8000.00);

-- Stock Transfers (Warehouse to Main Shop)
INSERT INTO stock_transfers (tenant_id, from_location_id, to_location_id, ref_no, transaction_date, status, notes, created_by)
VALUES (2, 4, 3, 'TR-TEST-001', DATE_SUB(NOW(), INTERVAL 3 DAY), 'received', 'Initial stock transfer', 4);
SET @tr1 = LAST_INSERT_ID();
INSERT INTO stock_transfer_items (transfer_id, product_id, quantity) VALUES (@tr1, 11, 10.00);

-- Stock Adjustments
INSERT INTO stock_adjustments (tenant_id, business_location_id, ref_no, transaction_date, adjustment_type, final_total, created_by)
VALUES (2, 3, 'ADJ-TEST-001', DATE_SUB(NOW(), INTERVAL 4 DAY), 'normal', 900.00, 4);
SET @adj1 = LAST_INSERT_ID();
INSERT INTO stock_adjustment_items (stock_adjustment_id, product_id, quantity, unit_price) VALUES (@adj1, 12, 1, 900.00);

-- Expenses
INSERT IGNORE INTO expense_categories (id, tenant_id, name) VALUES (10, 2, 'Rent');
INSERT IGNORE INTO expense_categories (id, tenant_id, name) VALUES (11, 2, 'Utilities');
INSERT INTO expenses (tenant_id, business_location_id, expense_category_id, ref_no, transaction_date, final_total, created_by)
VALUES (2, 3, 10, 'EXP-TEST-001', DATE_SUB(NOW(), INTERVAL 5 DAY), 500.00, 4);
INSERT INTO expenses (tenant_id, business_location_id, expense_category_id, ref_no, transaction_date, final_total, created_by)
VALUES (2, 3, 11, 'EXP-TEST-002', DATE_SUB(NOW(), INTERVAL 5 DAY), 200.00, 4);
