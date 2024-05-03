-- +goose Up
-- +goose StatementBegin

SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_definitions ADD name_slug VARCHAR(100) null;

UPDATE device_definitions AS dd
SET name_slug = dm.name_slug || '-' || dd.model_slug || '-' || CAST(dd.year AS VARCHAR(4))
    FROM device_makes AS dm
WHERE dd.device_make_id = dm.id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
ALTER TABLE device_definitions drop column name_slug;

-- +goose StatementEnd
