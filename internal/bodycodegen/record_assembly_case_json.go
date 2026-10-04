package bodycodegen

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// Saved JSON may be indented or reorder object keys without changing a typed
// case. Normalize its value carriers before comparing the recorded observation.
func canonicalRecordCaseValues(cases []RecordAssemblyCase) ([]RecordAssemblyCase, error) {
	values := append([]RecordAssemblyCase(nil), cases...)
	for i := range values {
		for _, raw := range []*json.RawMessage{&values[i].Inputs, &values[i].Expected, &values[i].Actual} {
			text, err := assemblyspec.CanonicalValue(string(*raw))
			if err != nil {
				return nil, err
			}
			*raw = json.RawMessage(text)
		}
	}
	return values, nil
}
