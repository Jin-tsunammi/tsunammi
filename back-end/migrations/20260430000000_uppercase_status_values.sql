-- +goose Up
-- +goose StatementBegin
UPDATE accounts SET status = UPPER(status) WHERE status IN ('active', 'inactive', 'pending', 'deleted');
UPDATE wallets SET status = UPPER(status) WHERE status IN ('success', 'import_pending', 'creation_pending');
UPDATE swap_campaigns SET status = UPPER(status) WHERE status IN ('in_use', 'done', 'budget_done', 'insufficient_funds', 'stop', 'error');
UPDATE swap_transactions SET status = UPPER(status) WHERE status IN ('in_use', 'done', 'budget_done', 'insufficient_funds', 'stop', 'error');
UPDATE swap_campaigns SET status = 'ACTIVE' WHERE status = 'IN_USE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE accounts SET status = LOWER(status) WHERE status IN ('ACTIVE', 'INACTIVE', 'PENDING', 'DELETED');
UPDATE wallets SET status = LOWER(status) WHERE status IN ('SUCCESS', 'IMPORT_PENDING', 'CREATION_PENDING');
UPDATE swap_campaigns SET status = LOWER(status) WHERE status IN ('IN_USE', 'DONE', 'BUDGET_DONE', 'INSUFFICIENT_FUNDS', 'STOP', 'ERROR');
UPDATE swap_transactions SET status = LOWER(status) WHERE status IN ('IN_USE', 'DONE', 'BUDGET_DONE', 'INSUFFICIENT_FUNDS', 'STOP', 'ERROR');
UPDATE swap_campaigns SET status = 'IN_USE' WHERE status = 'ACTIVE';
-- +goose StatementEnd
