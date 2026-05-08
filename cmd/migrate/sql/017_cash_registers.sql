CREATE TABLE IF NOT EXISTS cash_registers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT NOT NULL,
    business_location_id INT NOT NULL,
    user_id INT NOT NULL,
    opening_amount DECIMAL(15,2) DEFAULT 0.00,
    status ENUM('open', 'close') DEFAULT 'open',
    closed_at DATETIME NULL,
    closing_amount DECIMAL(15,2) DEFAULT 0.00,
    total_card_slips INT DEFAULT 0,
    total_cheques INT DEFAULT 0,
    closing_note TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (business_location_id) REFERENCES business_locations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
