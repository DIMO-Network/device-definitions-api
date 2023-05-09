-- +goose Up
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
ALTER TABLE integrations ADD COLUMN token_id int;
ALTER TABLE integrations ADD CONSTRAINT token_id_key UNIQUE (token_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
ALTER TABLE integrations drop column token_id;
-- +goose StatementEnd
