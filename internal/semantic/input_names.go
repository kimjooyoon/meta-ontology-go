package semantic

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/inputbinding"
	"strings"
)

// InputName is an authored value binding at one positional input.
// Its spelling participates in body binding; it does not create a role ID.
type InputName struct {
	Name string `json:"name"`
	Span Span   `json:"span"`
}

func normalizeInputNames(kind Kind, sequence []ID, names []InputName) ([]InputName, error) {
	if len(names) == 0 {
		return nil, nil
	}
	if kind != Activity {
		return nil, fmt.Errorf("%w: input names require an Activity", ErrInvalidNode)
	}
	arity := len(sequence)
	if arity == 0 {
		arity = 1
	}
	declared := make([]string, len(names))
	result := append([]InputName(nil), names...)
	for i := range result {
		declared[i] = result[i].Name
		result[i].Span = result[i].Span.Normalized()
		if err := result[i].Span.Validate(); err != nil {
			return nil, err
		}
	}
	if _, err := inputbinding.Resolve(arity, declared); err != nil {
		return nil, err
	}
	return result, nil
}

func writeCanonicalInputNames(b *strings.Builder, names []InputName, includeSpans bool) {
	if len(names) == 0 {
		return
	}
	writeCanonicalField(b, "input-names")
	for _, name := range names {
		writeCanonicalField(b, name.Name)
		if includeSpans {
			writeCanonicalSpan(b, name.Span)
		}
	}
	writeCanonicalField(b, "end-input-names")
}
