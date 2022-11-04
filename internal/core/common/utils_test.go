package common

import (
	"testing"

	"github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
)

func TestBuildExternalIds(t *testing.T) {

	json := null.JSONFrom([]byte(`{"edmunds": "123", "nhtsa": "qwert", "adac": "890" }`))

	got := BuildExternalIds(json)

	assert.Contains(t, got, models.ExternalID{Vendor: "edmunds", ID: "123"})
	assert.Contains(t, got, models.ExternalID{Vendor: "nhtsa", ID: "qwert"})
	assert.Contains(t, got, models.ExternalID{Vendor: "adac", ID: "890"})
}

func TestExternalIdsToGRPC(t *testing.T) {

	extIds := []models.ExternalID{
		{Vendor: "edmunds", ID: "123"},
		{Vendor: "nhtsa", ID: "qwert"},
		{Vendor: "adac", ID: "890"},
	}

	got := ExternalIdsToGRPC(extIds)

	assert.Equal(t, 3, len(got))

	assert.Equal(t, "edmunds", got[0].Vendor)
	assert.Equal(t, "123", got[0].Id)

	assert.Equal(t, "nhtsa", got[1].Vendor)
	assert.Equal(t, "qwert", got[1].Id)

	assert.Equal(t, "adac", got[2].Vendor)
	assert.Equal(t, "890", got[2].Id)
}
