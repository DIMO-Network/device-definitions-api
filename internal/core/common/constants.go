package common

const (
	PowerTrainType = "powertrain_type"

	// VehicleMetadataKey is the default json key where we store vehicle metadata in device_definitions.metadata
	VehicleMetadataKey     = "vehicle_info"
	AftermarketMetadataKey = "aftermarket-device"
)

// todo we should get rid of above and only use what we have in tableland as device_attributes

type RegionEnum string

const (
	AmericasRegion RegionEnum = "Americas"
	EuropeRegion   RegionEnum = "Europe"
)

const (
	SmartCarVendor = "SmartCar"
	AutoPiVendor   = "AutoPi"
)

const (
	DefaultDeviceType     = "vehicle"
	DefautlAutoPiTemplate = "130" // dimo device client template
)

func (r RegionEnum) String() string {
	return string(r)
}
