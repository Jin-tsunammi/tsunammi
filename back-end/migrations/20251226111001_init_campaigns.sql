-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS swap_campaign_types
(
    id   BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL
);

INSERT INTO swap_campaign_types (id, name)
VALUES (1, 'PULL UP'),
       (2, 'PULL DOWN');

CREATE TABLE IF NOT EXISTS swap_campaigns(

    id                            UUID PRIMARY KEY,
    type_id                       BIGINT                   NOT NULL,
    user_id                       BIGINT                   NOT NULL,
    project_id                    BIGINT                   NOT NULL,
    pool_id                       VARCHAR(128)             NOT NULL,
    budget                        DOUBLE PRECISION         NOT NULL,
    slippage                      BIGINT                   NOT NULL,
    started_price                 NUMERIC(42, 22)          NOT NULL,
    goal_price                    NUMERIC(42, 22)          NOT NULL,
    status                        VARCHAR(128)             NOT NULL,
    parallel_transactions_amount  INTEGER                  NOT NULL,
    min_transactions_budget       DOUBLE PRECISION         NOT NULL,
    max_transactions_budget       DOUBLE PRECISION         NOT NULL,
    min_time_between_transactions BIGINT                   NOT NULL,
    max_time_between_transactions BIGINT                   NOT NULL,
    transaction_speed             VARCHAR(128)             NOT NULL,
    token_mint_from               VARCHAR(128)             NOT NULL,
    token_mint_to                 VARCHAR(128)             NOT NULL,
    created_at                    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at                    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    goal_bps_change               BIGINT                    NOT NULL,
    using_jito                    BOOL        DEFAULT true  NOT NULL,

    CONSTRAINT fk_campaign_type
        FOREIGN KEY (type_id)
            REFERENCES swap_campaign_types (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_users_reference
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS swap_transactions(
    id                BIGSERIAL PRIMARY KEY,
    campaign_id       UUID            NOT NULL,
    transaction_hash  VARCHAR(256)    NOT NULL,
    pool_id           VARCHAR(256)    NOT NULL,
    token_mint_from   VARCHAR(256)    NOT NULL,
    token_mint_to     VARCHAR(256)    NOT NULL,
    address_from      VARCHAR(256)    NOT NULL,
    address_to        VARCHAR(256)    NOT NULL,
    amount_token_from NUMERIC(42, 22) NOT NULL,
    amount_token_to   NUMERIC(42, 22) NOT NULL,
    status            VARCHAR(256)    NOT NULL,
    message           VARCHAR(512)    NOT NULL,
    created_at        TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    CONSTRAINT fk_swap_transaction_campaign
        FOREIGN KEY (campaign_id)
        REFERENCES swap_campaigns(id)
        ON DELETE CASCADE
);


CREATE INDEX IF NOT EXISTS idx_swap_transactions_campaign_id ON swap_transactions (campaign_id);
CREATE INDEX IF NOT EXISTS idx_swap_transactions_hash ON swap_transactions (transaction_hash);

CREATE INDEX IF NOT EXISTS idx_swap_campaigns_project_id ON swap_campaigns (project_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS swap_campaigns;
DROP TABLE IF EXISTS swap_campaign_types;
DROP TABLE IF EXISTS swap_transactions;

DROP INDEX IF EXISTS idx_swap_campaigns_project_id;
DROP INDEX IF EXISTS idx_swap_transactions_campaign_id;
DROP INDEX IF EXISTS idx_swap_transactions_hash;
-- +goose StatementEnd
