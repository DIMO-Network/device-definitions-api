-- +goose Up
-- +goose StatementBegin

SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_definitions ADD name_slug VARCHAR(100) null;

UPDATE device_definitions AS dd
SET name_slug = dm.name_slug || '_' || dd.model_slug || '_' || CAST(dd.year AS VARCHAR(4))
    FROM device_makes AS dm
WHERE dd.device_make_id = dm.id;

ALTER TABLE device_definitions ALTER COLUMN name_slug set not null;

CREATE INDEX idx_device_definition_slug
    ON device_definitions_api.device_definitions(name_slug);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_definitions drop column name_slug;

-- +goose StatementEnd
