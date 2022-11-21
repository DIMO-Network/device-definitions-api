-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path to device_definitions_api, public;

ALTER TABLE device_makes ADD metadata jsonb NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path to device_definitions_api, public;

ALTER TABLE device_makes DROP COLUMN metadata;
-- +goose StatementEnd
