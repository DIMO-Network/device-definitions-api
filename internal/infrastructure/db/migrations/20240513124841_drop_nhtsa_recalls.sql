-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = device_definitions_api, public;
drop table device_nhtsa_recalls;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = device_definitions_api, public;
create table device_nhtsa_recalls
(
    id                     char(27)                                           not null
        primary key,
    device_definition_id   char(27)
        constraint fk_device_definition
            references device_definitions,
    data_record_id         integer                                            not null,
    data_campno            varchar(12)                                        not null,
    data_maketxt           varchar(25)                                        not null,
    data_modeltxt          varchar(256)                                       not null,
    data_yeartxt           integer                                            not null,
    data_mfgcampno         varchar(20)                                        not null,
    data_compname          varchar(256)                                       not null,
    data_mfgname           varchar(40)                                        not null,
    data_bgman             date,
    data_endman            date,
    data_rcltypecd         varchar(4)                                         not null,
    data_potaff            integer,
    data_odate             date,
    data_influenced_by     varchar(4)                                         not null,
    data_mfgtxt            varchar(40)                                        not null,
    data_rcdate            date                                               not null,
    data_datea             date                                               not null,
    data_rpno              varchar(3)                                         not null,
    data_fmvss             varchar(10)                                        not null,
    data_desc_defect       varchar(2000)                                      not null,
    data_conequence_defect varchar(2000)                                      not null,
    data_corrective_action varchar(2000)                                      not null,
    data_notes             varchar(2000)                                      not null,
    data_rcl_cmpt_id       char(27)                                           not null,
    data_mfr_comp_name     varchar(50)                                        not null,
    data_mfr_comp_desc     varchar(200)                                       not null,
    data_mfr_comp_ptno     varchar(100)                                       not null,
    created_at             timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at             timestamp with time zone default CURRENT_TIMESTAMP not null,
    metadata               jsonb,
    hash                   bytea                                              not null
);
grant select on device_nhtsa_recalls to readonly;
-- +goose StatementEnd
