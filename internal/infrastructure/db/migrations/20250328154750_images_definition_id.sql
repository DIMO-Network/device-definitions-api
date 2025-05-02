-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

alter table images add column definition_id text;

update images set definition_id = device_definitions.name_slug from device_definitions where images.device_definition_id = device_definitions.id;

alter table images alter column definition_id set not null;

-- foreign keys
alter table images
    drop constraint fkey_device_definition_id;
alter table device_styles
    drop constraint fk_device_definition;

alter table device_styles
    add constraint device_styles_device_definitions_name_slug_fk
        foreign key (definition_id) references device_definitions (name_slug);

alter table images
    add constraint images_device_definitions_name_slug_fk
        foreign key (definition_id) references device_definitions (name_slug);

-- remove old ksuid
alter table device_styles drop column device_definition_id;
alter table images drop column device_definition_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

alter table images drop column definition_id;

alter table device_styles add column device_definition_id text;
alter table images add column device_definition_id text;

-- +goose StatementEnd
