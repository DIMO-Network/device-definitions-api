-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

CREATE EXTENSION IF NOT EXISTS unaccent;
SELECT unaccent('ô ã À Æ ß Ä Ö ö');

ALTER TABLE device_definitions_api.device_makes
ADD COLUMN name_slug VARCHAR(100);

UPDATE device_definitions_api.device_makes
SET name_slug = unaccent(LOWER(REPLACE(name, ' ', '-')));

ALTER TABLE device_definitions_api.device_definitions
ADD COLUMN model_slug VARCHAR(100);

UPDATE device_definitions_api.device_definitions
SET model_slug = unaccent(LOWER(REPLACE(model, ' ', '-')));

CREATE INDEX idx_name_slug 
ON device_definitions_api.device_makes(name_slug);

CREATE INDEX idx_model_slug
ON device_definitions_api.device_definitions(model_slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
