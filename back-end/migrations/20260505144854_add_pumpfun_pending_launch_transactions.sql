-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pumpfun_pending_launch_transactions (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    transaction bytea NOT NULL,
    signer TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    valid_until TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pumpfun_pending_launch_transactions;
-- +goose StatementEnd
