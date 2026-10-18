package toolchainrelease

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// This independent oracle uses a regular expression, not the source's byte loop.
var packageLabelReference = regexp.MustCompile(`\A[a-z][a-z0-9-]{0,63}\z`)

func collectionMapReference(raw []byte) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || len(fields) != 1 || len(fields["values"]) == 0 {
		return nil, fmt.Errorf("collection map query requires only values")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(fields["values"], &values); err != nil || values == nil || len(values) > 256 {
		return nil, fmt.Errorf("collection map requires an array of at most 256 Text values")
	}
	result := make([]bool, len(values))
	for i, rawValue := range values {
		var value *string
		if err := json.Unmarshal(rawValue, &value); err != nil || value == nil {
			return nil, fmt.Errorf("collection map value %d requires Text", i)
		}
		result[i] = packageLabelReference.MatchString(*value)
	}
	return json.Marshal(struct {
		Values []bool `json:"values"`
	}{result})
}
