-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS providers
(
    id   INT PRIMARY KEY,
    name VARCHAR(50) NOT NULL DEFAULT ''
);
INSERT INTO providers
VALUES (1, 'Raydium'),
       (2, 'Pumpfun');
ALTER TABLE swap_campaigns
    ADD COLUMN IF NOT EXISTS provider_id INT NOT NULL DEFAULT 1 REFERENCES providers(id);
ALTER TABLE swap_campaigns
  DROP COLUMN IF EXISTS provider_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE swap_campaigns
  DROP COLUMN IF EXISTS provider_id;
DROP TABLE IF EXISTS providers;
-- +goose StatementEnd
