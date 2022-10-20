-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE integration_features ADD feature_weight float NULL;
BEGIN;
    UPDATE integration_features SET feature_weight = 1 WHERE feature_key = 'ev_battery';
    UPDATE integration_features SET feature_weight = 0.5 WHERE feature_key = 'battery_voltage';
    UPDATE integration_features SET feature_weight = 0.75 WHERE feature_key = 'fuel_percent_remaining';
    UPDATE integration_features SET feature_weight = 1 WHERE feature_key = 'odometer';
    UPDATE integration_features SET feature_weight = 0.5 WHERE feature_key = 'oil';
    UPDATE integration_features SET feature_weight = 0.75 WHERE feature_key = 'tires';
    UPDATE integration_features SET feature_weight = 0.75 WHERE feature_key = 'speed';
    UPDATE integration_features SET feature_weight = 1 WHERE feature_key = 'location';
    UPDATE integration_features SET feature_weight = 1 WHERE feature_key = 'battery_capacity';
    UPDATE integration_features SET feature_weight = 0.5 WHERE feature_key = 'charging';
    UPDATE integration_features SET feature_weight = 0.75 WHERE feature_key = 'range';
    UPDATE integration_features SET feature_weight = 1 WHERE feature_key = 'vin';
    UPDATE integration_features SET feature_weight = 0.5 WHERE feature_key = 'cell_tower';
    UPDATE integration_features SET feature_weight = 0.25 WHERE feature_key = 'engine_runtime';
    UPDATE integration_features SET feature_weight = 0.25 WHERE feature_key = 'ambient_temperature';
    UPDATE integration_features SET feature_weight = 0.25 WHERE feature_key = 'barometric_pressure';
    UPDATE integration_features SET feature_weight = 0.25 WHERE feature_key = 'coolant_temperature';
    UPDATE integration_features SET feature_weight = 0.25 WHERE feature_key = 'engine_load';
    UPDATE integration_features SET feature_weight = 0.25 WHERE feature_key = 'engine_speed';
    UPDATE integration_features SET feature_weight = 0.25 WHERE feature_key = 'throttle_position';
    UPDATE integration_features SET feature_weight = 0.25 WHERE feature_key = 'fuel_type';
COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE integration_features DROP COLUMN feature_weight;
-- +goose StatementEnd
