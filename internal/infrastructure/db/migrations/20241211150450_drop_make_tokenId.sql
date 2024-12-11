-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
alter table device_makes
    drop constraint device_makes_token_id_key CASCADE;

alter table device_makes drop column token_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
alter table device_makes add column token_id numeric(78);
-- +goose StatementEnd
