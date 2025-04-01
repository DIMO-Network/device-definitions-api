-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create table definition_transactions
(
    transaction_hash text not null primary key,
    definition_id    text not null,
    created_at       timestamptz not null default now(),
    manufacturer_id  bigint not null
);
create index definition_transactions_definition_id_index
    on definition_transactions (definition_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table definition_transactions;
-- +goose StatementEnd
