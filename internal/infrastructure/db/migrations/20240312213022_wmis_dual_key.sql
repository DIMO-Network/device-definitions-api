-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
alter table wmis
    drop constraint wmis_pkey;

alter table wmis
    add primary key (wmi, device_make_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;

alter table wmis
    drop constraint wmis_pkey;
alter table wmis
    add primary key (wmi);
-- +goose StatementEnd
