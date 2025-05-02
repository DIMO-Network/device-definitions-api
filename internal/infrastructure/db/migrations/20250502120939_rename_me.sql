-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table vin_numbers
    alter column vin type varchar(17) using vin::varchar(17);

alter table vin_numbers
    alter column wmi drop not null;

alter table vin_numbers
    alter column vds drop not null;

alter table vin_numbers
    alter column check_digit drop not null;

alter table vin_numbers
    alter column serial_number type varchar(10) using serial_number::varchar(10);

alter table vin_numbers
    alter column vis drop not null;

alter table vin_numbers
    add vin17_data jsonb;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

alter table vin_numbers
    drop vin17_data;

alter table vin_numbers
    alter column wmi set not null;

alter table vin_numbers
    alter column vds set not null;

alter table vin_numbers
    alter column check_digit set not null;

alter table vin_numbers
    alter column vis set not null;
-- +goose StatementEnd
