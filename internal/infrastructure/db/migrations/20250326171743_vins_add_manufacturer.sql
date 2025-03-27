-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

alter table vin_numbers
    add column manufacturer_name TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

alter table vin_numbers
    drop column manufacturer_name;
-- +goose StatementEnd
