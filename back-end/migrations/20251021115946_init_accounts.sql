-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS exchanges (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE
    );

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exchange_id INT NOT NULL REFERENCES exchanges(id) ON DELETE CASCADE,
    exchange_account_id BIGINT NOT NULL,
    api_name VARCHAR(255) NOT NULL,
    status VARCHAR(128) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    withdraw_limit integer
    );

INSERT INTO exchanges (name) VALUES ('KUCOIN') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS exchanges;
-- +goose StatementEnd
