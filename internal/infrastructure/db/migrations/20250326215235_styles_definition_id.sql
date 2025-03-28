-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

alter table device_styles add column definition_id text;

-- seed the manufacturer column, drop the device_make_id col, make manuf not null
update vin_numbers set manufacturer_name = device_makes.name from device_makes where vin_numbers.device_make_id = device_makes.id;

alter table vin_numbers alter column manufacturer_name set not null;

alter table vin_numbers drop column device_make_id;

update device_styles set definition_id = device_definitions.name_slug from device_definitions where device_styles.device_definition_id = device_definitions.id;

alter table device_styles alter column definition_id set not null;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

alter table device_styles drop column definition_id;
alter table vin_numbers add column device_make_id text;
-- +goose StatementEnd
