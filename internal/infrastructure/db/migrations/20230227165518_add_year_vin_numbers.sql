-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE vin_numbers ADD year int not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE vin_numbers drop column year;
-- +goose StatementEnd
