CREATE TABLE IF NOT EXISTS transaction_payments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT NOT NULL,
    sale_id INT NOT NULL,
    amount DECIMAL(15,2) DEFAULT 0.00,
    method ENUM('cash', 'card', 'cheque', 'bank_transfer', 'other', 'custom_pay_1', 'custom_pay_2', 'custom_pay_3') DEFAULT 'cash',
    transaction_no VARCHAR(100) NULL,
    note TEXT NULL,
    paid_on DATETIME NOT NULL,
    created_by INT NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
