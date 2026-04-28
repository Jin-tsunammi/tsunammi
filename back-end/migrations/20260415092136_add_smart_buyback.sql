-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS buyback_campaigns (
  id UUID PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider_id INT NOT NULL REFERENCES providers(id),
  project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  pool_id VARCHAR(128) NOT NULL,
  pool_program_id VARCHAR(128) NOT NULL,
  token_mint VARCHAR(128) NOT NULL,
  status VARCHAR(128) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TYPE buyback_target_type AS ENUM ('BUY', 'SELL');

CREATE TABLE IF NOT EXISTS buyback_targets (
  id UUID PRIMARY KEY,
  campaign_id UUID NOT NULL REFERENCES buyback_campaigns(id) ON DELETE CASCADE,
  type buyback_target_type NOT NULL,
  target_price NUMERIC(42, 22) NOT NULL,
  budget NUMERIC(42, 22) NOT NULL,
  remaining_budget NUMERIC(42, 22) NOT NULL,
  slippage BIGINT NOT NULL,
  parallel_transactions_amount INTEGER NOT NULL,
  min_transactions_amount NUMERIC(42, 22) NOT NULL,
  max_transactions_amount NUMERIC(42, 22) NOT NULL,
  min_time_between_transactions BIGINT NOT NULL,
  max_time_between_transactions BIGINT NOT NULL,
  transaction_speed VARCHAR(128) NOT NULL,
  using_jito BOOL NOT NULL,
  priority_fee NUMERIC(42, 22) NOT NULL,
  status VARCHAR(128) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
  start_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS buyback_transactions (
  id BIGSERIAL PRIMARY KEY,
  campaign_id UUID NOT NULL REFERENCES buyback_campaigns(id) ON DELETE CASCADE,
  target_id UUID NOT NULL REFERENCES buyback_targets(id) ON DELETE CASCADE,
  transaction_hash VARCHAR(256) NOT NULL,
  pool_id VARCHAR(256) NOT NULL,
  token_mint_from VARCHAR(256) NOT NULL,
  token_mint_to VARCHAR(256) NOT NULL,
  address_from VARCHAR(256) NOT NULL,
  address_to VARCHAR(256) NOT NULL,
  amount_token_from NUMERIC(42, 22) NOT NULL DEFAULT 0,
  amount_token_to NUMERIC(42, 22) NOT NULL DEFAULT 0,
  status VARCHAR(256) NOT NULL,
  message VARCHAR(512) NOT NULL,
  debug_message TEXT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_buyback_campaigns_user_id ON buyback_campaigns (user_id);
CREATE INDEX IF NOT EXISTS idx_buyback_campaigns_project_id ON buyback_campaigns (project_id);
CREATE INDEX IF NOT EXISTS idx_buyback_campaigns_status ON buyback_campaigns (status);
CREATE INDEX IF NOT EXISTS idx_buyback_targets_campaign_id ON buyback_targets (campaign_id);
CREATE INDEX IF NOT EXISTS idx_buyback_transactions_campaign_id ON buyback_transactions (campaign_id);
CREATE INDEX IF NOT EXISTS idx_buyback_transactions_target_id ON buyback_transactions (target_id);
CREATE INDEX IF NOT EXISTS idx_buyback_transactions_hash ON buyback_transactions (transaction_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_buyback_transactions_hash;
DROP INDEX IF EXISTS idx_buyback_transactions_target_id;
DROP INDEX IF EXISTS idx_buyback_transactions_campaign_id;
DROP INDEX IF EXISTS idx_buyback_targets_campaign_id;
DROP INDEX IF EXISTS idx_buyback_campaigns_status;
DROP INDEX IF EXISTS idx_buyback_campaigns_project_id;
DROP INDEX IF EXISTS idx_buyback_campaigns_user_id;

DROP TABLE IF EXISTS buyback_transactions;
DROP TABLE IF EXISTS buyback_targets;
DROP TABLE IF EXISTS buyback_campaigns;
DROP TYPE IF EXISTS buyback_target_type;
-- +goose StatementEnd
