-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_styles ADD metadata jsonb null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_styles drop column year;
-- +goose StatementEnd
