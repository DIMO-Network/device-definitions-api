# Resolved Decode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make a VIN decode narrow a multi-trim template to one trim and say how confident it is, instead of asserting one flat record for every configuration of a car.

**Architecture:** The catalog service reads the new template shape (`t/<id>.json`) instead of the flat definition. A pure matcher narrows a template's trims against signals the VIN decoder already produces, returning a trim plus a match quality. `decode_vin` emits that, and `DecodeVinResponse` gains fields for the trim, the template version it came from, and the match quality. A decode that cannot narrow says so rather than picking.

**Tech Stack:** Go 1.24, gomock, protobuf/gRPC. `go test -race ./...`.

**Spec:** `../definitions-worker/docs/superpowers/specs/2026-08-27-open-trim-registry.md`
**Contract:** `../definitions-worker/schema/resolved.schema.json` and `template.schema.json`

## Global Constraints

- **No fallback to `definitions/<id>.json`.** The template import is a prerequisite of this deploy, not a runtime concern. A dual read path would be a second thing to maintain and would mask an incomplete import — the failure this whole migration exists to stop.
- **Proto changes are additive only.** New field numbers. Never renumber or remove an existing field; consumers are deployed independently.
- **A decode that cannot narrow must say so.** `match.quality` is `exact`, `ambiguous`, or `model-only`. On `ambiguous`, emit only the attributes every candidate trim agrees on and name the candidates. Never pick one arbitrarily — that is precisely the bug this replaces.
- Attribute values are typed; absent, never empty.
- **Commit messages must contain no `Co-Authored-By` trailer and no Claude or AI attribution of any kind.**

## Why this exists

`toyota_camry_2020` in production declares `powertrain_type: ICE` while carrying the hybrid's fuel economy, tank size and OEM code. Every 2020 Camry resolves to it. The extraction pipeline now emits that model-year as ten trims — three HEV at 13.2 gal against seven ICE at 16. This task is what lets a decode pick the right one.

## File Structure

| File | Responsibility |
|---|---|
| `internal/core/models/template.go` *(new)* | `Template`, `Trim`, `TrimSelectors`, `Resolved`, `MatchQuality` — mirroring the JSON Schemas |
| `internal/core/services/trim_match.go` *(new)* | `MatchTrim(tmpl, signals) (Resolved, error)` — pure |
| `internal/infrastructure/gateways/device_definition_catalog_service.go` *(modify)* | Read `t/<id>.json`, decode `Template` |
| `internal/core/queries/decode_vin.go` *(modify)* | Call the matcher, populate the new response fields |
| `pkg/grpc/decoder.proto` *(modify)* | Additive fields |

---

### Task 1: Template types and catalog read

**Files:**
- Create: `internal/core/models/template.go`
- Modify: `internal/infrastructure/gateways/device_definition_catalog_service.go`
- Test: `internal/infrastructure/gateways/device_definition_catalog_service_test.go`

**Interfaces:**
- Produces: `models.Template`, `models.Trim`, `models.TrimSelectors`, `models.TemplateManufacturer`.
  `GetTemplateByID(ctx, id) (*models.Template, *big.Int, error)` on `DeviceDefinitionCatalogService`, replacing `GetDefinitionByID`. `GetDefinitionByIDFresh` becomes `GetTemplateByIDFresh` with the same read-through semantics.

- [ ] **Step 1: Write the failing test**

Add to `device_definition_catalog_service_test.go`. Use the real emitted shape — this is a verbatim trim of the Camry the pipeline produces:

