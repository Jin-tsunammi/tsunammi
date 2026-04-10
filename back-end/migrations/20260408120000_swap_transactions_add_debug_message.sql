-- +goose Up
-- +goose StatementBegin
ALTER TABLE swap_transactions
    ADD COLUMN IF NOT EXISTS debug_message TEXT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE swap_transactions
    DROP COLUMN IF EXISTS debug_message;
-- +goose StatementEnd
