-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;
CREATE TYPE powertrain AS ENUM
    ('ALL', 'HybridsAndICE', 'ICE', 'HEV', 'PHEV', 'BEV', 'FCEV');

alter table integration_features add column powertrain_type powertrain default 'ALL';
-- set some known expectations for BEV vs hybrids and ICE
update integration_features set powertrain_type = 'BEV' where feature_key = 'ev_battery';
update integration_features set powertrain_type = 'HybridsAndICE' where feature_key = 'battery_voltage';
update integration_features set powertrain_type = 'HybridsAndICE' where feature_key = 'fuel_percent_remaining';
update integration_features set powertrain_type = 'HybridsAndICE' where feature_key = 'oil';
update integration_features set powertrain_type = 'BEV' where feature_key = 'battery_capacity';
update integration_features set powertrain_type = 'BEV' where feature_key = 'charging';
update integration_features set powertrain_type = 'HybridsAndICE' where feature_key = 'coolant_temperature';
update integration_features set powertrain_type = 'BEV' where feature_key = 'battery_capacity';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = device_definitions_api, public;

alter table integration_features drop column powertrain_type;
drop type powertrain;

-- +goose StatementEnd
