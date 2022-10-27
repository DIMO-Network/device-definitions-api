-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';


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
REFERENCES device_definitions_api.device_definitions(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS images;
-- +goose StatementEnd
