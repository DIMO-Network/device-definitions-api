-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

SET search_path = device_definitions_api, public;

CREATE TABLE reviews (
    device_definition_id char(27) NOT NULL,
    "url" varchar NOT NULL,
    approved boolean NOT NULL
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

ALTER TABLE reviews ADD CONSTRAINT device_definition_id_pkey PRIMARY KEY (device_definition_id);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;
DROP TABLE reviews;
-- +goose StatementEnd