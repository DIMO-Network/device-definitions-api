-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

create table wmis
(
    wmi char(3) not null primary key,
    device_make_id char(27) not null,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp,
    CONSTRAINT fk_device_make_id FOREIGN KEY (device_make_id)
        REFERENCES device_makes (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    CONSTRAINT idx_wmi_device_make_id UNIQUE (wmi, device_make_id)
);

comment on table wmis is 'world manufacturer identifier';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

DROP TABLE wmis;

-- +goose StatementEnd
