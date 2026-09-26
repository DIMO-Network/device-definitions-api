package vocabulary

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MapAttributes is a port, and the risk a port carries is that the two copies
// drift silently. The extraction import and a decode-path create write into the
// same contract: if one folds a value the other drops, a template's attributes
// depend on which path happened to create it, which is unobservable from either
// side.
//
// testdata/vocabulary_parity.json is the OUTPUT of the JavaScript original, run
// against the real vocabulary. Regenerate it with:
//
//	node internal/core/services/testdata/regenerate-parity.mjs > \
//	  internal/core/services/testdata/vocabulary_parity.json
//
// A diff there is the two implementations disagreeing. Resolve it; do not
// regenerate to make this pass.
type parityCase struct {
	In      map[string]string `json:"in"`
	Mapped  map[string]any    `json:"mapped"`
	Dropped []struct {
		Name   string `json:"name"`
		Reason string `json:"reason"`
	} `json:"dropped"`
}

func TestMapAttributesMatchesTheJavaScriptOriginal(t *testing.T) {
	rawVocab, err := os.ReadFile("testdata/device-type-vehicle.json")
	require.NoError(t, err)
	var vocab DeviceType
	require.NoError(t, json.Unmarshal(rawVocab, &vocab))
	require.NotEmpty(t, vocab.Attributes, "the real vocabulary must load, or this compares nothing")

	rawCases, err := os.ReadFile("testdata/vocabulary_parity.json")
	require.NoError(t, err)
	var cases []parityCase
	require.NoError(t, json.Unmarshal(rawCases, &cases))
	require.NotEmpty(t, cases)

	for _, c := range cases {
		mapped, dropped := MapAttributes(c.In, &vocab)

		// Through JSON so an int on one side and a float on the other are
		// compared as the contract sees them, not as Go types.
		var got map[string]any
		b, err := json.Marshal(mapped)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(b, &got))
		if len(got) == 0 {
			got = map[string]any{}
		}
		want := c.Mapped
		if want == nil {
			want = map[string]any{}
		}
		assert.Equal(t, want, got, "mapped values differ for %v", c.In)

		require.Len(t, dropped, len(c.Dropped), "drop count differs for %v", c.In)
		for i, d := range c.Dropped {
			assert.Equal(t, d.Name, dropped[i].Name, "dropped name differs for %v", c.In)
			// The reason is the part a human reads to decide whether a drop was
			// a decision or a defect, so it has to match too, not just the count.
			assert.Equal(t, d.Reason, dropped[i].Reason, "drop reason differs for %v", c.In)
		}
	}
}
