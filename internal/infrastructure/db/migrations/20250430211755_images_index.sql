-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

create unique index images_definition_id_source_url_uindex
    on images (definition_id, source_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop index images_definition_id_source_url_uindex;
-- +goose StatementEnd