```go
const templateJSON = `{
  "id": "toyota_camry_2020",
  "deviceType": "vehicle",
  "manufacturer": {"slug": "toyota", "name": "Toyota", "tokenId": 131},
  "model": "Camry",
  "year": 2020,
  "attributes": {"number_of_doors": 4, "vehicle_type": "sedan"},
  "trims": [
    {"name": "LE", "selectors": {"manufacturerCode": ["2532"]},
     "attributes": {"powertrain_type": "ICE", "fuel_type": "gasoline", "fuel_tank_capacity_gal": 16, "mpg_city": 28}},
    {"name": "Hybrid LE", "selectors": {"manufacturerCode": ["2559"]},
     "attributes": {"powertrain_type": "HEV", "fuel_type": "gasoline", "fuel_tank_capacity_gal": 13.2, "mpg_city": 51}}
  ],
  "version": 3,
  "createdAt": "2026-08-27T00:00:00.000Z",
  "updatedAt": "2026-08-28T00:00:00.000Z"
}`

func TestTemplateUnmarshal(t *testing.T) {
	var tmpl coremodels.Template
	require.NoError(t, json.Unmarshal([]byte(templateJSON), &tmpl))

	assert.Equal(t, "toyota_camry_2020", tmpl.ID)
	assert.Equal(t, 2020, tmpl.Year)
	assert.Equal(t, 3, tmpl.Version)
	assert.Equal(t, "Toyota", tmpl.Manufacturer.Name)
	assert.Equal(t, 131, tmpl.Manufacturer.TokenID)

	// Attributes are typed, not stringified.
	assert.Equal(t, float64(4), tmpl.Attributes["number_of_doors"])

	require.Len(t, tmpl.Trims, 2)
	assert.Equal(t, "LE", tmpl.Trims[0].Name)
	assert.Equal(t, []string{"2532"}, tmpl.Trims[0].Selectors.ManufacturerCode)
	assert.Equal(t, "ICE", tmpl.Trims[0].Attributes["powertrain_type"])
	assert.Equal(t, 16.0, tmpl.Trims[0].Attributes["fuel_tank_capacity_gal"])
	assert.Equal(t, 13.2, tmpl.Trims[1].Attributes["fuel_tank_capacity_gal"])
}

func TestTemplateHasNoLegacyFields(t *testing.T) {
	// ksuid and tableId were removed from the model deliberately; a template
	// carrying them would mean the worker regressed.
	var raw map[string]any
	require.NoError(t, json.Unmarshal([]byte(templateJSON), &raw))
	assert.NotContains(t, raw, "ksuid")
	assert.NotContains(t, raw, "tableId")
}

func TestGetTemplateByIDReadsTheTemplateKey(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(templateJSON))
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{DefinitionsCatalogURL: srv.URL}, &logger)

	tmpl, tokenID, err := svc.GetTemplateByID(context.Background(), "toyota_camry_2020")
	require.NoError(t, err)
	assert.Equal(t, "/t/toyota_camry_2020.json", gotPath)
	assert.Equal(t, "toyota_camry_2020", tmpl.ID)
	assert.Equal(t, int64(131), tokenID.Int64())
}

func TestGetTemplateByIDDoesNotFallBackToTheOldKey(t *testing.T) {
	// A 404 on t/<id>.json means the import has not run. Falling back to
	// definitions/<id>.json would serve the pre-migration flat record and hide
	// an incomplete import behind apparently-working decodes.
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	logger := zerolog.Nop()
	svc := NewDeviceDefinitionCatalogService(&config.Settings{DefinitionsCatalogURL: srv.URL}, &logger)

	_, _, err := svc.GetTemplateByID(context.Background(), "toyota_camry_2020")
	require.Error(t, err)
	assert.Equal(t, []string{"/t/toyota_camry_2020.json"}, paths)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/infrastructure/gateways/ -run 'Template' -v`
Expected: compile failure — `coremodels.Template` undefined.

- [ ] **Step 3: Write minimal implementation**

```go
// internal/core/models/template.go
package models

// Template mirrors definitions-worker/schema/template.schema.json. Attribute
// values are typed in the contract — a number arrives as a number — so they
// are decoded as any rather than string.
type Template struct {
	ID           string                 `json:"id"`
	DeviceType   string                 `json:"deviceType"`
	Manufacturer TemplateManufacturer   `json:"manufacturer"`
	Model        string                 `json:"model"`
	Year         int                    `json:"year"`
	ImageURI     string                 `json:"imageURI,omitempty"`
	// HardwareTemplateID is DIMO hardware configuration, not a vehicle
	// attribute, which is why it sits outside Attributes and is never
	// validated against the vehicle vocabulary.
	HardwareTemplateID string           `json:"hardwareTemplateId,omitempty"`
	Attributes   map[string]any         `json:"attributes"`
	Trims        []Trim                 `json:"trims"`
	Version      int                    `json:"version"`
	Author       string                 `json:"author,omitempty"`
	CreatedAt    string                 `json:"createdAt,omitempty"`
	UpdatedAt    string                 `json:"updatedAt,omitempty"`
}

type TemplateManufacturer struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	TokenID int    `json:"tokenId,omitempty"`
}

type Trim struct {
	Name               string         `json:"name"`
	Selectors          TrimSelectors  `json:"selectors"`
	HardwareTemplateID string         `json:"hardwareTemplateId,omitempty"`
	Attributes         map[string]any `json:"attributes"`
}

type TrimSelectors struct {
	ManufacturerCode []string `json:"manufacturerCode,omitempty"`
	StyleName        []string `json:"styleName,omitempty"`
	VINPattern       string   `json:"vinPattern,omitempty"`
}
```

