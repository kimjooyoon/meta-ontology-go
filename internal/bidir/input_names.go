package bidir

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/inputbinding"
)

func validateInputNames(kind Kind, arity int, names []Reference) error {
	if len(names) == 0 {
		return nil
	}
	if kind != ActivityKind {
		return fmt.Errorf("input names require an Activity")
	}
	declared := make([]string, len(names))
	for i, name := range names {
		if name.ID != "" || name.Namespace != "" {
			return fmt.Errorf("an input name is a binding, not an entity identity")
		}
		if err := name.Span.Validate(); err != nil {
			return err
		}
		declared[i] = name.Name
	}
	_, err := inputbinding.Resolve(arity, declared)
	return err
}

func equalInputNames(a, b []Reference) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name {
			return false
		}
	}
	return true
}
