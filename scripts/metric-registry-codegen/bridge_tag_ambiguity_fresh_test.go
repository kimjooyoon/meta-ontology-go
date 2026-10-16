package main

import (
	"path/filepath"
	"strconv"
	"testing"
)

func TestAmbiguousNativeTagsFresh1707(t *testing.T) {
	for _, tc := range []struct{ name, tags, reason string }{
		{"different-IDs", `gooo:"first" gooo:"second"`, `target field Value tag: repeated key "gooo"`},
		{"repeated-ID", `gooo:"first" gooo:"first"`, `target field Value tag: repeated key "gooo"`},
		{"ignored-tail", `gooo:"first" dangling`, `target field Value tag: invalid tag key/value`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "row.go")
			writeBridgeFixtureFresh1699(t, path, "package fixture; type Row struct { Value string "+strconv.Quote(tc.tags)+" }")
			if _, err := targetFields(path, "Row"); err == nil || err.Error() != tc.reason {
				t.Fatalf("expected %q, got %v", tc.reason, err)
			}
		})
	}
}
