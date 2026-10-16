package languagesyntax

import (
	"encoding/json"
	"testing"
)

func TestMetaSourceHistoryBoundaryFresh1862(t *testing.T) {
	prior := []string{"source-A", "source-B"}
	for _, item := range []struct {
		name    string
		current []string
		accept  bool
	}{
		{"append", []string{"source-A", "source-B", "source-C"}, true},
		{"reorder", []string{"source-B", "source-A"}, false},
		{"remove", []string{"source-A"}, false},
		{"replace", []string{"source-A", "source-C"}, false},
		{"insert", []string{"source-C", "source-A", "source-B"}, false},
	} {
		t.Run(item.name, func(t *testing.T) {
			if metaSourceHistoryRetained67143(prior, item.current) != item.accept {
				t.Fatal("historical meta source identities lost their append boundary")
			}
		})
	}
	t.Run("append-needs-current-source-binding", func(t *testing.T) {
		current := expectedRegistry()
		current.MetaSources = append(append([]string{}, current.MetaSources...), "undeclared-meta-source.gooo")
		raw, err := json.Marshal(current)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodeRegistry(raw); err == nil {
			t.Fatal("arbitrary appended meta source bypassed current native/JSON binding")
		}
	})
}
