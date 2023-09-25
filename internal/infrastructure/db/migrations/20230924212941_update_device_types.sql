-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;

alter table device_types
    alter column id type varchar(50) using id::varchar(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
alter table device_types
    alter column id type char(50) using id::char(50);
-- +goose StatementEnd
