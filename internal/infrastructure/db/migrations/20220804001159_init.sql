-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

CREATE TABLE IF NOT EXISTS device_makes
(
    id character(27) COLLATE pg_catalog."default" NOT NULL,
    name text COLLATE pg_catalog."default" NOT NULL,
    external_ids jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    token_id numeric(78,0),
    logo_url text COLLATE pg_catalog."default",
    oem_platform_name text COLLATE pg_catalog."default",
    CONSTRAINT device_makes_pkey PRIMARY KEY (id),
    CONSTRAINT device_makes_name_key UNIQUE (name),
    CONSTRAINT device_makes_token_id_key UNIQUE (token_id)
);

CREATE TABLE IF NOT EXISTS device_definitions
(
    id character(27) COLLATE pg_catalog."default" NOT NULL,
    model character varying(100) COLLATE pg_catalog."default" NOT NULL,
    year smallint NOT NULL,
    image_url text COLLATE pg_catalog."default",
    metadata jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source text COLLATE pg_catalog."default",
    verified boolean NOT NULL DEFAULT false,
    external_id text COLLATE pg_catalog."default",
    device_make_id character(27) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT device_definitions_pkey PRIMARY KEY (id),
    CONSTRAINT idx_device_make_id_model_year UNIQUE (device_make_id, model, year),
    CONSTRAINT fk_device_make_id FOREIGN KEY (device_make_id)
        REFERENCES device_makes (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS device_styles
(
    id character(27) COLLATE pg_catalog."default" NOT NULL,
    device_definition_id character(27) COLLATE pg_catalog."default" NOT NULL,
    name text COLLATE pg_catalog."default" NOT NULL,
    external_style_id text COLLATE pg_catalog."default" NOT NULL,
    source text COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sub_model text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT device_styles_pkey PRIMARY KEY (id),
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_id)
        REFERENCES device_definitions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE UNIQUE INDEX device_definition_name_sub_modelx
    ON device_styles USING btree
    (device_definition_id COLLATE pg_catalog."default" ASC NULLS LAST, source COLLATE pg_catalog."default" ASC NULLS LAST, name COLLATE pg_catalog."default" ASC NULLS LAST, sub_model COLLATE pg_catalog."default" ASC NULLS LAST);

CREATE UNIQUE INDEX device_definition_style_idx
    ON device_styles USING btree
    (device_definition_id COLLATE pg_catalog."default" ASC NULLS LAST, source COLLATE pg_catalog."default" ASC NULLS LAST, external_style_id COLLATE pg_catalog."default" ASC NULLS LAST);
    

CREATE TYPE integration_style AS ENUM
    ('Addon', 'OEM', 'Webhook');

CREATE TYPE integration_type AS ENUM
    ('Hardware', 'API');

CREATE TABLE IF NOT EXISTS integrations
(
    id character(27) COLLATE pg_catalog."default" NOT NULL,
    type integration_type NOT NULL,
    style integration_style NOT NULL,
    vendor character varying(50) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    refresh_limit_secs integer NOT NULL DEFAULT 3600,
    metadata jsonb,
    CONSTRAINT integrations_pkey PRIMARY KEY (id),
    CONSTRAINT idx_integrations_vendor UNIQUE (vendor)
);

COMMENT ON COLUMN integrations.refresh_limit_secs
    IS 'How often can integration be called in seconds';

CREATE TABLE IF NOT EXISTS device_integrations
(
    device_definition_id character(27) COLLATE pg_catalog."default" NOT NULL,
    integration_id character(27) COLLATE pg_catalog."default" NOT NULL,
    capabilities jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    region text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT pkey_device_region PRIMARY KEY (device_definition_id, integration_id, region),
    CONSTRAINT fk_device_definition FOREIGN KEY (device_definition_id)
        REFERENCES device_definitions (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE,
    CONSTRAINT fk_integration FOREIGN KEY (integration_id)
        REFERENCES integrations (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

DROP TABLE device_integrations;
DROP TABLE integrations;

DROP TYPE integration_style;
DROP TYPE integration_type;

DROP INDEX device_definition_style_idx;
DROP INDEX device_definition_name_sub_modelx;

DROP TABLE device_styles;
DROP TABLE device_definitions;
DROP TABLE device_makes;

-- +goose StatementEnd
