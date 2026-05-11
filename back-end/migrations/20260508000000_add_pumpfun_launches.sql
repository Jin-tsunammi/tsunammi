-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pumpfun_launches (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL,
    create_tx BYTEA NOT NULL,
    buy_tx BYTEA,
    wallet_buy_txs TEXT[] NOT NULL DEFAULT '{}',
    mint_pubkey TEXT NOT NULL,
    signer TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    CONSTRAINT pumpfun_launches_status_check CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pumpfun_launches;
-- +goose StatementEnd
