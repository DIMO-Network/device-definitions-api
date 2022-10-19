-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

SET search_path = device_definitions_api, public;


CREATE TABLE IF NOT EXISTS device_types
(
    id character(50) PRIMARY KEY NOT NULL, -- not use ksuid. Use slug id
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             name text not null,
                             properties jsonb
                             );

alter table device_definitions
    add column device_type_id character(50) null

alter table device_definitions
    add constraint fk_device_types
        foreign key (device_type_id) references device_types
            on delete cascade;
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

ALTER TABLE device_definitions
    DROP COLUMN device_type_id;