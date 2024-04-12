-- +goose Up
-- +goose StatementBegin

SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_definitions ADD trx_hash_hex VARCHAR(100) null;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_definitions drop column trx_hash_hex;

-- +goose StatementEnd
