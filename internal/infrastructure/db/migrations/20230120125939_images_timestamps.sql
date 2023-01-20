-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO device_definitions_api,public;

alter table images add column created_at timestamptz NOT NULL DEFAULT current_timestamp;
alter table images add column updated_at timestamptz NOT NULL DEFAULT current_timestamp;

ALTER TABLE images ADD CONSTRAINT images_ddid_url_key UNIQUE (device_definition_id, source_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO device_definitions_api,public;

alter table images drop column created_at;
alter table images drop column updated_at;
alter table images drop constraint images_ddid_url_key;
-- +goose StatementEnd
