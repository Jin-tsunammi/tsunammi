-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) UNIQUE NOT NULL
);

INSERT INTO roles (id, name) VALUES (1, 'DEFAULT');

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255),
  public_key VARCHAR(128),
  role_id INT REFERENCES roles(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE users ADD CONSTRAINT users_email_public_key_unique UNIQUE (email, public_key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_public_key_unique;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
