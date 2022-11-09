-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = device_definitions_api, public;

CREATE TABLE IF NOT EXISTS device_types
(
    id          character(50) PRIMARY KEY NOT NULL, -- do not use ksuid. Use slug id
    created_at  timestamp with time zone  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp with time zone  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name        text                      not null,
    metaDataKey text                      not null,
    properties  jsonb
);

alter table device_definitions
    add column device_type_id character(50) null;

alter table device_definitions
    add constraint fk_device_types
        foreign key (device_type_id) references device_types
            on delete cascade;

delete from device_types;
INSERT INTO device_types (id, name, metaDataKey, properties)
VALUES ('vehicle', 'Vehicle information', 'vehicle_info', '{
  "properties": [
    {
      "name": "fuel_type",
      "label": "Fuel Type",
      "description": "",
      "type": "string",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "driven_wheels",
      "label": "Driven Wheels",
      "description": "",
      "type": "number",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "number_of_doors",
      "label": "Number of Doors",
      "description": "",
      "type": "number",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "base_MSRP",
      "label": "Base MSRP",
      "description": "",
      "type": "number",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "EPA_class",
      "label": "EPA Class",
      "description": "",
      "type": "string",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "vehicle_type",
      "label": "Vehicle Type",
      "description": "",
      "type": "string",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "MPG_highway",
      "label": "MPG Highway",
      "description": "",
      "type": "number",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "MPG_city",
      "label": "MPG City",
      "description": "",
      "type": "number",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "fuel_tank_capacity_gal",
      "label": "Fuel tank capacity gal",
      "description": "",
      "type": "number",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "MPG",
      "label": "MPG",
      "description": "",
      "type": "number",
      "required": false,
      "defaultValue": "",
      "options": []
    },
    {
      "name": "generation",
      "type": "number",
      "label": "Generation",
      "options": [],
      "required": false,
      "description": "Manufacturer Model generation",
      "default_value": ""
    },
    {
      "name": "manufacturer_code",
      "type": "string",
      "label": "Manufacturer Code",
      "options": [],
      "required": false,
      "description": "Manufacturer internal code to describe model body",
      "default_value": ""
    },
    {
      "name": "wheelbase",
      "type": "number",
      "label": "Wheelbase",
      "options": [],
      "required": false,
      "description": "wheelbase is the distance between front and rear wheels",
      "default_value": ""
    }
  ]
}');
select *
from device_types;
update device_definitions set device_type_id = 'vehicle'; -- update all records to default

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path to device_definitions_api, public;

drop table device_types;
alter table device_definitions drop column device_type_id

-- +goose StatementEnd