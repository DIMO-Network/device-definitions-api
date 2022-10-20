package common

const (
	DefaultDeviceType = "vehicle"
)

type (
	ErrorResponse struct {
		Field string
		Tag   string
		Value string
	}
)
