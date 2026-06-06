-- +goose Up
CREATE TABLE IF NOT EXISTS savings_balance_details (
    id INT AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    branch_code VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    account_no VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    cif_no VARCHAR(255) NOT NULL,
    acc_alt_no VARCHAR(255) NULL,
    interest_rate DECIMAL(10, 2) NOT NULL,
    account_status VARCHAR(255) NOT NULL,
    account_registered_date DATE NULL,
    currency VARCHAR(255) NOT NULL,
    debit_balance DECIMAL(20, 2) NOT NULL,
    credit_balance DECIMAL(20, 2) NOT NULL,
    accrue_interest_debit DECIMAL(20, 2) NOT NULL,
    accrue_interest_credit DECIMAL(20, 2) NOT NULL,
    marketing_id VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY unique_savings_balance (date, account_no)
);

-- +goose Down
DROP TABLE IF EXISTS savings_balance_details;
