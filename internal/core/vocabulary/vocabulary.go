package vocabulary

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// This is a port of definitions-worker/scripts/extract/vocabulary.mjs. The two
// must agree: the extraction import and a decode-path create write into the
// same contract, and a value one of them folds and the other drops produces a
// template whose attributes depend on which path created it.
//
// The governing rule, unchanged from the original: map a different name for
// the same thing; drop a value whose meaning would have to be invented.

// DeviceType is /schema/device-type-vehicle.json.
type DeviceType struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Type         string   `json:"type"`
	Unit         string   `json:"unit,omitempty"`
	Options      []string `json:"options,omitempty"`
	Minimum      *float64 `json:"minimum,omitempty"`
	Maximum      *float64 `json:"maximum,omitempty"`
	VariesByTrim bool     `json:"variesByTrim,omitempty"`
}

// DroppedAttribute records a value that was read and deliberately not carried.
// Reporting only parse failures is what makes a silent drop silent: the
// extraction's own follow-ups name this as the recurring defect class.
type DroppedAttribute struct {
	Name   string
	Value  string
	Reason string
}

func (d DroppedAttribute) String() string {
	return fmt.Sprintf("%s=%q: %s", d.Name, d.Value, d.Reason)
}

// rename maps a source field onto the vocabulary's name for it. The unit
// belongs in the vocabulary's `unit`, not in the attribute name.
var rename = map[string]string{
	"epa_class": "emissions_standard",
	"wheelbase": "wheelbase_in",
	"mpg":       "mpg_combined",
}

// valueMap is the vocabulary the catalog never had: Options[] was defined in
// 2022 and never populated, so the data holds 14 spellings of fuel_type, 17 of
// driven_wheels and 15 of epa_class.
//
// fuel_type hybrid -> gasoline is deliberate. A hybrid burns gasoline, and
// powertrain derivation has already turned "Hybrid" into powertrain_type HEV
// before this runs, so folding the leftover fuel_type string onto its actual
// fuel is what stops a hybrid record claiming a fuel the vocabulary lacks.
var valueMap = map[string]map[string]string{
	"fuel_type": {
		"gasoline": "gasoline", "petrol": "gasoline", "diesel": "diesel",
		"electric": "electric", "hybrid": "gasoline",
		"flexible-fuel": "flex_fuel", "flexible fuel vehicle (ffv)": "flex_fuel",
		"hydrogen": "hydrogen", "cng": "cng", "lpg": "lpg",
	},
	"driven_wheels": {
		"fwd": "FWD", "front-wheel drive": "FWD", "front wheel drive": "FWD",
		"rwd": "RWD", "rear-wheel drive": "RWD", "rear wheel drive": "RWD",
		"awd": "AWD", "all wheel drive": "AWD", "all-wheel drive": "AWD",
		"4wd": "4WD", "4x4": "4WD", "4x4 - four-wheel drive": "4WD",
		"4wd/4-wheel drive/4x4": "4WD",
		// Slash-joined forms where both halves name the same axle: the source
		// said the same thing twice, so folding them asserts nothing new.
		// Unlike 4x2, there is no axle-count/drive-type conflict to arbitrate.
		"fwd/front-wheel drive": "FWD",
		"rwd/rear-wheel drive":  "RWD", "rwd/ rear wheel drive": "RWD",
		"awd/all-wheel drive": "AWD",
		// '4x2' is deliberately absent -- see ambiguous below.
	},
	"vehicle_type": {
		"passenger car": "sedan", "sedan/saloon": "sedan", "sedan": "sedan",
		"hatchback/liftback/notchback": "hatchback", "hatchback": "hatchback",
		"wagon": "wagon", "coupe": "coupe", "cabriolet/convertible": "convertible",
		"multipurpose passenger vehicle (mpv)":                    "suv",
		"sport utility vehicle (suv)/multi purpose vehicle (mpv)": "suv",
		"suv":     "suv",
		"minivan": "minivan", "van": "van", "pickup": "pickup",
		"truck": "truck", "conventional type truck": "truck",
		"motorcycle": "motorcycle", "incomplete vehicle": "chassis_cab",
		"incomplete - chassis cab (single cab)": "chassis_cab",
		// A crossover is a different name for a category the vocabulary already
		// offers, not an assertion the source did not make.
		"crossover utility vehicle (cuv)": "suv",
	},
	"emissions_standard": {
		"euro 4": "euro_4", "euro 5": "euro_5",
		"euro 6": "euro_6", "euro6": "euro_6", "euro 6 b": "euro_6",
		"euro 6 w": "euro_6", "euro 6 dg": "euro_6",
		"euro 6d": "euro_6d", "euro6d-temp": "euro_6d", "euro 6d-temp": "euro_6d",
		"euro 6d-isc-fcm": "euro_6d", "euro 6d/euro 6": "euro_6d",
	},
}

