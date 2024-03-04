-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE vin_numbers ADD datgroup_data jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE vin_numbers drop datgroup_data;
-- +goose StatementEnd
