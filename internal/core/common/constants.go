package common

const (
	PowerTrainType = "powertrain_type"
	// VehicleMetadataKey is the default json key where we store vehicle metadata in device_definitions.metadata
	VehicleMetadataKey = "vehicle_info"
)

type RegionEnum string

const (
	AmericasRegion RegionEnum = "Americas"
	EuropeRegion   RegionEnum = "Europe"
)

const (
	SmartCarVendor = "SmartCar"
	TeslaVendor    = "Tesla"
	AutoPiVendor   = "AutoPi"
)

const (
	DefaultDeviceType = "vehicle"
)

func (r RegionEnum) String() string {
	return string(r)
}
