package common

const (
	PowerTrainType = "powertrain_type"
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
