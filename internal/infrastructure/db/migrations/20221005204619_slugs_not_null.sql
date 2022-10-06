-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;

alter table device_makes alter column name_slug set not null;
alter table device_definitions alter column model_slug set not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
alter table device_makes alter column name_slug drop not null;
alter table device_definitions alter column model_slug drop not null;
-- +goose StatementEnd
