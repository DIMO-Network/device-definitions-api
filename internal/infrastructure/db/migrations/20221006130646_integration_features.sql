-- +goose Up
-- +goose StatementBegin
SET search_path to device_definitions_api, public;

CREATE TABLE IF NOT EXISTS integration_features
(
    feature_key         varchar(50) PRIMARY KEY not null,
    elastic_property    varchar(50) not null,
    display_name        varchar(50) not null,
    css_icon            varchar(100),
    created_at          timestamptz not null default current_timestamp,
    updated_at          timestamptz not null default current_timestamp
);

CREATE UNIQUE INDEX elastic_property_idx ON integration_features (elastic_property);

INSERT INTO integration_features (feature_key, display_name, elastic_property) VALUES 
    ('ev_battery', 'EV Battery', 'soc'),
    ('battery_voltage', 'Battery Voltage', 'batteryVoltage'),
    ('fuel_tank', 'Fuel Tank', 'fuelPercentRemaining'),
    ('odometer', 'Odometer', 'odometer'),
    ('oil', 'Engine Oil Life', 'oil'),
    ('tires', 'Tires', 'tires.frontLeft'),
    ('speed', 'Speed', 'speed');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path to device_definitions_api, public;

DROP TABLE integration_features
-- +goose StatementEnd
