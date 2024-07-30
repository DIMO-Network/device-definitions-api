-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
alter table device_definitions
    add constraint device_definitions_name_slug_uniq
        unique (name_slug);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
SELECT 'down SQL query';
-- +goose StatementEnd
