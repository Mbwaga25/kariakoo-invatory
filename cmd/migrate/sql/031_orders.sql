-- Orders table: Supports both Store Orders and Bulk Orders
CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT NOT NULL,
    order_type ENUM('StoreOrder', 'BulkOrder') NOT NULL DEFAULT 'StoreOrder',
    ref_no VARCHAR(50) NOT NULL,
    
    -- Who placed the order (Shop Keeper)
    placed_by INT NOT NULL,
    
    -- For Bulk Orders: which shop/store the order is from
    order_from VARCHAR(255) DEFAULT NULL,
    from_shop_id INT DEFAULT NULL,
    from_store_id INT DEFAULT NULL,
    
    -- Destination location
    to_location_id INT NOT NULL,
    
    -- Order status flow
    status ENUM('pending', 'accepted', 'rejected', 'completed') NOT NULL DEFAULT 'pending',
    
    -- Payment tracking
    payment_status ENUM('paid', 'unpaid', 'incomplete') NOT NULL DEFAULT 'unpaid',
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    amount_paid DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    remaining_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    
    -- Store Keeper who processed the order
    processed_by INT DEFAULT NULL,
    processed_at DATETIME DEFAULT NULL,
    
    notes TEXT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (placed_by) REFERENCES users(id),
    FOREIGN KEY (processed_by) REFERENCES users(id),
    FOREIGN KEY (to_location_id) REFERENCES business_locations(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Order items
CREATE TABLE IF NOT EXISTS order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    unit_price DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Order payment history (for tracking partial payments)
CREATE TABLE IF NOT EXISTS order_payments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    payment_method VARCHAR(50) DEFAULT 'cash',
    notes TEXT DEFAULT NULL,
    paid_by INT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (paid_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
