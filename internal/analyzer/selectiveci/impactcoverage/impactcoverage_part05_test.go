package impactcoverage

import (
	"reflect"
	"strings"
	"testing"
)

func duplicateBaseSource(t *testing.T, input Input) []byte {
	t.Helper()
	data, err := input.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	base, err := input.Base.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	baseText := string(base)
	start := strings.Index(baseText, `,"sources":[`) + len(`,"sources":[`)
	end := strings.Index(baseText[start:], `],"digest":`) + start
	record := baseText[start:end]
	duplicate := strings.Replace(baseText, `],"digest":`, `,`+record+`],"digest":`, 1)
	return []byte(strings.Replace(string(data), baseText, duplicate, 1))
}
func TestOutputJSONRoundTripAndExpectedLabelIsolation(t *testing.T) {
	base := snap(t, "map", "reg", boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a"))
	head := snap(t, "map", "reg", boundSource("pkg/a.go", "a-2", "urn:gooo:entity:a"))
	fixture := struct {
		Input Input
		Label string
	}{NewInput(&base, &head), "expected:COMPLETE"}
	first := Observe(fixture.Input)
	firstEncoded, err := EncodeJSON(first)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeJSON(firstEncoded)
	if err != nil || !reflect.DeepEqual(first, decoded) {
		t.Fatalf("output round trip = %#v/%v", decoded, err)
	}
	fixture.Label = "expected:UNKNOWN"
	second := Observe(fixture.Input)
	if first.InputDigest != second.InputDigest || first.OutputDigest != second.OutputDigest ||
		!reflect.DeepEqual(first, second) {
		t.Fatalf("expected label changed observation: %#v/%#v", first, second)
	}
}
