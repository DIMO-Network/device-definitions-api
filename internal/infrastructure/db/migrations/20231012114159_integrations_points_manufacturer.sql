-- +goose Up
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
ALTER TABLE integrations ADD COLUMN points INT NOT NULL DEFAULT 0;
ALTER TABLE integrations ADD COLUMN manufacturer_token_id INT;
ALTER TABLE integrations ADD CONSTRAINT integrations_manufacturer_token_id_fkey FOREIGN KEY (manufacturer_token_id) REFERENCES device_makes(token_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path = device_definitions_api, public;
ALTER TABLE integrations DROP COLUMN manufacturer_token_id;
ALTER TABLE integrations DROP COLUMN points;
-- +goose StatementEnd
