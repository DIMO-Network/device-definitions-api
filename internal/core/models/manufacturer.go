package models

type Manufacturer struct {
	TokenID int    `json:"tokenId"`
	Name    string `json:"name"`
	TableID int    `json:"tableId"`
	Owner   string `json:"owner"`
}

type DeviceDefinition struct {
	DeviceDefinitionID string       `json:"deviceDefinitionId"`
	Manufacturer       Manufacturer `json:"manufacturer"`
	Model              string       `json:"model"`
	Year               int          `json:"year"`
}
