package common

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

func (r RegionEnum) String() string {
	return string(r)
}
