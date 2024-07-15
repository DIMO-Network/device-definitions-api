-- +goose Up
-- +goose StatementBegin

SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_definitions_api.device_definitions
    ALTER COLUMN trx_hash_hex TYPE varchar(100)[] USING string_to_array(trx_hash_hex, ',');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SELECT 'down SQL query';

-- +goose StatementEnd
