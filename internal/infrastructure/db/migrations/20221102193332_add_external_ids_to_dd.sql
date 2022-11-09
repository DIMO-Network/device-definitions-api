-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE device_definitions
    ADD external_ids jsonb;

UPDATE device_definitions
SET external_ids = CASE
                       WHEN external_id IS NULL OR external_id = ''
                           THEN '{}'::jsonb
                       ELSE jsonb_build_object(COALESCE(source, ''), external_id)
    END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE device_definitions
    DROP COLUMN external_ids;

-- +goose StatementEnd
