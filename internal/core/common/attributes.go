package common

import (
	"fmt"
	"sort"
	"strconv"
)

// AttributeString renders a typed template attribute for the legacy
// string-valued shapes the API still serves.
//
// Template attributes decode into map[string]any, so every number arrives as a
// float64. strconv rather than fmt so 15.8 renders "15.8" and not "1.58e+01",
// and a price of 1250000 renders "1250000" and not "1.25e+06". Both the
// definition path and the device-style path call this, so one definition's
// attributes read the same whichever endpoint served them.
func AttributeString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	default:
		return fmt.Sprint(v)
	}
}

// SortedAttributeNames returns a template's attribute names in a stable order.
//
// Ranging over the map directly puts the attributes in a different order on
// every call, so two requests for the same definition answered with the same
// attributes in a different arrangement.
func SortedAttributeNames(attributes map[string]any) []string {
	names := make([]string, 0, len(attributes))
	for name := range attributes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
