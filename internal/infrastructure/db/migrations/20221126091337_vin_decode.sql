-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;
ALTER TABLE device_makes ADD wmis character(3) array NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE device_makes DROP column wmis;

-- +goose StatementEnd
