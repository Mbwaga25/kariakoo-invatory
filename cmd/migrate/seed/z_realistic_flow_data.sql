-- Seed data to match Realistic System Flow
-- Categories (Protector/Cover Types)
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (101, 1, '21D', '21D Screen Protector');
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (102, 1, 'Privacy', 'Privacy Screen Protector');
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (103, 1, 'Clear', 'Clear Screen Protector / Cover');
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (104, 1, 'Matte', 'Matte Screen Protector / Cover');
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (105, 1, 'Full Glue', 'Full Glue Screen Protector');

INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (201, 2, '21D', '21D Screen Protector');
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (202, 2, 'Privacy', 'Privacy Screen Protector');
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (203, 2, 'Clear', 'Clear Screen Protector / Cover');
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (204, 2, 'Matte', 'Matte Screen Protector / Cover');
INSERT IGNORE INTO categories (id, tenant_id, name, description) VALUES (205, 2, 'Full Glue', 'Full Glue Screen Protector');

-- Brands (Phone Models)
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (101, 1, 'iPhone 11', 'Apple iPhone 11');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (102, 1, 'iPhone 12', 'Apple iPhone 12');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (103, 1, 'iPhone 13', 'Apple iPhone 13');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (104, 1, 'Samsung S21', 'Samsung Galaxy S21');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (105, 1, 'Samsung S22', 'Samsung Galaxy S22');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (106, 1, 'Tecno Spark 10', 'Tecno Spark 10');

INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (201, 2, 'iPhone 11', 'Apple iPhone 11');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (202, 2, 'iPhone 12', 'Apple iPhone 12');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (203, 2, 'iPhone 13', 'Apple iPhone 13');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (204, 2, 'Samsung S21', 'Samsung Galaxy S21');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (205, 2, 'Samsung S22', 'Samsung Galaxy S22');
INSERT IGNORE INTO brands (id, tenant_id, name, description) VALUES (206, 2, 'Tecno Spark 10', 'Tecno Spark 10');

-- Products (Combinations)
-- Tenant 1
INSERT IGNORE INTO products (id, tenant_id, product_type, name, sku, category_id, brand_id, alert_quantity) 
VALUES (101, 1, 'Protector', 'Protector - 21D - iPhone 11', 'PRO-21D-IP11', 101, 101, 50);

INSERT IGNORE INTO products (id, tenant_id, product_type, name, sku, category_id, brand_id, alert_quantity) 
VALUES (102, 1, 'Protector', 'Protector - Privacy - iPhone 12', 'PRO-PRIV-IP12', 102, 102, 50);

INSERT IGNORE INTO products (id, tenant_id, product_type, name, sku, category_id, brand_id, alert_quantity) 
VALUES (103, 1, 'Cover', 'Cover - Clear - Samsung S21', 'COV-CLR-S21', 103, 104, 50);

-- Tenant 2
INSERT IGNORE INTO products (id, tenant_id, product_type, name, sku, category_id, brand_id, alert_quantity) 
VALUES (201, 2, 'Protector', 'Protector - 21D - iPhone 11', 'PRO-21D-IP11-T2', 201, 201, 50);

INSERT IGNORE INTO products (id, tenant_id, product_type, name, sku, category_id, brand_id, alert_quantity) 
VALUES (202, 2, 'Protector', 'Protector - Privacy - iPhone 12', 'PRO-PRIV-IP12-T2', 202, 202, 50);

INSERT IGNORE INTO products (id, tenant_id, product_type, name, sku, category_id, brand_id, alert_quantity) 
VALUES (203, 2, 'Cover', 'Cover - Clear - Samsung S21', 'COV-CLR-S21-T2', 203, 204, 50);

-- Stock for Tenant 1
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (101, 1, 100);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (101, 2, 50);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (102, 1, 30);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (103, 1, 20);

-- Stock for Tenant 2
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (201, 3, 100);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (201, 4, 50);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (202, 3, 30);
INSERT IGNORE INTO product_locations (product_id, location_id, qty_available) VALUES (203, 3, 20);
