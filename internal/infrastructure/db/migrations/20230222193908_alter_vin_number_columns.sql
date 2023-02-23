-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;

ALTER TABLE device_definitions_api.vin_numbers
ALTER COLUMN Vis TYPE char(8);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

-- +goose StatementEnd
