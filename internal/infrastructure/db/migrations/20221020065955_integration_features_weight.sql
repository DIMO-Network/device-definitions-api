-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE integration_features ADD feature_weight float NULL;

UPDATE integration_features SET feature_weight = t.weight FROM (
   VALUES ('ev_battery', 1),
    ('battery_voltage', 0.5),
    ('fuel_percent_remaining', 0.75),
    ('odometer', 1),
    ('oil', 0.5),
    ('tires', 0.75),
    ('speed', 0.75),
    ('location', 1),
    ('battery_capacity', 1),
    ('charging', 0.5),
    ('range', 0.75),
    ('vin', 1),
    ('cell_tower', 0.5),
    ('engine_runtime', 0.25),
    ('ambient_temperature', 0.25),
    ('barometric_pressure', 0.25),
    ('coolant_temperature', 0.25),
    ('engine_load', 0.25),
    ('engine_speed', 0.25),
    ('throttle_position', 0.25),
    ('fuel_type', 0.25)
) AS t(id, weight) 
WHERE  integration_features.feature_key  = t.id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE integration_features DROP COLUMN feature_weight;
-- +goose StatementEnd
