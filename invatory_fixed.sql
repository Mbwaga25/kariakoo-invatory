-- ============================================================
-- Kariakoo Inventory - Fixed Migration SQL
-- Generated: 2026-05-22
-- Fixes:
--   1. SET FOREIGN_KEY_CHECKS=0 to avoid import errors
--   2. DROP TABLE IF EXISTS for safe re-import
--   3. product_type DEFAULT 'Protector' enforced
--   4. Removed orphan product_locations rows (product_id 1,20-23)
--   5. Added missing user id=6 so cash_registers FK is satisfied
--   6. All products with full seed data included
-- ============================================================

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET FOREIGN_KEY_CHECKS = 0;
SET time_zone = "+00:00";

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

-- ============================================================
-- DROP TABLES (reverse dependency order)
-- ============================================================
DROP TABLE IF EXISTS `business_settings`;
DROP TABLE IF EXISTS `transaction_payments`;
DROP TABLE IF EXISTS `sale_items`;
DROP TABLE IF EXISTS `sales`;
DROP TABLE IF EXISTS `purchase_items`;
DROP TABLE IF EXISTS `purchases`;
DROP TABLE IF EXISTS `order_payments`;
DROP TABLE IF EXISTS `order_items`;
DROP TABLE IF EXISTS `orders`;
DROP TABLE IF EXISTS `stock_adjustment_items`;
DROP TABLE IF EXISTS `stock_adjustments`;
DROP TABLE IF EXISTS `stock_transfer_items`;
DROP TABLE IF EXISTS `stock_transfers`;
DROP TABLE IF EXISTS `product_locations`;
DROP TABLE IF EXISTS `products`;
DROP TABLE IF EXISTS `unit_conversions`;
DROP TABLE IF EXISTS `units`;
DROP TABLE IF EXISTS `categories`;
DROP TABLE IF EXISTS `brands`;
DROP TABLE IF EXISTS `expense_categories`;
DROP TABLE IF EXISTS `expenses`;
DROP TABLE IF EXISTS `contacts`;
DROP TABLE IF EXISTS `notifications`;
DROP TABLE IF EXISTS `cash_registers`;
DROP TABLE IF EXISTS `store_locations`;
DROP TABLE IF EXISTS `stores`;
DROP TABLE IF EXISTS `tenant_modules`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `business_locations`;
DROP TABLE IF EXISTS `tenants`;

-- ============================================================
-- CREATE TABLES
-- ============================================================

