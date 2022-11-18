-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE reviews
    ADD COLUMN id character(27) COLLATE pg_catalog."default" NOT NULL;

ALTER TABLE reviews DROP CONSTRAINT device_definition_id_pkey;
ALTER TABLE reviews ADD CONSTRAINT review_oid PRIMARY KEY (id);

ALTER TABLE reviews
    ADD COLUMN comments text NOT NULL;

ALTER TABLE reviews
    ADD COLUMN approved_by text NOT NULL;


ALTER TABLE reviews
    ADD COLUMN position int NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE reviews
DROP COLUMN id;

ALTER TABLE reviews
DROP COLUMN comments;

ALTER TABLE reviews
DROP COLUMN approved_by;

ALTER TABLE reviews
DROP COLUMN position;

-- +goose StatementEnd
