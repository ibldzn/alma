-- +goose Up
CREATE TABLE IF NOT EXISTS time_deposit_today (
    id INT AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    branch_code VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    account_no VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    cif_no VARCHAR(255) NOT NULL,
    certificate_no VARCHAR(255) NOT NULL,
    interest_rate DECIMAL(10, 2) NOT NULL,
    start_date DATE NOT NULL,
    maturity_date DATE NOT NULL,
    term INT NOT NULL,
    automatic_rollover BOOLEAN NOT NULL,
    compound_interest BOOLEAN NOT NULL,
    currency VARCHAR(255) NOT NULL,
    nominal DECIMAL(20, 2) NOT NULL,
    interest_accrual DECIMAL(20, 2) NOT NULL,
    marketing_id VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS time_deposit_today;
