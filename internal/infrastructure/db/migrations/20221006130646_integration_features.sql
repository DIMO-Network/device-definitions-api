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
    ('fuel_percent_remaining', 'Fuel Remaining', 'fuelPercentRemaining'),
    ('odometer', 'Odometer', 'odometer'),
    ('oil', 'Engine Oil Life', 'oil'),
    ('tires', 'Tires', 'tires.frontLeft'),
    ('speed', 'Speed', 'speed'),
    ('location', 'Location', 'latitude'),
    ('battery_capacity', 'Battery Capacity', 'batteryCapacity'),
    ('charging', 'Charging Status', 'charging'),
    ('range', 'Range', 'range'),
    ('vin', 'VIN', 'vin'),
    ('cell_tower', 'Cell Tower Info', 'cell.ip'),
    ('engine_runtime', 'Engine Run Time', 'runTime'),
    ('ambient_temperature', 'Ambient Temperature', 'ambientTemp'),
    ('barometric_pressure', 'Barometric Pressure', 'barometricPressure'),
    ('coolant_temperature', 'Coolant Temperature', 'coolantTemp'),
    ('engine_load', 'Engine Load', 'engineLoad'),
    ('engine_speed', 'Engine Speed', 'engineSpeed'),
    ('throttle_position', 'Gas Pedal Position', 'throttlePosition');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET search_path to device_definitions_api, public;

DROP TABLE integration_features
-- +goose StatementEnd
