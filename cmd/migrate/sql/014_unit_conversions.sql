CREATE TABLE IF NOT EXISTS unit_conversions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    unit_id INT NOT NULL,
    base_unit_id INT NOT NULL,
    operator ENUM('multiply', 'divide') DEFAULT 'multiply',
    operation_value DECIMAL(15,4) NOT NULL,
    FOREIGN KEY (unit_id) REFERENCES units(id) ON DELETE CASCADE,
    FOREIGN KEY (base_unit_id) REFERENCES units(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
