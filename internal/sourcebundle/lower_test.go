package sourcebundle

import (
	"reflect"
	"strings"
	"testing"
)

func TestLowerRejectsCrossMemberIdentityConflicts(t *testing.T) {
	const header = "package example\nnamespace example\n"
	first := Source{"a.gooo", header + "entity First id \"urn:example:first\"\n"}
	for _, declaration := range []string{
		"entity Second id \"urn:example:first\"", // same ID, different display name
		"entity First id \"urn:example:other\"",  // same name, different ID
		"activity Unknown(Missing) -> First",
	} {
		got, err := Lower([]Source{first, {"b.gooo", header + declaration + "\n"}})
		if err == nil || !reflect.DeepEqual(got, Bundle{}) {
			t.Fatalf("invalid source set produced a partial bundle: declaration=%q err=%v", declaration, err)
		}
	}
}

func TestLowerReportsThePhysicalParserLocation(t *testing.T) {
	_, err := Lower([]Source{{"a.gooo", "package example\nnamespace example\n"},
		{"later.gooo", "package example\nnamespace example\n?"}})
	if err == nil || !strings.Contains(err.Error(), "later.gooo:3:1") {
		t.Fatalf("parser location lost from later member: %v", err)
	}
}
