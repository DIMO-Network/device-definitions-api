-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE device_integrations ADD features jsonb NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE device_integrations DROP COLUMN features;
-- +goose StatementEnd
