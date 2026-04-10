-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects
DROP CONSTRAINT IF EXISTS projects_name_key;

ALTER TABLE projects
ADD CONSTRAINT projects_name_user_id_key UNIQUE(name, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects
DROP CONSTRAINT IF EXISTS projects_name_user_id_key;

ALTER TABLE projects
ADD CONSTRAINT projects_name_key UNIQUE(name);
-- +goose StatementEnd
