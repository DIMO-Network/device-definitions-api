-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table device_definitions
    drop constraint device_definitions_pkey cascade;

alter table device_definitions
    add primary key (name_slug);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
