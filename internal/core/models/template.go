//nolint:tagliatelle
package models

// Template mirrors definitions-worker/schema/template.schema.json: a vehicle
// trim-configuration template covering one model-year plus the trims it
// shipped in. Attributes true of every trim live on Attributes; anything
// that differs between trims lives on the trim itself.
//
// Attribute values are typed in the contract -- a number arrives as a
// number -- so they are decoded as any rather than string.
type Template struct {
	ID           string               `json:"id"`
	DeviceType   string               `json:"deviceType"`
	Manufacturer TemplateManufacturer `json:"manufacturer"`
	Model        string               `json:"model"`
	Year         int                  `json:"year"`
	ImageURI     string               `json:"imageURI,omitempty"`
	// HardwareTemplateID is DIMO hardware configuration, not a vehicle
	// attribute, which is why it sits outside Attributes and is never
	// validated against the vehicle vocabulary.
	HardwareTemplateID string         `json:"hardwareTemplateId,omitempty"`
	Attributes         map[string]any `json:"attributes"`
	Trims              []Trim         `json:"trims"`
	Version            int            `json:"version"`
	Author             string         `json:"author,omitempty"`
	CreatedAt          string         `json:"createdAt,omitempty"`
	UpdatedAt          string         `json:"updatedAt,omitempty"`
}

type TemplateManufacturer struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	TokenID int    `json:"tokenId,omitempty"`
}

// Trim is one configuration a Template's model-year shipped in. Every
// template carries at least one, even a model-year sold in a single
// configuration.
type Trim struct {
	Name      string        `json:"name"`
	Selectors TrimSelectors `json:"selectors"`
	// HardwareTemplateID overrides the template's value for this trim only;
	// same "not a vehicle attribute" rule as Template.HardwareTemplateID.
	HardwareTemplateID string         `json:"hardwareTemplateId,omitempty"`
	Attributes         map[string]any `json:"attributes"`
}

// TrimSelectors are the signals a decoded VIN is matched against to narrow a
// Template to exactly one Trim. A trim matches when every selector present
// matches.
type TrimSelectors struct {
	ManufacturerCode []string `json:"manufacturerCode,omitempty"`
	StyleName        []string `json:"styleName,omitempty"`
	VINPattern       string   `json:"vinPattern,omitempty"`
}
