-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS deposit_orders (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    min_amount DECIMAL(30, 10) NOT NULL,
    max_amount DECIMAL(30, 10) NOT NULL,
    wallet_ids BIGINT[] NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    project_id INT REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS deposits (
    id SERIAL PRIMARY KEY,
    deposit_order_id INT REFERENCES deposit_orders(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL,
    wallet_id BIGINT REFERENCES wallets(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    transaction_id VARCHAR(255) NOT NULL,
    amount           numeric(42, 22)              NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS deposit_orders;
DROP TABLE IF EXISTS deposits;
-- +goose StatementEnd
