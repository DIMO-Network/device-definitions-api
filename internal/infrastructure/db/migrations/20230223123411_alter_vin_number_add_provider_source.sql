-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;

ALTER TABLE vin_numbers ADD decode_provider text COLLATE pg_catalog."default";

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE vin_numbers DROP COLUMN decode_provider;

-- +goose StatementEnd
