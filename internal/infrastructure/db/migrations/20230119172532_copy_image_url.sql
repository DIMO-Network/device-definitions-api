-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path TO device_definitions_api,public;

insert into images (id, device_definition_id, fuel_api_id, width, height, source_url, dimo_s3_url, color)
    select COALESCE(gen_random_ksuid(), gen_random_ksuid(), gen_random_ksuid()), id, null, null, null, image_url, null, 'default'
    from device_definitions where image_url is not null and length(image_url) > 1;

alter table device_definitions drop column image_url;
alter table images add column not_exact_image boolean not null default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path TO device_definitions_api,public;

alter table device_definitions add column image_url text;
alter table images drop column not_exact_image;

-- +goose StatementEnd
