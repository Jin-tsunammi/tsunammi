-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_actions(
    id         SERIAL PRIMARY KEY,
    action     VARCHAR(128) NOT NULL,
    value      VARCHAR(255) NOT NULL,
    user_id    INT REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_actions;
-- +goose StatementEnd
