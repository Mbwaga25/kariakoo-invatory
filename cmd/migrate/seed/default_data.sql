-- Default Initial Data
-- Password for all users: 123456

-- Seed Tenant 1
INSERT IGNORE INTO tenants (id, name) VALUES (1, 'Main Business');

-- Seed Locations for Tenant 1
INSERT IGNORE INTO business_locations (id, tenant_id, name) VALUES (1, 1, 'Headquarters');
INSERT IGNORE INTO business_locations (id, tenant_id, name) VALUES (2, 1, 'Branch Office');

-- Seed Users
-- Super Admin (No tenant/location)
INSERT IGNORE INTO users (id, tenant_id, location_id, name, email, password_hash, role) 
VALUES (1, NULL, NULL, 'Super Admin', 'superadmin@test.com', '$2a$10$f38iMscz6l/J1rC6R3n0m.Tf8p9kZ3W5K.L5U7h8/Qk/T9yZ.1yO.', 'SuperAdmin');

-- Shop Admin for Tenant 1
INSERT IGNORE INTO users (id, tenant_id, location_id, name, email, password_hash, role) 
VALUES (2, 1, 1, 'Shop Admin', 'shopadmin@test.com', '$2a$10$f38iMscz6l/J1rC6R3n0m.Tf8p9kZ3W5K.L5U7h8/Qk/T9yZ.1yO.', 'ShopAdmin');

-- Shop Keeper for Tenant 1
INSERT IGNORE INTO users (id, tenant_id, location_id, name, email, password_hash, role) 
VALUES (3, 1, 2, 'Shop Keeper', 'shopkeeper@test.com', '$2a$10$f38iMscz6l/J1rC6R3n0m.Tf8p9kZ3W5K.L5U7h8/Qk/T9yZ.1yO.', 'ShopKeeper');

-- Store Keeper for Tenant 1
INSERT IGNORE INTO users (id, tenant_id, location_id, name, email, password_hash, role) 
VALUES (4, 1, 1, 'Store Keeper', 'storekeeper@test.com', '$2a$10$f38iMscz6l/J1rC6R3n0m.Tf8p9kZ3W5K.L5U7h8/Qk/T9yZ.1yO.', 'StoreKeeper');

-- Business Settings for Tenant 1
INSERT IGNORE INTO business_settings (tenant_id, business_name, currency, currency_symbol) 
VALUES (1, 'Main Business', 'TZS', 'TSh');

-- Enable store management module
INSERT IGNORE INTO tenant_modules (tenant_id, module_key, is_installed) VALUES (1, 'store_management', 1);
