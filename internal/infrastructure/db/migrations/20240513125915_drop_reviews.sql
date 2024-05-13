-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
drop table reviews;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
create table reviews
(
    device_definition_id char(27)                                           not null
        constraint device_definition_id_fkey
            references device_definitions,
    url                  varchar                                            not null,
    image_url            varchar                                            not null,
    channel              varchar,
    approved             boolean                                            not null,
    created_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    id                   char(27)                                           not null
        constraint review_oid
            primary key,
    comments             text                                               not null,
    approved_by          text                                               not null,
    position             integer                                            not null
);
grant select on reviews to readonly;
-- +goose StatementEnd
