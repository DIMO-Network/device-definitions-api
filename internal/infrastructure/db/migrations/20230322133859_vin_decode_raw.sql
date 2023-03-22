-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE vin_numbers ADD vincario_data jsonb;
ALTER TABLE vin_numbers ADD drivly_data jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE vin_numbers drop vincario_data;
ALTER TABLE vin_numbers drop drivly_data;
-- +goose StatementEnd
