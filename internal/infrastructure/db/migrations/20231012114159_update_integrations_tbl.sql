-- +goose Up
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
ALTER TABLE integrations ADD COLUMN points int;
ALTER TABLE integrations ADD COLUMN manufacturer_token_id numeric(78);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
ALTER TABLE integrations DROP COLUMN points;
ALTER TABLE integrations DROP COLUMN manufacturer_token_id;
-- +goose StatementEnd
