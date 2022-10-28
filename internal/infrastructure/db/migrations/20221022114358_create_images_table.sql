-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

CREATE TABLE IF NOT EXISTS images
(
    id CHAR(27) PRIMARY KEY,
    device_definition_id char(27) NOT NULL,
    fuel_api_id TEXT,
    width INT,
    height INT,
    source_url TEXT NOT NULL,
    dimo_s3_url TEXT,
    color TEXT NOT NULL
);

ALTER TABLE images 
ADD CONSTRAINT fkey_device_definition_id 
FOREIGN KEY (device_definition_id) 
REFERENCES device_definitions(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
DROP TABLE IF EXISTS images;
-- +goose StatementEnd
