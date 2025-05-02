-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table wmis add column manufacturer_name text;
update wmis set manufacturer_name = device_makes.name
from device_makes where wmis.device_make_id = device_makes.id;
alter table wmis drop constraint fk_device_make_id;
alter table wmis drop column device_make_id;
alter table wmis
    add constraint wmis_pk
        primary key (wmi, manufacturer_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
