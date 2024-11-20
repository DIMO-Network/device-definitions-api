-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
alter table vin_numbers drop column device_definition_id;
alter table vin_numbers alter column definition_id set not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
alter table vin_numbers add column device_definition_id text not null default '';
-- +goose StatementEnd
