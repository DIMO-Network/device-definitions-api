-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;

create index if not exists device_nhtsa_recalls_device_definition_id_index
    on device_nhtsa_recalls (device_definition_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
drop index device_nhtsa_recalls_device_definition_id_index;

-- +goose StatementEnd
