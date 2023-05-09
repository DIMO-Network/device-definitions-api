-- +goose Up
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
ALTER TABLE integrations ADD COLUMN token_id int;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
ALTER TABLE integrations drop column token_id;
-- +goose StatementEnd
