-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS device_definitions_api.integration_features
(
    elastic_property    varchar(50) PRIMARY KEY not null,
    feature_key         varchar(50) not null,
    display_name        varchar(50) not null,
    css_icon            varchar(100),
    created_at          timestamptz not null default current_timestamp,
    updated_at          timestamptz not null default current_timestamp
);
CREATE UNIQUE INDEX feature_key_idx ON device_definitions_api.integration_features (feature_key);
CREATE UNIQUE INDEX css_icon_idx ON device_definitions_api.integration_features (css_icon);

INSERT INTO device_definitions_api.integration_features (feature_key, display_name, elastic_property) VALUES 
    ('ev_battery', 'EV Battery', 'soc'),
    ('battery_voltage', 'Battery Voltage', 'battery_voltage'),
    ('fuel_tank', 'Fuel Tank', 'fuel_percent_remaining'),
    ('odometer', 'Odometer', 'odometer'),
    ('oil', 'Engine Oil Life', 'oil'),
    ('tires', 'Tires', 'tires.frontLeft'),
    ('speed', 'Speed', 'speed');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP INDEX feature_key_idx;
DROP INDEX css_icon_idx;

DROP TABLE device_definitions_api.integration_features
-- +goose StatementEnd
