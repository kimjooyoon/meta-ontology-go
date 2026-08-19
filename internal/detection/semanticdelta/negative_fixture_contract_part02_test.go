package semanticdelta

import (
	"bytes"
	"errors"
	"testing"
)

func loadNegativeFixture(t *testing.T, name string) Request {
	t.Helper()
	data, err := negativeFixtureFiles.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	request, err := Decode(data)
	if err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return request
}
func assertNegativeReportFixture(t *testing.T, input string, report Report, expected string) {
	t.Helper()
	actual, err := EncodeReportJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	want, err := negativeFixtureFiles.ReadFile("testdata/" + expected)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, want) {
		t.Fatalf("report for %s is not canonical:\nactual=%s\nwant=%s", input, actual, want)
	}
	var scopeErr *ScopeError
	if _, err := Evaluate(bytes.NewReader(mustNegativeInput(t, input)), bytes.NewBuffer(nil), FormatJSON); !errors.As(err, &scopeErr) {
		t.Fatalf("Evaluate(%s) error = %v, want ScopeError", input, err)
	}
}
func mustNegativeInput(t *testing.T, name string) []byte {
	t.Helper()
	data, err := negativeFixtureFiles.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
