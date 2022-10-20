-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS images
(
    id character(27) NOT NULL,
    device_definition_id NOT NULL,
    fuel_api_id TEXT,
    width INT,
    height INT,
    source_url TEXT NOT NULL,
    dimo_s3_url TEXT,
    color TEXT NOT NULL,
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS images;
-- +goose StatementEnd
