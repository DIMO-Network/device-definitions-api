-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;

ALTER TABLE device_definitions_api.device_makes
    ADD hardware_template_id varchar(500);

ALTER TABLE device_definitions_api.device_styles
    ADD hardware_template_id varchar(500);

ALTER TABLE device_definitions_api.device_definitions
    ADD hardware_template_id varchar(500);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE device_definitions_api.device_makes
DROP COLUMN hardware_template_id;

ALTER TABLE device_definitions_api.device_styles
DROP COLUMN hardware_template_id;

ALTER TABLE device_definitions_api.device_definitions
DROP COLUMN hardware_template_id;

-- +goose StatementEnd
