-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

CREATE TABLE vin_numbers (
     vin char(17) NOT NULL,
     wmi char(3) NOT NULL, -- World manufacturer Identifier
     vds char(6) NOT NULL, -- Vehicle description Section
     check_digit char(1) NOT NULL,
     serial_number char(6) NOT NULL,
     vis char(6) NOT NULL, -- Vehicle Identification Section
     device_make_id character(27) COLLATE pg_catalog."default" NOT NULL,
     device_definition_id character(27) COLLATE pg_catalog."default" NOT NULL,
     created_at timestamptz NOT NULL DEFAULT current_timestamp,
     updated_at timestamptz NOT NULL DEFAULT current_timestamp,
     CONSTRAINT vin_numbers_pkey PRIMARY KEY (vin),
     CONSTRAINT vin_numbers_name_key UNIQUE (vin)
);

ALTER TABLE vin_numbers ADD CONSTRAINT vin_numbers_device_make_id_fkey FOREIGN KEY (device_make_id) REFERENCES device_makes (id);
ALTER TABLE vin_numbers ADD CONSTRAINT vin_numbers_device_definition_id_fkey FOREIGN KEY (device_definition_id) REFERENCES device_definitions (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

DROP TABLE vin_numbers;
-- +goose StatementEnd
