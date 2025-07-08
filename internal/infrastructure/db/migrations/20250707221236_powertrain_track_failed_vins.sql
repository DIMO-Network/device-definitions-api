-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table device_definitions_api.vin_numbers add column powertrain_type powertrain;

create table device_definitions_api.failed_vin_decodes (
    vin varchar(17) not null primary key,
    vendors_tried text[],
    vincario_data     jsonb,
    drivly_data       jsonb,
    autoiso_data      jsonb,
    datgroup_data     jsonb,
    vin17_data        jsonb,
    manufacturer_name text,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
alter table device_definitions_api.vin_numbers drop column powertrain_type;
drop table device_definitions_api.failed_vin_decodes;
-- +goose StatementEnd
