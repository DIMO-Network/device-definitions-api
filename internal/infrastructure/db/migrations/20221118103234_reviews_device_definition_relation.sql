-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE reviews ADD CONSTRAINT device_definition_id_fkey FOREIGN KEY (device_definition_id) REFERENCES device_definitions (id)

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE reviews DROP CONSTRAINT device_definition_id_fkey;
-- +goose StatementEnd
