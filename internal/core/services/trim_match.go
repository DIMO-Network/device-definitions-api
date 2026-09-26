package services

import (
	"reflect"
	"regexp"
	"strings"

	coremodels "github.com/DIMO-Network/device-definitions-api/internal/core/models"
)

// MatchSignals are the decoded-VIN signals MatchTrim narrows a Template
// against. A zero-value field means that signal was not available; it never
// participates in matching.
type MatchSignals struct {
	ManufacturerCode string
	StyleName        string
	VIN              string
}

// MatchQuality reports how confidently MatchTrim narrowed a Template to a
// single Trim.
type MatchQuality string

const (
	// MatchExact means exactly one trim's selectors matched the signals.
	MatchExact MatchQuality = "exact"
	// MatchAmbiguous means more than one trim's selectors matched. Picking
	// one arbitrarily would silently reintroduce the blended-record bug this
	// matcher exists to fix, so no trim is chosen.
	MatchAmbiguous MatchQuality = "ambiguous"
	// MatchModelOnly means no trim's selectors matched. Only the template's
	// own attributes are returned -- never a partial merge from a trim that
	// nearly matched.
	MatchModelOnly MatchQuality = "model-only"
)

// Resolved is the outcome of narrowing a Template to (at most) one Trim.
type Resolved struct {
	DefinitionID       string
	TemplateVersion    int
	Trim               string
	Quality            MatchQuality
	MatchedBy          []string
	Candidates         []string
	Attributes         map[string]any
	HardwareTemplateID string
}

// MatchTrim narrows tmpl to the single Trim whose selectors agree with sig.
//
// A trim matches when every selector it declares agrees with sig; a trim
// with no selectors at all matches unconditionally (the single-trim,
// no-disambiguation-needed case). Exactly one matching trim is an exact
// match: its attributes are merged over the template's, and its
// powertrain/tank/etc. always come from that same trim -- never a mix of
// trims. Zero matches is model-only: only the template's own attributes are
// returned. More than one match is ambiguous: no trim is chosen, and only
// the attributes every candidate agrees on are returned, alongside the
// names of the trims still in contention.
//
// HardwareTemplateID is DIMO hardware configuration, not a vehicle
// attribute; it is resolved independently of Attributes, taking the matched
// trim's value when set and otherwise the template's. On an ambiguous or
// model-only result, no single trim is chosen, so the template's value is
// used unchanged -- taking a candidate's hardware value when we have
// explicitly declined to pick a candidate would still be picking one, in
// the one field that decides what hardware DIMO ships. The template's value
// is what every candidate shares by construction, so it's the only
// non-arbitrary answer available.
//
// MatchTrim is pure: no I/O, no logging, no clock. The same tmpl and sig
// always produce the same Resolved.
func MatchTrim(tmpl *coremodels.Template, sig MatchSignals) Resolved {
	res := Resolved{
		DefinitionID:       tmpl.ID,
		TemplateVersion:    tmpl.Version,
		Attributes:         copyAttributes(tmpl.Attributes),
		HardwareTemplateID: tmpl.HardwareTemplateID,
	}

	var matches []trimMatch
	for _, trim := range tmpl.Trims {
		if ok, matchedBy := selectorsMatch(trim.Selectors, sig); ok {
			matches = append(matches, trimMatch{trim: trim, matchedBy: matchedBy})
		}
	}

	switch len(matches) {
	case 0:
		// res.HardwareTemplateID stays the template's value: no trim
		// matched, so there is no candidate to defer to.
		res.Quality = MatchModelOnly
	case 1:
		m := matches[0]
		res.Quality = MatchExact
		res.Trim = m.trim.Name
		res.MatchedBy = m.matchedBy
		for k, v := range m.trim.Attributes {
			res.Attributes[k] = v
		}
		if m.trim.HardwareTemplateID != "" {
			res.HardwareTemplateID = m.trim.HardwareTemplateID
		}
	default:
		// res.HardwareTemplateID stays the template's value here too. Do
		// NOT "improve" this by taking the first (or any) candidate's
		// HardwareTemplateID: we have explicitly declined to pick a
		// candidate trim, and doing so anyway -- even for just this field --
		// reintroduces arbitrary selection in the one field that decides
		// what hardware DIMO ships. The template's value is what every
		// candidate shares by construction, so it's the only answer that
		// isn't a guess.
		res.Quality = MatchAmbiguous
		res.Candidates = make([]string, len(matches))
		for i, m := range matches {
			res.Candidates[i] = m.trim.Name
		}
		for k, v := range agreedAttributes(matches) {
			res.Attributes[k] = v
		}
	}

	return res
}

