//nolint:all
package models

type PowerTrainTypeRuleData struct {
	PowerTrainTypeList []PowerTrainType           `yaml:"types"`
	DrivlyList         []PowerTrainTypeOptionData `yaml:"drivly"`
	VincarioList       []PowerTrainTypeOptionData `yaml:"vincario"`
}

type PowerTrainType struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Default bool     `yaml:"default"`
	Makes   []string `yaml:"makes"`
	Models  []string `yaml:"models"`
}

type PowerTrainTypeOptionData struct {
	Type   string   `yaml:"type"`
	Values []string `yaml:"values"`
}

type DrivlyData struct {
	VIN   string `json:"vin"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Fuel  string `json:"fuel"`
}

type VincarioData struct {
	VIN      string `json:"VIN"`
	Make     string `json:"Make"`
	Model    string `json:"Model"`
	FuelType string `json:"FuelType"`
}
