package common

import (
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"testing"
)

func TestBuildExternalIds(t *testing.T) {

	json := null.JSONFrom([]byte(`{"edmunds": "123", "nhtsa": "qwert", "adac": "890" }`))

	got := BuildExternalIds(json)

	assert.Equal(t, "edmunds", got[0].Vendor)
	assert.Equal(t, "123", got[0].ID)

	assert.Equal(t, "nhtsa", got[1].Vendor)
	assert.Equal(t, "qwert", got[1].ID)

	assert.Equal(t, "adac", got[2].Vendor)
	assert.Equal(t, "890", got[2].ID)
}
