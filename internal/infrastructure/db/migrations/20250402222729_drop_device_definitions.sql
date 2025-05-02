-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

drop table device_definitions cascade;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

create table device_definitions
(
    id                   char(27)                                           not null,
    model                varchar(100)                                       not null,
    year                 smallint                                           not null,
    metadata             jsonb,
    created_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at           timestamp with time zone default CURRENT_TIMESTAMP not null,
    source               text,
    verified             boolean                  default false             not null,
    external_id          text,
    device_make_id       char(27)                                           not null
        constraint fk_device_make_id
            references device_makes
            on update cascade on delete restrict,
    model_slug           varchar(100)                                       not null,
    device_type_id       varchar(50)
        constraint fk_device_types
            references device_types
            on delete cascade,
    external_ids         jsonb,
    hardware_template_id varchar(500),
    trx_hash_hex         varchar(100)[],
    name_slug            varchar(100)                                       not null
        primary key
        constraint device_definitions_name_slug_uniq
            unique,
    constraint idx_device_make_id_model_year
        unique (device_make_id, model, year)
);
-- +goose StatementEnd