// '4x2' and '6x4' state how many wheels are driven, not which: 4x2 is RWD on
// most pickups and FWD on most cars. A wrong guess looks exactly like a
// verified fact downstream and nobody rechecks it; a dropped attribute is
// recoverable by a Console contributor who knows the vehicle.
const wheelCountAmbiguous = "ambiguous: names a driven wheel count, not an axle — could be RWD or FWD (or split further on multi-axle trucks) depending on the vehicle"

// '6L'/'5L'/'4L'/'U' are a real epa_class coding nobody on this migration has
// identified. Unlike 4x2 the value does not assert the wrong thing; its
// meaning is unknown, so there is nothing to map it to without guessing.
const unidentifiedEPACode = "unmapped: a known epa_class code with no vocabulary equivalent, deliberately left unmapped pending identification of the source coding scheme"

// ambiguous routes deliberate drops through their own reason. The generic
// "value not in the vocabulary" reason is used for genuine spelling gaps and
// typos, and would make these decisions invisible to anyone reading the drop
// report rather than this file -- which is everyone the report is for.
var ambiguous = map[string]map[string]string{
	"driven_wheels": {
		"4x2": wheelCountAmbiguous,
		"6x4": wheelCountAmbiguous,
	},
	"emissions_standard": {
		"6l": unidentifiedEPACode,
		"5l": unidentifiedEPACode,
		"4l": unidentifiedEPACode,
		"u":  unidentifiedEPACode,
	},
}

// dead values are the placeholders the old model used for "unknown". The
// contract makes them unrepresentable: an unknown attribute is absent.
var dead = map[string]struct{}{"": {}, "<nil>": {}, "null": {}, "N/A": {}}

var leadingNumber = regexp.MustCompile(`^([\d.]+)`)

// MapAttributes folds source attribute name/value pairs onto the vocabulary,
// returning typed values and an explicit account of everything it did not
// carry. Order of the input does not affect the result.
func MapAttributes(attrs map[string]string, vocab *DeviceType) (map[string]any, []DroppedAttribute) {
	mapped := map[string]any{}
	dropped := []DroppedAttribute{}
	if vocab == nil {
		return mapped, dropped
	}
	defs := make(map[string]Attribute, len(vocab.Attributes))
	for _, a := range vocab.Attributes {
		defs[a.Name] = a
	}

	// Sorted so the drop report is stable between runs; map iteration order
	// would otherwise reorder it on every call.
	for _, rawName := range sortedKeys(attrs) {
		rawValue := attrs[rawName]
		drop := func(reason string) {
			dropped = append(dropped, DroppedAttribute{Name: rawName, Value: rawValue, Reason: reason})
		}
		// Tested against the raw value, as the original does: a
		// whitespace-only value is not dead, it fails its type check below and
		// is reported as a drop rather than vanishing.
		if _, isDead := dead[rawValue]; isDead {
			continue
		}
		name := rawName
		if renamed, ok := rename[rawName]; ok {
			name = renamed
		}
		def, ok := defs[name]
		if !ok {
			drop("no such attribute in the vocabulary")
			continue
		}

		value := strings.TrimSpace(rawValue)

		if reason, isAmbiguous := ambiguous[name][strings.ToLower(value)]; isAmbiguous {
			drop(reason)
			continue
		}

		if name == "wheelbase_in" {
			m := leadingNumber.FindStringSubmatch(value)
			if m == nil {
				drop("unparseable wheelbase")
				continue
			}
			value = m[1]
		}

		switch def.Type {
		case "enum":
			hit, found := valueMap[name][strings.ToLower(value)]
			if !found {
				for _, opt := range def.Options {
					if opt == value {
						hit, found = value, true
						break
					}
				}
			}
			if !found {
				drop("value not in the vocabulary and no mapping")
				continue
			}
			mapped[name] = hit
		case "number", "integer":
			n, err := strconv.ParseFloat(value, 64)
			if err != nil || math.IsInf(n, 0) || math.IsNaN(n) {
				drop("not a number")
				continue
			}
			if def.Type == "integer" && n != math.Trunc(n) {
				drop("not an integer")
				continue
			}
			if def.Minimum != nil && n < *def.Minimum {
				drop("below the vocabulary range")
				continue
			}
			if def.Maximum != nil && n > *def.Maximum {
				drop("above the vocabulary range")
				continue
			}
			if def.Type == "integer" {
				mapped[name] = int(n)
				continue
			}
			mapped[name] = n
		case "boolean":
			if value != "true" && value != "false" {
				drop("not a boolean")
				continue
			}
			mapped[name] = value == "true"
		default:
			mapped[name] = value
		}
	}
	return mapped, dropped
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
