CREATE TABLE IF NOT EXISTS tenant_modules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT NOT NULL,
    module_key VARCHAR(50) NOT NULL,
    is_installed BOOLEAN DEFAULT TRUE,
    installed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY (tenant_id, module_key),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Pre-install all modules for existing tenants
INSERT IGNORE INTO tenant_modules (tenant_id, module_key)
SELECT t.id, m.key FROM tenants t
CROSS JOIN (
    SELECT 'sales' as `key` UNION
    SELECT 'purchases' UNION
    SELECT 'stock_adjustments' UNION
    SELECT 'stock_transfers' UNION
    SELECT 'expenses' UNION
    SELECT 'reports'
) m;
