package main

import (
	"path/filepath"
	"strconv"
	"testing"
)

func TestNativeTagSyntaxFresh1709(t *testing.T) {
	for _, tc := range []struct{ name, tags, id, reason string }{
		{"plain", `gooo:"urn:value"`, "urn:value", ""},
		{"other-key", `json:"va:lue,omitempty" gooo:"urn:value"`, "urn:value", ""},
		{"spaces", `   gooo:"urn:value"  json:"value"  `, "urn:value", ""},
		{"escaped-ID", `gooo:"urn:\u0061"`, "urn:a", ""},
		{"no-separator", `gooo:"urn:value"json:"value"`, "", "tag keys require a space"},
		{"tab-separator", "gooo:\"urn:value\"\tjson:\"value\"", "", "tag keys require a space"},
		{"unclosed-value", `gooo:"urn:value`, "", "invalid quoted tag value"},
		{"unquoted-value", `gooo:urn:value`, "", "invalid quoted tag value"},
		{"control-key", "bad\tkey:\"value\" gooo:\"urn:value\"", "", "invalid tag key/value"},
		{"missing-ID", `json:"value"`, "", ""},
		{"empty-ID", `gooo:""`, "", ""},
		{"repeated-other-key", `json:"a" gooo:"urn:value" json:"b"`, "", `repeated key "json"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "row.go")
			writeBridgeFixtureFresh1699(t, path, "package fixture; type Row struct { Value string "+strconv.Quote(tc.tags)+" }")
			fields, err := targetFields(path, "Row")
			if tc.id == "" {
				reason := "target field Value requires a gooo ID tag"
				if tc.reason != "" {
					reason = "target field Value tag: " + tc.reason
				}
				if err == nil || err.Error() != reason {
					t.Fatalf("expected %q, got %v", reason, err)
				}
				return
			}
			if err != nil || len(fields) != 1 || fields[tc.id] != (targetField{name: "Value", kind: "string"}) {
				t.Fatalf("valid metadata changed its typed ID: %v, %v", fields, err)
			}
		})
	}
}
