-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
drop table device_definitions_api.device_makes;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- no going back
-- +goose StatementEnd
