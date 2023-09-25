-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;

alter table device_definitions
    alter column device_type_id type varchar(50) using device_type_id::varchar(50);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
alter table device_definitions
    alter column device_type_id type char(50) using device_type_id::char(50);
-- +goose StatementEnd