type trimMatch struct {
	trim      coremodels.Trim
	matchedBy []string
}

// selectorsMatch reports whether every selector sel declares agrees with
// sig, and which selector kinds matched. A selector that sel does not
// declare (empty slice / empty string) never blocks a match; a selector sel
// does declare only matches when sig supplies the corresponding signal and
// it agrees.
func selectorsMatch(sel coremodels.TrimSelectors, sig MatchSignals) (bool, []string) {
	var matchedBy []string

	if len(sel.ManufacturerCode) > 0 {
		if sig.ManufacturerCode == "" || !containsExact(sel.ManufacturerCode, sig.ManufacturerCode) {
			return false, nil
		}
		matchedBy = append(matchedBy, "manufacturerCode")
	}

	if len(sel.StyleName) > 0 {
		if sig.StyleName == "" || !containsFold(sel.StyleName, sig.StyleName) {
			return false, nil
		}
		matchedBy = append(matchedBy, "styleName")
	}

	if sel.VINPattern != "" {
		if sig.VIN == "" || !vinMatchesPattern(sel.VINPattern, sig.VIN) {
			return false, nil
		}
		matchedBy = append(matchedBy, "vinPattern")
	}

	return true, matchedBy
}

func containsExact(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func containsFold(values []string, want string) bool {
	for _, v := range values {
		if strings.EqualFold(v, want) {
			return true
		}
	}
	return false
}

// vinMatchesPattern treats a selector's vinPattern as an anchored regular
// expression over the full 17-character VIN. It is anchored here, not left
// to the pattern's author: under the open-contribution model these
// selectors are headed for, vinPattern is contributor-authored and the
// least-reviewed input in this path. An unanchored pattern matches anywhere
// in the string, so a pattern meant to identify one trim by a VIN prefix
// would silently claim every VIN that merely contains it as a substring.
// Anchoring here means correctness doesn't depend on every contributor
// remembering to anchor their own pattern.
//
// A malformed pattern is data the matcher cannot trust, so it fails safe
// (no match) rather than panicking -- consistent with MatchTrim never
// panicking on odd template data. Go's regexp is RE2: a hostile pattern is
// a compile error or a linear-time match, never catastrophic backtracking.
func vinMatchesPattern(pattern, vin string) bool {
	re, err := regexp.Compile(`^(?:` + pattern + `)$`)
	if err != nil {
		return false
	}
	return re.MatchString(vin)
}

// copyAttributes returns an independent copy of base, never nil, so callers
// can freely layer trim attributes on top without mutating the template.
func copyAttributes(base map[string]any) map[string]any {
	out := make(map[string]any, len(base))
	for k, v := range base {
		out[k] = v
	}
	return out
}

// agreedAttributes returns only the attributes on which every candidate
// trim agrees, both on the key being present and its value being equal.
func agreedAttributes(matches []trimMatch) map[string]any {
	if len(matches) == 0 {
		return nil
	}
	agreed := make(map[string]any)
	for k, v := range matches[0].trim.Attributes {
		agree := true
		for _, m := range matches[1:] {
			other, ok := m.trim.Attributes[k]
			if !ok || !reflect.DeepEqual(other, v) {
				agree = false
				break
			}
		}
		if agree {
			agreed[k] = v
		}
	}
	return agreed
}