In the catalog service, replace `fetchDoc`'s key with `"/t/" + url.PathEscape(id) + ".json"`, decode into `models.Template`, and return it. Rename `GetDefinitionByID` → `GetTemplateByID` and `GetDefinitionByIDFresh` → `GetTemplateByIDFresh` throughout, regenerating the gomock file.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/infrastructure/gateways/ -run 'Template' -v && go build ./...`

- [ ] **Step 5: Regenerate mocks and fix every caller**

Run: `go generate ./... && go build ./... && go test -race ./... 2>&1 | tail -30`
Every compile error is a caller of the renamed methods. Fix them; do not add adapters.

- [ ] **Step 6: Commit**

```bash
git add -A internal pkg
git commit -m "feat(catalog): read vehicle templates instead of flat definitions"
```

---

### Task 2: The trim matcher

**Files:**
- Create: `internal/core/services/trim_match.go`
- Test: `internal/core/services/trim_match_test.go`

**Interfaces:**
- Consumes: `models.Template` (Task 1).
- Produces:
```go
type MatchSignals struct {
	ManufacturerCode string
	StyleName        string
	VIN              string
}
type MatchQuality string
const (
	MatchExact     MatchQuality = "exact"
	MatchAmbiguous MatchQuality = "ambiguous"
	MatchModelOnly MatchQuality = "model-only"
)
type Resolved struct {
	DefinitionID    string
	TemplateVersion int
	Trim            string
	Quality         MatchQuality
	MatchedBy       []string
	Candidates      []string
	Attributes      map[string]any
	HardwareTemplateID string
}
func MatchTrim(tmpl *models.Template, sig MatchSignals) Resolved
```

- [ ] **Step 1: Write the failing test**

