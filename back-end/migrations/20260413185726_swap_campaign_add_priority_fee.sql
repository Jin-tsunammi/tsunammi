-- +goose Up
-- +goose StatementBegin
ALTER TABLE swap_campaigns
ADD COLUMN priority_fee DOUBLE PRECISION NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE swap_campaigns
DROP COLUMN IF EXISTS priority_fee;
-- +goose StatementEnd
