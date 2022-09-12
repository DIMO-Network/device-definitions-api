package common

type RegionEnum string

const (
	AmericasRegion RegionEnum = "Americas"
	EuropeRegion   RegionEnum = "Europe"
)

func (r RegionEnum) String() string {
	return string(r)
}