```go
package services

import (
	"testing"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func camry() *coremodels.Template {
	return &coremodels.Template{
		ID: "toyota_camry_2020", DeviceType: "vehicle",
		Manufacturer: coremodels.TemplateManufacturer{Slug: "toyota", Name: "Toyota", TokenID: 131},
		Model: "Camry", Year: 2020, Version: 3,
		Attributes: map[string]any{"number_of_doors": float64(4)},
		Trims: []coremodels.Trim{
			{Name: "LE", Selectors: coremodels.TrimSelectors{ManufacturerCode: []string{"2532"}},
				Attributes: map[string]any{"powertrain_type": "ICE", "fuel_tank_capacity_gal": 16.0, "driven_wheels": "FWD"}},
			{Name: "Hybrid LE", Selectors: coremodels.TrimSelectors{ManufacturerCode: []string{"2559"}},
				Attributes: map[string]any{"powertrain_type": "HEV", "fuel_tank_capacity_gal": 13.2, "driven_wheels": "FWD"}},
		},
	}
}

func TestMatchTrim_ExactOnManufacturerCode(t *testing.T) {
	r := MatchTrim(camry(), MatchSignals{ManufacturerCode: "2559"})
	assert.Equal(t, MatchExact, r.Quality)
	assert.Equal(t, "Hybrid LE", r.Trim)
	assert.Equal(t, []string{"manufacturerCode"}, r.MatchedBy)
	assert.Equal(t, "HEV", r.Attributes["powertrain_type"])
	assert.Equal(t, 13.2, r.Attributes["fuel_tank_capacity_gal"])
	// Template-level attributes merge in.
	assert.Equal(t, float64(4), r.Attributes["number_of_doors"])
	assert.Equal(t, 3, r.TemplateVersion)
}

func TestMatchTrim_TheBlendedCamryIsNoLongerPossible(t *testing.T) {
	// The production record declares ICE while carrying the hybrid's 13.2 gal
	// tank. Whichever trim matches, powertrain and tank must come from the
	// SAME trim.
	for _, tc := range []struct{ code, pt string; tank float64 }{
		{"2532", "ICE", 16.0},
		{"2559", "HEV", 13.2},
	} {
		r := MatchTrim(camry(), MatchSignals{ManufacturerCode: tc.code})
		assert.Equal(t, tc.pt, r.Attributes["powertrain_type"])
		assert.Equal(t, tc.tank, r.Attributes["fuel_tank_capacity_gal"])
	}
}

func TestMatchTrim_AmbiguousEmitsOnlyAgreedAttributes(t *testing.T) {
	tmpl := camry()
	// Both trims claim the same code: a template defect, not a decode failure.
	tmpl.Trims[1].Selectors.ManufacturerCode = []string{"2532"}
	r := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2532"})
	assert.Equal(t, MatchAmbiguous, r.Quality)
	assert.ElementsMatch(t, []string{"LE", "Hybrid LE"}, r.Candidates)
	assert.Empty(t, r.Trim)
	// They disagree on powertrain and tank, agree on drive.
	assert.NotContains(t, r.Attributes, "powertrain_type")
	assert.NotContains(t, r.Attributes, "fuel_tank_capacity_gal")
	assert.Equal(t, "FWD", r.Attributes["driven_wheels"])
}

func TestMatchTrim_ModelOnlyWhenNothingMatches(t *testing.T) {
	r := MatchTrim(camry(), MatchSignals{ManufacturerCode: "9999"})
	assert.Equal(t, MatchModelOnly, r.Quality)
	assert.Empty(t, r.Trim)
	// Only template-level attributes survive; no trim's values leak in.
	assert.Equal(t, float64(4), r.Attributes["number_of_doors"])
	assert.NotContains(t, r.Attributes, "powertrain_type")
}

func TestMatchTrim_SingleTrimTemplateMatchesWithoutSelectors(t *testing.T) {
	tmpl := camry()
	tmpl.Trims = tmpl.Trims[:1]
	tmpl.Trims[0].Selectors = coremodels.TrimSelectors{}
	r := MatchTrim(tmpl, MatchSignals{})
	assert.Equal(t, MatchExact, r.Quality)
	assert.Equal(t, "LE", r.Trim)
}

func TestMatchTrim_StyleNameIsCaseInsensitive(t *testing.T) {
	tmpl := camry()
	tmpl.Trims[0].Selectors = coremodels.TrimSelectors{StyleName: []string{"LE 4dr Sedan"}}
	r := MatchTrim(tmpl, MatchSignals{StyleName: "le 4dr sedan"})
	assert.Equal(t, MatchExact, r.Quality)
	assert.Equal(t, "LE", r.Trim)
}

func TestMatchTrim_TrimOverridesTemplateHardwareTemplateID(t *testing.T) {
	tmpl := camry()
	tmpl.HardwareTemplateID = "130"
	tmpl.Trims[1].HardwareTemplateID = "115"
	r := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2559"})
	assert.Equal(t, "115", r.HardwareTemplateID)
	r2 := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2532"})
	assert.Equal(t, "130", r2.HardwareTemplateID)
}

func TestMatchTrim_EmptyTemplateIsModelOnlyNotAPanic(t *testing.T) {
	tmpl := camry()
	tmpl.Trims = nil
	r := MatchTrim(tmpl, MatchSignals{ManufacturerCode: "2532"})
	assert.Equal(t, MatchModelOnly, r.Quality)
	require.NotNil(t, r.Attributes)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/services/ -run MatchTrim -v`
Expected: compile failure.

- [ ] **Step 3: Write minimal implementation**

Match every trim whose selectors all agree with the signals. Exactly one match is `exact`; several is `ambiguous`; none is `model-only`. Merge template attributes first, then the matched trim's. For `ambiguous`, include only attributes on which every candidate agrees. `HardwareTemplateID` takes the matched trim's if set, else the template's.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -race ./internal/core/services/ -run MatchTrim -v`