-- Table: tenants
CREATE TABLE `tenants` (
  `id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `subscription_plan` varchar(50) DEFAULT 'basic',
  `is_active` tinyint(1) DEFAULT 1,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: business_locations
CREATE TABLE `business_locations` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `location_type` enum('store','shop') DEFAULT 'shop',
  `location_id` varchar(100) DEFAULT NULL,
  `address` text DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `state` varchar(100) DEFAULT NULL,
  `country` varchar(100) DEFAULT NULL,
  `zip_code` varchar(20) DEFAULT NULL,
  `mobile` varchar(20) DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: users
CREATE TABLE `users` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) DEFAULT NULL,
  `location_id` int(11) DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `role` varchar(50) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: brands
CREATE TABLE `brands` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: categories
CREATE TABLE `categories` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `parent_id` int(11) DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: units
CREATE TABLE `units` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `actual_name` varchar(255) NOT NULL,
  `short_name` varchar(255) NOT NULL,
  `allow_decimal` tinyint(1) DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: unit_conversions
CREATE TABLE `unit_conversions` (
  `id` int(11) NOT NULL,
  `unit_id` int(11) NOT NULL,
  `base_unit_id` int(11) NOT NULL,
  `operator` enum('multiply','divide') DEFAULT 'multiply',
  `operation_value` decimal(15,4) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: products  (product_type DEFAULT 'Protector' is enforced)
CREATE TABLE `products` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `product_type` enum('Protector','Cover') NOT NULL DEFAULT 'Protector',
  `name` varchar(255) NOT NULL,
  `sku` varchar(100) DEFAULT NULL,
  `barcode_type` varchar(50) DEFAULT NULL,
  `unit_id` int(11) DEFAULT NULL,
  `brand_id` int(11) DEFAULT NULL,
  `category_id` int(11) DEFAULT NULL,
  `tax_id` int(11) DEFAULT NULL,
  `type` enum('single','variable','combo') DEFAULT 'single',
  `enable_stock` tinyint(1) DEFAULT 1,
  `alert_quantity` decimal(15,4) DEFAULT NULL,
  `purchase_price` decimal(15,4) DEFAULT 0.0000,
  `selling_price` decimal(15,4) DEFAULT 0.0000,
  `image` varchar(255) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: product_locations
CREATE TABLE `product_locations` (
  `id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `location_id` int(11) NOT NULL,
  `qty_available` decimal(15,4) DEFAULT 0.0000,
  `selling_price` decimal(15,4) DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT 1,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: business_settings
CREATE TABLE `business_settings` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `business_name` varchar(255) NOT NULL,
  `start_date` date DEFAULT NULL,
  `currency` varchar(10) DEFAULT 'USD',
  `currency_symbol` varchar(5) DEFAULT '$',
  `time_zone` varchar(100) DEFAULT 'UTC',
  `tax_number` varchar(50) DEFAULT NULL,
  `tax_name` varchar(50) DEFAULT NULL,
  `financial_year_start` enum('January','February','March','April','May','June','July','August','September','October','November','December') DEFAULT 'January',
  `stock_expiry_setting` enum('keep_stock','add_to_expired') DEFAULT 'keep_stock',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: cash_registers
CREATE TABLE `cash_registers` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `business_location_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `opening_amount` decimal(15,2) DEFAULT 0.00,
  `status` enum('open','close') DEFAULT 'open',
  `closed_at` datetime DEFAULT NULL,
  `closing_amount` decimal(15,2) DEFAULT 0.00,
  `total_card_slips` int(11) DEFAULT 0,
  `total_cheques` int(11) DEFAULT 0,
  `closing_note` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: contacts
CREATE TABLE `contacts` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `type` enum('supplier','customer','both') NOT NULL,
  `name` varchar(255) NOT NULL,
  `business_name` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobile` varchar(20) NOT NULL,
  `tax_number` varchar(50) DEFAULT NULL,
  `opening_balance` decimal(15,2) DEFAULT 0.00,
  `address` text DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `state` varchar(100) DEFAULT NULL,
  `country` varchar(100) DEFAULT NULL,
  `zip_code` varchar(20) DEFAULT NULL,
  `created_by` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: expense_categories
CREATE TABLE `expense_categories` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `code` varchar(50) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: expenses
CREATE TABLE `expenses` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `business_location_id` int(11) NOT NULL,
  `expense_category_id` int(11) DEFAULT NULL,
  `ref_no` varchar(50) NOT NULL,
  `transaction_date` datetime NOT NULL,
  `final_total` decimal(15,2) DEFAULT 0.00,
  `expense_for` int(11) DEFAULT NULL,
  `additional_notes` text DEFAULT NULL,
  `created_by` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: notifications
CREATE TABLE `notifications` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `title` varchar(255) NOT NULL,
  `message` text DEFAULT NULL,
  `is_read` tinyint(1) DEFAULT 0,
  `link` varchar(255) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: purchases
CREATE TABLE `purchases` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `business_location_id` int(11) NOT NULL,
  `supplier_id` int(11) DEFAULT NULL,
  `ref_no` varchar(50) NOT NULL,
  `purchase_date` datetime NOT NULL,
  `status` enum('ordered','received','pending') DEFAULT 'received',
  `payment_status` enum('paid','due','partial') DEFAULT 'due',
  `total_before_tax` decimal(15,2) DEFAULT 0.00,
  `tax_amount` decimal(15,2) DEFAULT 0.00,
  `discount_amount` decimal(15,2) DEFAULT 0.00,
  `final_total` decimal(15,2) DEFAULT 0.00,
  `created_by` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: purchase_items
CREATE TABLE `purchase_items` (
  `id` int(11) NOT NULL,
  `purchase_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `quantity` decimal(15,2) NOT NULL,
  `purchase_price` decimal(15,2) NOT NULL,
  `line_total` decimal(15,2) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: sales
CREATE TABLE `sales` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `business_location_id` int(11) DEFAULT NULL,
  `customer_id` int(11) DEFAULT NULL,
  `invoice_no` varchar(50) NOT NULL,
  `transaction_date` datetime NOT NULL,
  `due_date` date DEFAULT NULL,
  `status` enum('final','draft','proforma') DEFAULT 'final',
  `payment_status` enum('paid','due','partial') DEFAULT 'due',
  `total_before_tax` decimal(15,2) DEFAULT 0.00,
  `tax_amount` decimal(15,2) DEFAULT 0.00,
  `discount_amount` decimal(15,2) DEFAULT 0.00,
  `final_total` decimal(15,2) DEFAULT 0.00,
  `created_by` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: sale_items
CREATE TABLE `sale_items` (
  `id` int(11) NOT NULL,
  `sale_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `quantity` decimal(15,4) NOT NULL,
  `unit_price` decimal(15,2) NOT NULL,
  `line_total` decimal(15,2) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: transaction_payments
CREATE TABLE `transaction_payments` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `sale_id` int(11) NOT NULL,
  `amount` decimal(15,2) DEFAULT 0.00,
  `method` enum('cash','card','cheque','bank_transfer','other','custom_pay_1','custom_pay_2','custom_pay_3') DEFAULT 'cash',
  `transaction_no` varchar(100) DEFAULT NULL,
  `note` text DEFAULT NULL,
  `paid_on` datetime NOT NULL,
  `created_by` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: stock_adjustments
CREATE TABLE `stock_adjustments` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `business_location_id` int(11) NOT NULL,
  `ref_no` varchar(50) NOT NULL,
  `transaction_date` datetime NOT NULL,
  `adjustment_type` enum('normal','abnormal') DEFAULT 'normal',
  `final_total` decimal(15,2) DEFAULT 0.00,
  `total_amount_recovered` decimal(15,2) DEFAULT 0.00,
  `additional_notes` text DEFAULT NULL,
  `created_by` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: stock_adjustment_items
CREATE TABLE `stock_adjustment_items` (
  `id` int(11) NOT NULL,
  `stock_adjustment_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `quantity` decimal(15,2) NOT NULL,
  `unit_price` decimal(15,2) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: stock_transfers
CREATE TABLE `stock_transfers` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `from_location_id` int(11) NOT NULL,
  `to_location_id` int(11) NOT NULL,
  `ref_no` varchar(50) NOT NULL,
  `status` enum('pending','in_transit','received','cancelled') DEFAULT 'pending',
  `final_total` decimal(15,2) DEFAULT 0.00,
  `transaction_date` datetime NOT NULL,
  `received_at` timestamp NULL DEFAULT NULL,
  `created_by` int(11) DEFAULT NULL,
  `notes` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: stock_transfer_items
CREATE TABLE `stock_transfer_items` (
  `id` int(11) NOT NULL,
  `transfer_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `quantity` decimal(15,4) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: orders
CREATE TABLE `orders` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `order_type` enum('StoreOrder','BulkOrder') NOT NULL DEFAULT 'StoreOrder',
  `ref_no` varchar(50) NOT NULL,
  `placed_by` int(11) NOT NULL,
  `order_from` varchar(255) DEFAULT NULL,
  `from_shop_id` int(11) DEFAULT NULL,
  `from_store_id` int(11) DEFAULT NULL,
  `to_location_id` int(11) NOT NULL,
  `status` enum('pending','accepted','rejected','completed') NOT NULL DEFAULT 'pending',
  `payment_status` enum('paid','unpaid','incomplete') NOT NULL DEFAULT 'unpaid',
  `total_amount` decimal(15,2) NOT NULL DEFAULT 0.00,
  `amount_paid` decimal(15,2) NOT NULL DEFAULT 0.00,
  `remaining_amount` decimal(15,2) NOT NULL DEFAULT 0.00,
  `processed_by` int(11) DEFAULT NULL,
  `processed_at` datetime DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: order_items
CREATE TABLE `order_items` (
  `id` int(11) NOT NULL,
  `order_id` int(11) NOT NULL,
  `product_id` int(11) NOT NULL,
  `quantity` decimal(15,2) NOT NULL DEFAULT 0.00,
  `from_shop_qty` decimal(15,2) NOT NULL DEFAULT 0.00,
  `from_store_qty` decimal(15,2) NOT NULL DEFAULT 0.00,
  `unit_price` decimal(15,2) NOT NULL DEFAULT 0.00,
  `subtotal` decimal(15,2) NOT NULL DEFAULT 0.00
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: order_payments
CREATE TABLE `order_payments` (
  `id` int(11) NOT NULL,
  `order_id` int(11) NOT NULL,
  `amount` decimal(15,2) NOT NULL,
  `payment_method` varchar(50) DEFAULT 'cash',
  `notes` text DEFAULT NULL,
  `paid_by` int(11) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: stores
CREATE TABLE `stores` (
  `id` int(11) NOT NULL,
  `location_id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT 1,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Table: store_locations
CREATE TABLE `store_locations` (
  `id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Table: tenant_modules
CREATE TABLE `tenant_modules` (
  `id` int(11) NOT NULL,
  `tenant_id` int(11) NOT NULL,
  `module_key` varchar(50) NOT NULL,
  `is_installed` tinyint(1) DEFAULT 1,
  `installed_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- SEED DATA
-- ============================================================

-- tenants
INSERT INTO `tenants` (`id`, `name`, `subscription_plan`, `is_active`, `created_at`) VALUES
(1, 'Main Business', 'basic', 1, '2026-05-16 06:50:22'),
(2, 'Test Business Admin', 'basic', 1, '2026-05-16 06:50:22');

-- business_locations
INSERT INTO `business_locations` (`id`, `tenant_id`, `name`, `location_type`, `location_id`, `address`, `city`, `state`, `country`, `zip_code`, `mobile`, `email`, `created_at`) VALUES
(1, 1, 'Store 1', 'store', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2026-05-16 06:50:22'),
(2, 1, 'Store 2', 'store', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2026-05-16 06:50:22'),
(3, 2, 'Main Shop', 'shop', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2026-05-16 06:50:22'),
(4, 2, 'Warehouse', 'shop', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2026-05-16 06:50:22'),
(5, 1, 'Shop 1', 'shop', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2026-05-16 08:17:54'),
(6, 1, 'Shop 2', 'shop', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '2026-05-16 08:17:54');

-- users  (password hash = "admin123" bcrypt)
-- Added user id=6 and id=7 to satisfy cash_registers and contacts references
INSERT INTO `users` (`id`, `tenant_id`, `location_id`, `name`, `email`, `password_hash`, `role`, `created_at`) VALUES
(1, NULL, NULL, 'Super Admin',   'superadmin@test.com',    '$2a$10$US4Kx/FZUvtDOC.bOezYseNUDBLq0Vj//p2sdl1K8V4OKyuvHymlK', 'SuperAdmin',  '2026-05-16 06:50:22'),
(2, 1,    1,    'Shop Admin',    'shopadmin@test.com',     '$2a$10$US4Kx/FZUvtDOC.bOezYseNUDBLq0Vj//p2sdl1K8V4OKyuvHymlK', 'ShopAdmin',   '2026-05-16 06:50:22'),
(3, 1,    2,    'Shop Keeper',   'shopkeeper@test.com',    '$2a$10$US4Kx/FZUvtDOC.bOezYseNUDBLq0Vj//p2sdl1K8V4OKyuvHymlK', 'ShopKeeper',  '2026-05-16 06:50:22'),
(4, 1,    1,    'Store Keeper',  'storekeeper@test.com',   '$2a$10$US4Kx/FZUvtDOC.bOezYseNUDBLq0Vj//p2sdl1K8V4OKyuvHymlK', 'StoreKeeper', '2026-05-16 06:50:22'),
(5, 1,    5,    'Shop Keeper 2', 'shopkeeper2@test.com',   '$2a$10$US4Kx/FZUvtDOC.bOezYseNUDBLq0Vj//p2sdl1K8V4OKyuvHymlK', 'ShopKeeper',  '2026-05-16 06:50:22'),
(6, 1,    2,    'Cashier',       'cashier@test.com',       '$2a$10$US4Kx/FZUvtDOC.bOezYseNUDBLq0Vj//p2sdl1K8V4OKyuvHymlK', 'ShopKeeper',  '2026-05-16 06:50:22'),
(7, 1,    1,    'Staff 7',       'staff7@test.com',        '$2a$10$US4Kx/FZUvtDOC.bOezYseNUDBLq0Vj//p2sdl1K8V4OKyuvHymlK', 'ShopKeeper',  '2026-05-16 06:50:22');

-- brands
INSERT INTO `brands` (`id`, `tenant_id`, `name`, `description`, `created_at`) VALUES
(101, 1, 'iPhone 11',    'Apple iPhone 11',    '2026-05-16 06:50:22'),
(102, 1, 'iPhone 12',    'Apple iPhone 12',    '2026-05-16 06:50:22'),
(103, 1, 'iPhone 13',    'Apple iPhone 13',    '2026-05-16 06:50:22'),
(104, 1, 'Samsung S21',  'Samsung Galaxy S21', '2026-05-16 06:50:22'),
(105, 1, 'Samsung S22',  'Samsung Galaxy S22', '2026-05-16 06:50:22'),
(106, 1, 'Tecno Spark 10','Tecno Spark 10',    '2026-05-16 06:50:22'),
(201, 2, 'iPhone 11',    'Apple iPhone 11',    '2026-05-16 06:50:22'),
(202, 2, 'iPhone 12',    'Apple iPhone 12',    '2026-05-16 06:50:22'),
(203, 2, 'iPhone 13',    'Apple iPhone 13',    '2026-05-16 06:50:22'),
(204, 2, 'Samsung S21',  'Samsung Galaxy S21', '2026-05-16 06:50:22'),
(205, 2, 'Samsung S22',  'Samsung Galaxy S22', '2026-05-16 06:50:22'),
(206, 2, 'Tecno Spark 10','Tecno Spark 10',    '2026-05-16 06:50:22'),
(207, 1, 'tecno spark',  NULL,                 '2026-05-18 07:59:44');

-- categories
INSERT INTO `categories` (`id`, `tenant_id`, `parent_id`, `name`, `description`, `created_at`) VALUES
(101, 1, NULL, '21D',      '21D Screen Protector',             '2026-05-16 06:50:22'),
(102, 1, NULL, 'Privacy',  'Privacy Screen Protector',         '2026-05-16 06:50:22'),
(103, 1, NULL, 'Clear',    'Clear Screen Protector / Cover',   '2026-05-16 06:50:22'),
(104, 1, NULL, 'Matte',    'Matte Screen Protector / Cover',   '2026-05-16 06:50:22'),
(105, 1, NULL, 'Full Glue','Full Glue Screen Protector',       '2026-05-16 06:50:22'),
(201, 2, NULL, '21D',      '21D Screen Protector',             '2026-05-16 06:50:22'),
(202, 2, NULL, 'Privacy',  'Privacy Screen Protector',         '2026-05-16 06:50:22'),
(203, 2, NULL, 'Clear',    'Clear Screen Protector / Cover',   '2026-05-16 06:50:22'),
(204, 2, NULL, 'Matte',    'Matte Screen Protector / Cover',   '2026-05-16 06:50:22'),
(205, 2, NULL, 'Full Glue','Full Glue Screen Protector',       '2026-05-16 06:50:22'),
(207, 1, NULL, '31e',      NULL,                               '2026-05-18 07:59:27');

-- units
INSERT INTO `units` (`id`, `tenant_id`, `actual_name`, `short_name`, `allow_decimal`, `created_at`) VALUES
(1, 1, 'safs', 'sf', 0, '2026-05-16 06:56:36');

-- business_settings (keep only the latest per tenant)
INSERT INTO `business_settings` (`id`, `tenant_id`, `business_name`, `start_date`, `currency`, `currency_symbol`, `time_zone`, `tax_number`, `tax_name`, `financial_year_start`, `stock_expiry_setting`, `created_at`, `updated_at`) VALUES
(9,  1, 'Main Business',      NULL, 'TZS', 'TSh', 'UTC', NULL, NULL, 'January', 'keep_stock', '2026-05-16 06:50:22', '2026-05-16 06:50:22'),
(10, 2, 'Test Business Admin',NULL, 'TZS', 'TSh', 'UTC', NULL, NULL, 'January', 'keep_stock', '2026-05-16 06:50:22', '2026-05-16 06:50:22');

-- ============================================================
-- PRODUCTS  (all products for both tenants)
-- ============================================================
INSERT INTO `products` (`id`, `tenant_id`, `product_type`, `name`, `sku`, `barcode_type`, `unit_id`, `brand_id`, `category_id`, `tax_id`, `type`, `enable_stock`, `alert_quantity`, `purchase_price`, `selling_price`, `image`, `description`, `created_at`) VALUES
-- Tenant 2 products
(10,  2, 'Protector', 'iPhone 15',                         'IP15',              NULL, NULL, NULL, NULL, NULL, 'single', 1, NULL,   800.0000, 1200.0000, NULL, NULL, '2026-05-16 06:50:22'),
(11,  2, 'Protector', 'Samsung S24',                       'S24',               NULL, NULL, NULL, NULL, NULL, 'single', 1, NULL,   700.0000, 1100.0000, NULL, NULL, '2026-05-16 06:50:22'),
(12,  2, 'Protector', 'MacBook Air',                       'MBA',               NULL, NULL, NULL, NULL, NULL, 'single', 1, NULL,   900.0000, 1500.0000, NULL, NULL, '2026-05-16 06:50:22'),
-- Tenant 1 products
(101, 1, 'Protector', 'Protector - 21D - iPhone 11',       'PRO-21D-IP11',      NULL, NULL, 101,  101,  NULL, 'single', 1, 50.0000, 0.0000, 0.0000, NULL, NULL, '2026-05-16 06:50:22'),
(102, 1, 'Protector', 'Protector - Privacy - iPhone 12',   'PRO-PRIV-IP12',     NULL, NULL, 102,  102,  NULL, 'single', 1, 50.0000, 0.0000, 0.0000, NULL, NULL, '2026-05-16 06:50:22'),
(103, 1, 'Cover',     'Cover - Clear - Samsung S21',       'COV-CLR-S21',       NULL, NULL, 104,  105,  NULL, 'single', 1, 50.0000, 0.0000, 0.0000, NULL, NULL, '2026-05-16 06:50:22'),
-- Tenant 2 products (with tenant prefix SKUs)
(201, 2, 'Protector', 'Protector - 21D - iPhone 11',       'PRO-21D-IP11-T2',   NULL, NULL, 201,  201,  NULL, 'single', 1, 50.0000, 0.0000, 0.0000, NULL, NULL, '2026-05-16 06:50:22'),
(202, 2, 'Protector', 'Protector - Privacy - iPhone 12',   'PRO-PRIV-IP12-T2',  NULL, NULL, 202,  202,  NULL, 'single', 1, 50.0000, 0.0000, 0.0000, NULL, NULL, '2026-05-16 06:50:22'),
(203, 2, 'Cover',     'Cover - Clear - Samsung S21',       'COV-CLR-S21-T2',    NULL, NULL, 204,  203,  NULL, 'single', 1, 50.0000, 0.0000, 0.0000, NULL, NULL, '2026-05-16 06:50:22'),
-- Tenant 1 additional product
(204, 1, 'Protector', 'Protector - 31e - tecno spark',     'SKU-1779091261',    NULL, NULL, 207,  207,  NULL, 'single', 1, 50.0000, 0.0000, 0.0000, NULL, NULL, '2026-05-18 08:01:01');

-- ============================================================
-- PRODUCT LOCATIONS  (only for products that actually exist)
-- ============================================================
INSERT INTO `product_locations` (`id`, `product_id`, `location_id`, `qty_available`, `selling_price`, `is_active`, `created_at`) VALUES
-- Tenant 2 products at their locations
(2,  10,  3, 50.0000,  NULL, 1, '2026-05-08 16:28:54'),
(3,  11,  3, 30.0000,  NULL, 1, '2026-05-08 16:28:54'),
(4,  12,  3,  5.0000,  NULL, 1, '2026-05-08 16:28:54'),
-- Tenant 1 products at store 1 and store 2
(54, 101, 1, 97.0000,  NULL, 1, '2026-05-16 06:50:22'),
(55, 101, 2, 50.0000,  NULL, 1, '2026-05-16 06:50:22'),
(56, 102, 1, 48.0000,  NULL, 1, '2026-05-16 06:50:22'),
(57, 103, 1, 29.0000,  NULL, 1, '2026-05-16 06:50:22'),
-- Tenant 2 products at their locations
(58, 201, 3, 100.0000, NULL, 1, '2026-05-16 06:50:22'),
(59, 201, 4,  50.0000, NULL, 1, '2026-05-16 06:50:22'),
(60, 202, 3,  30.0000, NULL, 1, '2026-05-16 06:50:22'),
(61, 203, 3,  20.0000, NULL, 1, '2026-05-16 06:50:22'),
-- Tenant 1 additional product
(67, 204, 2,   0.0000, NULL, 1, '2026-05-18 08:01:01');

-- cash_registers (user_id=6 now exists)
INSERT INTO `cash_registers` (`id`, `tenant_id`, `business_location_id`, `user_id`, `opening_amount`, `status`, `closed_at`, `closing_amount`, `total_card_slips`, `total_cheques`, `closing_note`, `created_at`) VALUES
(6, 1, 2, 6, 0.00, 'open', NULL, 0.00, 0, 0, NULL, '2026-05-09 13:18:01'),
(7, 1, 1, 4, 0.00, 'open', NULL, 0.00, 0, 0, NULL, '2026-05-12 08:15:45');

-- contacts (created_by=7 now exists)
INSERT INTO `contacts` (`id`, `tenant_id`, `type`, `name`, `business_name`, `email`, `mobile`, `tax_number`, `opening_balance`, `address`, `city`, `state`, `country`, `zip_code`, `created_by`, `created_at`) VALUES
(1, 1, 'customer', 'juma', '', '', '07948098', '', 0.00, '', '', '', '', '', 7, '2026-05-12 11:58:33');

-- expense_categories
INSERT INTO `expense_categories` (`id`, `tenant_id`, `name`, `code`) VALUES
(10, 2, 'Rent',      NULL),
(11, 2, 'Utilities', NULL),
(20, 1, 'Packaging', NULL);

-- expenses
INSERT INTO `expenses` (`id`, `tenant_id`, `business_location_id`, `expense_category_id`, `ref_no`, `transaction_date`, `final_total`, `expense_for`, `additional_notes`, `created_by`, `created_at`) VALUES
(21, 1, 1, 20, 'EXP-MAIN-001', '2026-05-16 09:50:22', 50.00, NULL, NULL, 2, '2026-05-16 06:50:22'),
(22, 2, 3, 10, 'EXP-TEST-001', '2026-05-11 09:50:22', 500.00, NULL, NULL, 4, '2026-05-16 06:50:22'),
(23, 2, 3, 11, 'EXP-TEST-002', '2026-05-11 09:50:22', 200.00, NULL, NULL, 4, '2026-05-16 06:50:22');

-- purchases
INSERT INTO `purchases` (`id`, `tenant_id`, `business_location_id`, `supplier_id`, `ref_no`, `purchase_date`, `status`, `payment_status`, `total_before_tax`, `tax_amount`, `discount_amount`, `final_total`, `created_by`, `created_at`) VALUES
(1,  2, 3, NULL, 'PUR-TEST-001', '2026-05-06 19:28:54', 'received', 'paid', 0.00, 0.00, 0.00, 8000.00, 4, '2026-05-08 16:28:54'),
(16, 1, 1, NULL, 'PUR-MAIN-001', '2026-05-16 09:50:22', 'received', 'paid', 0.00, 0.00, 0.00, 100.00,  2, '2026-05-16 06:50:22'),
(17, 2, 3, NULL, 'PUR-TEST-001', '2026-05-14 09:50:22', 'received', 'paid', 0.00, 0.00, 0.00, 8000.00, 4, '2026-05-16 06:50:22'),
(18, 1, 2, NULL, '',             '2026-05-16 07:25:36', 'received', 'paid', 0.00, 0.00, 0.00, 0.00,    4, '2026-05-16 07:25:36'),
(19, 1, 1, NULL, '',             '2026-05-16 08:55:31', 'received', 'paid', 0.00, 0.00, 0.00, 0.00,    2, '2026-05-16 08:55:31');

-- purchase_items
INSERT INTO `purchase_items` (`id`, `purchase_id`, `product_id`, `quantity`, `purchase_price`, `line_total`) VALUES
(1,  1,  10,  10.00, 800.00, 8000.00),
(17, 17, 10,  10.00, 800.00, 8000.00),
(19, 19, 103, 10.00, 0.00,   0.00),
(20, 19, 102, 10.00, 0.00,   0.00),
(21, 19, 102,  1.00, 0.00,   0.00),
(22, 19, 102, 10.00, 0.00,   0.00);

-- sales
INSERT INTO `sales` (`id`, `tenant_id`, `business_location_id`, `customer_id`, `invoice_no`, `transaction_date`, `due_date`, `status`, `payment_status`, `total_before_tax`, `tax_amount`, `discount_amount`, `final_total`, `created_by`, `created_at`) VALUES
(1,  2, 3, NULL, 'INV-TEST-001', '2026-05-07 19:28:54', NULL, 'final', 'paid', 0.00, 0.00, 0.00, 2400.00, 4, '2026-05-08 16:28:54'),
(16, 1, 1, NULL, 'INV-MAIN-001', '2026-05-16 09:50:22', NULL, 'final', 'paid', 0.00, 0.00, 0.00, 250.00,  2, '2026-05-16 06:50:22'),
(17, 2, 3, NULL, 'INV-TEST-001', '2026-05-15 09:50:22', NULL, 'final', 'paid', 0.00, 0.00, 0.00, 2400.00, 4, '2026-05-16 06:50:22');

-- sale_items
INSERT INTO `sale_items` (`id`, `sale_id`, `product_id`, `quantity`, `unit_price`, `line_total`) VALUES
(1,  1,  10, 2.0000, 1200.00, 2400.00),
(16, 16, 10, 2.0000, 1200.00, 2400.00),
(17, 17, 10, 2.0000, 1200.00, 2400.00);

-- transaction_payments
INSERT INTO `transaction_payments` (`id`, `tenant_id`, `sale_id`, `amount`, `method`, `transaction_no`, `note`, `paid_on`, `created_by`) VALUES
(1,  2, 1,  2400.00, 'cash', NULL, NULL, '2026-05-08 19:28:54', 4),
(16, 1, 16,  250.00, 'cash', NULL, NULL, '2026-05-16 09:50:22', 2),
(17, 2, 17, 2400.00, 'cash', NULL, NULL, '2026-05-16 09:50:22', 4);

-- stock_adjustments
INSERT INTO `stock_adjustments` (`id`, `tenant_id`, `business_location_id`, `ref_no`, `transaction_date`, `adjustment_type`, `final_total`, `total_amount_recovered`, `additional_notes`, `created_by`, `created_at`) VALUES
(8, 2, 3, 'ADJ-TEST-001', '2026-05-12 09:50:22', 'normal', 900.00, 0.00, NULL, 4, '2026-05-16 06:50:22');

-- stock_adjustment_items
INSERT INTO `stock_adjustment_items` (`id`, `stock_adjustment_id`, `product_id`, `quantity`, `unit_price`) VALUES
(8, 8, 12, 1.00, 900.00);

-- stock_transfers
INSERT INTO `stock_transfers` (`id`, `tenant_id`, `from_location_id`, `to_location_id`, `ref_no`, `status`, `final_total`, `transaction_date`, `received_at`, `created_by`, `notes`) VALUES
(11, 2, 4, 3, 'TR-TEST-001', 'received', 0.00, '2026-05-13 09:50:22', NULL, 4, 'Initial stock transfer');

-- stock_transfer_items
INSERT INTO `stock_transfer_items` (`id`, `transfer_id`, `product_id`, `quantity`) VALUES
(8, 11, 11, 10.0000);

-- stores
INSERT INTO `stores` (`id`, `location_id`, `name`, `description`, `is_active`, `created_at`) VALUES
(3, 3, 'Main Shop Store', NULL, 1, '2026-05-08 16:33:44'),
(4, 4, 'Warehouse Store', NULL, 1, '2026-05-08 16:33:44');

-- orders
INSERT INTO `orders` (`id`, `tenant_id`, `order_type`, `ref_no`, `placed_by`, `order_from`, `from_shop_id`, `from_store_id`, `to_location_id`, `status`, `payment_status`, `total_amount`, `amount_paid`, `remaining_amount`, `processed_by`, `processed_at`, `notes`, `created_at`, `updated_at`) VALUES
(6,  1, 'BulkOrder',  'ORD-0001', 4, 'juma', NULL, 1, 1, 'accepted', 'paid',    0.00, 6000.00, -6000.00, 4, '2026-05-21 10:51:38', '', '2026-05-21 07:31:38', '2026-05-21 07:51:51'),
(7,  1, 'StoreOrder', 'ORD-0002', 4, '',     NULL, 1, 5, 'pending',  'unpaid',  0.00,    0.00,     0.00, NULL, NULL, '', '2026-05-21 07:32:45', '2026-05-21 07:32:45'),
(8,  1, 'BulkOrder',  'ORD-0003', 4, 'juma', NULL, 1, 1, 'pending',  'unpaid',  0.00,    0.00,     0.00, NULL, NULL, '', '2026-05-21 07:33:08', '2026-05-21 07:33:08'),
(9,  1, 'BulkOrder',  'ORD-0004', 4, 'juma', NULL, 1, 1, 'pending',  'incomplete', 0.00, 0.00,   0.00, NULL, NULL, '', '2026-05-21 07:48:24', '2026-05-21 07:48:24'),
(10, 1, 'StoreOrder', 'ORD-0005', 3, '',     NULL, 1, 5, 'pending',  'unpaid',  0.00,    0.00,     0.00, NULL, NULL, '', '2026-05-21 07:53:03', '2026-05-21 07:53:03'),
(11, 1, 'BulkOrder',  'ORD-0006', 3, 'juma', NULL, 1, 2, 'pending',  'paid',    0.00,    0.00,     0.00, NULL, NULL, '', '2026-05-21 07:53:32', '2026-05-21 07:53:32'),
(12, 1, 'BulkOrder',  'ORD-0007', 3, 'juma', NULL, 1, 2, 'pending',  'unpaid',  0.00,    0.00,     0.00, NULL, NULL, '', '2026-05-21 08:54:39', '2026-05-21 08:54:39'),
(13, 1, 'StoreOrder', 'ORD-0008', 3, '',     NULL, 1, 5, 'pending',  'unpaid',  0.00,    0.00,     0.00, NULL, NULL, '', '2026-05-21 08:55:08', '2026-05-21 08:55:08');

-- order_items
INSERT INTO `order_items` (`id`, `order_id`, `product_id`, `quantity`, `from_shop_qty`, `from_store_qty`, `unit_price`, `subtotal`) VALUES
(1, 6,  101, 1.00, 0.00, 1.00, 0.00, 0.00),
(2, 7,  101, 1.00, 0.00, 1.00, 0.00, 0.00),
(3, 8,  101, 1.00, 0.00, 1.00, 0.00, 0.00),
(4, 9,  204, 1.00, 0.00, 1.00, 0.00, 0.00),
(5, 10, 101, 2.00, 0.00, 2.00, 0.00, 0.00),
(6, 11, 101, 1.00, 0.00, 1.00, 0.00, 0.00),
(7, 12, 101, 1.00, 0.00, 1.00, 0.00, 0.00),
(8, 13, 101, 1.00, 0.00, 1.00, 0.00, 0.00);

-- order_payments
INSERT INTO `order_payments` (`id`, `order_id`, `amount`, `payment_method`, `notes`, `paid_by`, `created_at`) VALUES
(1, 6, 6000.00, 'cash', NULL, 4, '2026-05-21 07:51:51');

-- tenant_modules
INSERT INTO `tenant_modules` (`id`, `tenant_id`, `module_key`, `is_installed`, `installed_at`, `updated_at`) VALUES
(1,  1, 'stock_adjustments', 0, '2026-05-08 17:08:52', '2026-05-09 16:35:56'),
(2,  1, 'sales',             0, '2026-05-08 17:08:53', '2026-05-21 07:12:28'),
(3,  1, 'purchases',         0, '2026-05-08 17:08:56', '2026-05-09 16:35:52'),
(4,  1, 'stock_transfers',   0, '2026-05-08 17:08:57', '2026-05-21 07:12:42'),
(5,  1, 'expenses',          0, '2026-05-08 17:08:59', '2026-05-21 07:13:22'),
(8,  1, 'reports',           0, '2026-05-08 17:09:07', '2026-05-21 07:12:53'),
(12, 1, 'store_management',  1, '2026-05-12 09:26:16', '2026-05-12 09:26:16');

-- ============================================================
-- PRIMARY KEYS & INDEXES
-- ============================================================
ALTER TABLE `brands`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`);

ALTER TABLE `business_locations`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`);

ALTER TABLE `business_settings`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`);

ALTER TABLE `cash_registers`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `business_location_id` (`business_location_id`),
  ADD KEY `user_id` (`user_id`);

ALTER TABLE `categories`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `parent_id` (`parent_id`);

ALTER TABLE `contacts`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`);

ALTER TABLE `expenses`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `business_location_id` (`business_location_id`),
  ADD KEY `expense_category_id` (`expense_category_id`);

ALTER TABLE `expense_categories`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`);

ALTER TABLE `notifications`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_id` (`user_id`);

ALTER TABLE `orders`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `placed_by` (`placed_by`),
  ADD KEY `processed_by` (`processed_by`),
  ADD KEY `to_location_id` (`to_location_id`);

ALTER TABLE `order_items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `order_id` (`order_id`),
  ADD KEY `product_id` (`product_id`);

ALTER TABLE `order_payments`
  ADD PRIMARY KEY (`id`),
  ADD KEY `order_id` (`order_id`),
  ADD KEY `paid_by` (`paid_by`);

ALTER TABLE `products`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`);

ALTER TABLE `product_locations`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `product_id` (`product_id`,`location_id`),
  ADD KEY `location_id` (`location_id`);

ALTER TABLE `purchases`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `business_location_id` (`business_location_id`);

ALTER TABLE `purchase_items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `purchase_id` (`purchase_id`),
  ADD KEY `product_id` (`product_id`);

ALTER TABLE `sales`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`);

ALTER TABLE `sale_items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `sale_id` (`sale_id`),
  ADD KEY `product_id` (`product_id`);

ALTER TABLE `stock_adjustments`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `business_location_id` (`business_location_id`);

ALTER TABLE `stock_adjustment_items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `stock_adjustment_id` (`stock_adjustment_id`),
  ADD KEY `product_id` (`product_id`);

ALTER TABLE `stock_transfers`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `from_store_id` (`from_location_id`),
  ADD KEY `to_store_id` (`to_location_id`),
  ADD KEY `created_by` (`created_by`);

ALTER TABLE `stock_transfer_items`
  ADD PRIMARY KEY (`id`),
  ADD KEY `transfer_id` (`transfer_id`),
  ADD KEY `product_id` (`product_id`);

ALTER TABLE `stores`
  ADD PRIMARY KEY (`id`),
  ADD KEY `location_id` (`location_id`);

ALTER TABLE `store_locations`
  ADD PRIMARY KEY (`id`);

ALTER TABLE `tenants`
  ADD PRIMARY KEY (`id`);

ALTER TABLE `tenant_modules`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `tenant_id` (`tenant_id`,`module_key`);

ALTER TABLE `transaction_payments`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `sale_id` (`sale_id`);

ALTER TABLE `units`
  ADD PRIMARY KEY (`id`),
  ADD KEY `tenant_id` (`tenant_id`);

ALTER TABLE `unit_conversions`
  ADD PRIMARY KEY (`id`),
  ADD KEY `unit_id` (`unit_id`),
  ADD KEY `base_unit_id` (`base_unit_id`);

ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `email` (`email`),
  ADD KEY `tenant_id` (`tenant_id`),
  ADD KEY `location_id` (`location_id`);

-- ============================================================
-- AUTO_INCREMENT VALUES
-- ============================================================
ALTER TABLE `brands`             MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=208;
ALTER TABLE `business_locations` MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=7;
ALTER TABLE `business_settings`  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=11;
ALTER TABLE `cash_registers`     MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=8;
ALTER TABLE `categories`         MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=208;
ALTER TABLE `contacts`           MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;
ALTER TABLE `expenses`           MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=24;
ALTER TABLE `expense_categories` MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=21;
ALTER TABLE `notifications`      MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `orders`             MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=14;
ALTER TABLE `order_items`        MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;
ALTER TABLE `order_payments`     MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;
ALTER TABLE `products`           MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=205;
ALTER TABLE `product_locations`  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=68;
ALTER TABLE `purchases`          MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=20;
ALTER TABLE `purchase_items`     MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=23;
ALTER TABLE `sales`              MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;
ALTER TABLE `sale_items`         MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;
ALTER TABLE `stock_adjustments`  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;
ALTER TABLE `stock_adjustment_items` MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;
ALTER TABLE `stock_transfers`    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=12;
ALTER TABLE `stock_transfer_items` MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;
ALTER TABLE `stores`             MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=5;
ALTER TABLE `store_locations`    MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `tenants`            MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;
ALTER TABLE `tenant_modules`     MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=22;
ALTER TABLE `transaction_payments` MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;
ALTER TABLE `units`              MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;
ALTER TABLE `unit_conversions`   MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `users`              MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=8;

-- ============================================================
-- FOREIGN KEY CONSTRAINTS
-- ============================================================
ALTER TABLE `brands`
  ADD CONSTRAINT `brands_ibfk_1` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE;

ALTER TABLE `business_locations`
  ADD CONSTRAINT `business_locations_ibfk_1` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE;

ALTER TABLE `business_settings`
  ADD CONSTRAINT `business_settings_ibfk_1` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE;

-- ============================================================
-- RE-ENABLE FOREIGN KEY CHECKS
-- ============================================================
SET FOREIGN_KEY_CHECKS = 1;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
