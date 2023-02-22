-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;

ALTER TABLE device_definitions_api.vin_numbers
    ADD style_id varchar(27);


ALTER TABLE device_definitions_api.vin_numbers
    ADD CONSTRAINT fkey_style_id
        FOREIGN KEY (style_id)
            REFERENCES device_styles(id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

ALTER TABLE device_definitions_api.vin_numbers
DROP COLUMN style_id;

-- +goose StatementEnd
