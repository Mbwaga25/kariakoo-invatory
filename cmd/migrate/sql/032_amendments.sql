-- Amendment: Add product_type column to products
ALTER TABLE products ADD COLUMN IF NOT EXISTS product_type ENUM('Protector', 'Cover') DEFAULT 'Protector' AFTER tenant_id;

-- Amendment: Add from_shop_qty and from_store_qty to order_items for bulk order split
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS from_shop_qty DECIMAL(15,2) NOT NULL DEFAULT 0.00 AFTER quantity;
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS from_store_qty DECIMAL(15,2) NOT NULL DEFAULT 0.00 AFTER from_shop_qty;

-- Amendment: Make supplier_id nullable in purchases (we're removing suppliers)
ALTER TABLE purchases MODIFY COLUMN supplier_id INT DEFAULT NULL;
