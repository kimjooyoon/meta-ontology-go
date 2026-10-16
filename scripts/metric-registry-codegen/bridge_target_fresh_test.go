package main

import (
	"path/filepath"
	"testing"
)

func TestStableTargetIdentityFresh1701(t *testing.T) {
	for _, tc := range []struct{ name, source, reason string }{
		{"missing-tag", "type Row struct { Value string }", "target requires one named scalar and a gooo ID tag"},
		{"empty-ID", "type Row struct { Value string `gooo:\"\"` }", "target field Value requires a gooo ID tag"},
		{"missing-ID", "type Row struct { Value string `json:\"value\"` }", "target field Value requires a gooo ID tag"},
		{"duplicate-ID", "type Row struct { A string `gooo:\"same\"`; B string `gooo:\"same\"` }",
			"target field identity is ambiguous: \"same\""},
		{"duplicate-native-name", "type Row struct { A string `gooo:\"one\"`; A string `gooo:\"two\"` }",
			"target field identity is ambiguous: \"two\""},
		{"grouped", "type Row struct { A, B string `gooo:\"one\"` }", "target requires one named scalar and a gooo ID tag"},
		{"nested", "type Row struct { Value []string `gooo:\"one\"` }", "target requires one named scalar and a gooo ID tag"},
		{"alias", "type Row = struct { Value string `gooo:\"one\"` }", "target type is ambiguous or not a declared record"},
		{"generic", "type Row[T any] struct { Value string `gooo:\"one\"` }", "target type is ambiguous or not a declared record"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "row.go")
			writeBridgeFixtureFresh1699(t, path, "package fixture; "+tc.source)
			if _, err := targetFields(path, "Row"); err == nil || err.Error() != tc.reason {
				t.Fatalf("expected %q, got %v", tc.reason, err)
			}
		})
	}
}
