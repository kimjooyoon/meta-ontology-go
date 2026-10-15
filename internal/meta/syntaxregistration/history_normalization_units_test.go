package syntaxregistration

import (
	"maps"
	"reflect"
	"testing"
)

func historyNormalizationUnits(t *testing.T, source *goSource, history historyOrigins, changes map[string][]byte) map[string][]byte {
	t.Helper()
	root, _ := source.location(history.root.Pos())
	owners := map[string]bool{root: true}
	for _, helper := range history.helpers {
		path, _ := source.location(helper.Pos())
		owners[path] = true
	}
	if !reflect.DeepEqual(sortedPaths(owners), sortedPaths(changes)) {
		t.Fatalf("normalization owners: want %v, got %v", sortedPaths(owners), sortedPaths(changes))
	}
	updated := maps.Clone(source.units)
	for path, raw := range changes {
		if _, exists := source.units[path]; !exists {
			t.Fatal("normalization introduced an unowned source unit", path)
		}
		updated[path] = raw
	}
	return updated
}