- [ ] **Step 5: Commit**

```bash
git add internal/core/services/trim_match.go internal/core/services/trim_match_test.go
git commit -m "feat(decode): narrow a template to one trim, or say why not"
```

---

### Task 3: Emit the resolved decode

**Files:**
- Modify: `pkg/grpc/decoder.proto`, `internal/core/queries/decode_vin.go`
- Test: `internal/core/queries/decode_vin_test.go`

**Interfaces:**
- Consumes: `MatchTrim` (Task 2), `GetTemplateByID` (Task 1).
- Proto gains, additively: `string trim = 12;`, `int32 template_version = 13;`, `string match_quality = 14;`, `repeated string match_candidates = 15;`.

- [ ] **Step 1: Add the proto fields and regenerate**

Append to `DecodeVinResponse` — **new field numbers only**, nothing renumbered or removed:

```protobuf
  // Trim the VIN narrowed to. Empty when match_quality is not "exact".
  string trim = 12;
  // Version of the template this answer came from, so a caller can pin,
  // cache and reproduce a decode.
  int32 template_version = 13;
  // "exact", "ambiguous", or "model-only".
  string match_quality = 14;
  // Trim names still in contention when match_quality is "ambiguous".
  repeated string match_candidates = 15;
```

Run: `make gen` (or the repo's protoc target) and `go build ./...`

- [ ] **Step 2: Write the failing test**

Add to `decode_vin_test.go`, following its existing gomock setup: a decode whose upstream returns a style matching `2559` populates `Trim: "Hybrid LE"`, `MatchQuality: "exact"`, `TemplateVersion: 3`, and `Powertrain: "HEV"`; a decode matching nothing yields `MatchQuality: "model-only"` with an empty `Trim` and no powertrain from any trim; and an ambiguous template yields `MatchQuality: "ambiguous"` with `MatchCandidates` populated.

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/core/queries/ -run DecodeVIN -v`

- [ ] **Step 4: Wire the matcher into `decode_vin`**

Replace the `GetDefinitionByID` call with `GetTemplateByID`, build `MatchSignals` from what the decoder already produces — `vinInfo.StyleName`, the VIN itself, and the OEM code where a source supplies one — call `MatchTrim`, and populate the new response fields from the result. Keep `resp.Powertrain` populated from the resolved attributes rather than from `ResolvePowerTrainFromVinInfo`, which was the old derivation.

**Leave `processDeviceStyle` alone.** It writes to Postgres and is the source the extraction pipeline reads; changing it is a separate decision.

- [ ] **Step 5: Run the full suite**

Run: `go test -race ./... 2>&1 | tail -20` and `golangci-lint run ./... 2>&1 | tail -20`

- [ ] **Step 6: Commit**

```bash
git add -A pkg internal
git commit -m "feat(decode): return the matched trim and how confident the match is"
```

---

## Self-review

**Spec coverage.** `resolved.schema.json`'s `match.quality`, `templateVersion`, `trim` and candidate reporting are implemented in Tasks 2 and 3. The template contract is Task 1. `hardwareTemplateId`'s template-default-with-trim-override resolution is covered in Task 2 — the spec item that no previous plan implemented.

**Deliberately not covered.** Dropping dd-api's search, CRUD and Typesense sync belongs to its own plan; it is a large deletion with its own review surface. `processDeviceStyle` keeps writing styles to Postgres, because that table is the extraction pipeline's input and stopping the writes is a data decision, not a refactor.

**The riskiest assumption** is that `manufacturerCode` reaches `decode_vin` as a usable signal. The templates key trims on it, and the extraction measured 12,275 drivly rows carrying it, but whether the *decoder* surfaces it per-VIN is unverified in this repo. If it does not, matching falls back to `styleName` and quality degrades to `model-only` more often than it should — which is visible in the response rather than silent, but Task 3 must report how often it happens on real decodes.

**Type consistency.** `models.Template` and its children are defined in Task 1 and consumed unchanged in Tasks 2 and 3. `Resolved` and `MatchQuality` are defined in Task 2 and consumed in Task 3. Attribute values are `any` throughout, matching the typed-value contract.
