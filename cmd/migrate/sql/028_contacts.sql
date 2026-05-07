CREATE TABLE IF NOT EXISTS contacts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT NOT NULL,
    type ENUM('supplier', 'customer', 'both') NOT NULL,
    name VARCHAR(255) NOT NULL,
    business_name VARCHAR(255) NULL,
    email VARCHAR(255) NULL,
    mobile VARCHAR(20) NOT NULL,
    tax_number VARCHAR(50) NULL,
    opening_balance DECIMAL(15,2) DEFAULT 0.00,
    address TEXT NULL,
    city VARCHAR(100) NULL,
    state VARCHAR(100) NULL,
    country VARCHAR(100) NULL,
    zip_code VARCHAR(20) NULL,
    created_by INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
