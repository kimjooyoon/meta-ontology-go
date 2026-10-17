package bodyexecution

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/textbound"
)

func TestDecodedTextFieldBoundsFresh1905(t *testing.T) {
	for _, tc := range []struct {
		name  string
		bound textbound.UTF8
		value string
		pass  bool
	}{
		{"legacy-at", 0, strings.Repeat("x", 1024), true},
		{"legacy-above", 0, strings.Repeat("x", 1025), false},
		{"expanded-at", textbound.New(1030), strings.Repeat("x", 1030), true},
		{"expanded-above", textbound.New(1030), strings.Repeat("x", 1031), false},
		{"utf8-at", textbound.New(1030), strings.Repeat("한", 343) + "x", true},
		{"utf8-above", textbound.New(1030), strings.Repeat("한", 343) + "xx", false},
		{"nul-escape-at", textbound.New(1030), strings.Repeat("\x00", 1030), true},
		{"nul-escape-above", textbound.New(1030), strings.Repeat("\x00", 1031), false},
		{"zero-empty", textbound.New(0), "", true},
		{"zero-nonempty", textbound.New(0), "x", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := bodycodegen.RecordType{Name: "Box", ID: "bound1905://box", Fields: []bodycodegen.RecordField{
				{Name: "value", ID: "bound1905://box/value", TypeID: "urn:gooo:type:string", TextBound: tc.bound},
			}}
			raw, err := json.Marshal(map[string]string{"value": tc.value})
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := canonicalRecord(raw, record)
			if (err == nil) != tc.pass {
				t.Fatal("decoded byte contract disagreed", len(tc.value), err)
			}
			if !tc.pass {
				return
			}
			var actual map[string]string
			if err := json.Unmarshal(canonical, &actual); err != nil || actual["value"] != tc.value {
				t.Fatal("accepted Text changed or was truncated", actual, err)
			}
		})
	}
}

func TestBoundedTextCollectionsFresh1905(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		pass bool
	}{
		{"empty-array", `[]`, true},
		{"two-empty-elements", `["",""]`, true},
		{"nonempty-element", `["","x"]`, false},
		{"null-array", `null`, false},
		{"null-element", `[null]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := canonicalRecordFieldWithBound([]byte(tc.raw), "Text", "many", textbound.New(0))
			if (err == nil) != tc.pass {
				t.Fatal("per-element bound or cardinality changed", tc.raw, err)
			}
		})
	}
	if _, err := canonicalRecordFieldWithBound([]byte(`[]`), "Boolean", "many", textbound.New(0)); err == nil {
		t.Fatal("empty collection hid a bound on the wrong scalar type")
	}
}

func TestBoundDoesNotWidenRepeatedCountFresh1905(t *testing.T) {
	for _, count := range []int{256, 257} {
		raw, _ := json.Marshal(make([]string, count))
		_, err := canonicalRecordFieldWithBound(raw, "Text", "many", textbound.New(0))
		if (err == nil) != (count == 256) {
			t.Fatal("element-byte bound changed the independent array budget", count, err)
		}
	}
}

func TestBoundedRecordCannotAdmitUndeclaredFieldsFresh1905(t *testing.T) {
	record := bodycodegen.RecordType{Name: "Box", Fields: []bodycodegen.RecordField{{
		Name: "value", TypeID: "urn:gooo:type:string", TextBound: textbound.New(1030),
	}}}
	if _, err := canonicalRecord([]byte(`{"value":"valid","extra":""}`), record); err == nil ||
		!strings.Contains(err.Error(), "undeclared field") {
		t.Fatal("source byte domain admitted a new field", err)
	}
}
