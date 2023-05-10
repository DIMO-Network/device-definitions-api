package models

import (
	"encoding/json"
	"fmt"
)

type PowertrainType string

const (
	ICE  PowertrainType = "ICE"
	HEV  PowertrainType = "HEV"
	PHEV PowertrainType = "PHEV"
	BEV  PowertrainType = "BEV"
	FCEV PowertrainType = "FCEV"
)

func (p PowertrainType) String() string {
	return string(p)
}

func (p *PowertrainType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Potentially an invalid value.
	switch bv := PowertrainType(s); bv {
	case ICE, HEV, PHEV, BEV, FCEV:
		*p = bv
		return nil
	default:
		return fmt.Errorf("unrecognized value: %s", s)
	}
}

// IntegrationsMetadata represents json stored in integrations table metadata jsonb column
type IntegrationsMetadata struct {
	AutoPiDefaultTemplateID      int                    `json:"autoPiDefaultTemplateId"`
	AutoPiPowertrainToTemplateID map[PowertrainType]int `json:"autoPiPowertrainToTemplateId,omitempty"`
}

type GetIntegrationQueryResult struct {
	ID                           string                 `json:"id"`
	Type                         string                 `json:"type"`
	Style                        string                 `json:"style"`
	Vendor                       string                 `json:"vendor"`
	AutoPiDefaultTemplateID      int                    `json:"auto_pi_default_template_id"`
	RefreshLimitSecs             int                    `json:"refresh_limit_secs"`
	AutoPiPowertrainToTemplateID map[PowertrainType]int `json:"auto_pi_power_train_to_template_id,omitempty"`
	TokenID                      int                    `json:"token_id"`
}
