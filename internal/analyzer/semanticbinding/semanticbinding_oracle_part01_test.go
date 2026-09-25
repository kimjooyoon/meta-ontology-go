package semanticbinding

import (
	"errors"
	"testing"
)

type fixtureObservation struct {
	result          Result
	records         []fixtureRecord
	oracleCanonical string
}

func assertFixture(t *testing.T, name string) fixtureObservation {
	t.Helper()
	source, want := loadFixture(t, name)
	result, err := Extract(Input{Sources: []SourceFile{{
		Filename: name + ".go", PackagePath: "billing", Source: source,
	}}})
	if want.Status == "accepted" {
		assertAcceptedFixture(t, result, err)
	} else {
		assertRejectedFixture(t, result, err, want.Diagnostic)
		return fixtureObservation{result: result}
	}
	records := recordsForResult(t, name, source, result)
	if len(records) != len(want.Records) {
		t.Fatalf("records = %#v, want oracle records %#v", records, want.Records)
	}
	for index := range want.Records {
		assertRecord(t, index, records[index], want.Records[index])
	}
	oracleCanonical := canonicalFixture(want.Records)
	if oracleCanonical != want.Canonical {
		t.Fatalf("oracle canonical = %q, want literal %q", oracleCanonical, want.Canonical)
	}
	if result.Digest == "" || result.Digest != result.CanonicalDigest || result.Digest != result.StableHash() {
		t.Fatalf("result digest fields are inconsistent: %#v", result)
	}
	return fixtureObservation{result: result, records: records, oracleCanonical: oracleCanonical}
}
func assertAcceptedFixture(t *testing.T, result Result, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if result.Status != StatusBound || result.FullSuiteFallback || len(result.Unknowns) != 0 {
		t.Fatalf("result = %#v, want complete BOUND result", result)
	}
}
func assertRejectedFixture(t *testing.T, result Result, err error, diagnostic string) {
	t.Helper()
	if err == nil {
		t.Fatal("Extract accepted a rejected oracle fixture")
	}
	if result.Status != StatusUnknown || !result.FullSuiteFallback || len(result.Unknowns) != 1 {
		t.Fatalf("result = %#v, want UNKNOWN with full-suite fallback", result)
	}
	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatalf("Extract error = %T %v, want *Error", err, err)
	}
	if typed.Code != oracleCode(diagnostic) {
		t.Fatalf("diagnostic code = %q, want oracle %q", typed.Code, diagnostic)
	}
	if result.Unknowns[0].Code != typed.Code || result.Unknowns[0].Span != typed.Span ||
		!result.Unknowns[0].FullSuiteFallback {
		t.Fatalf("UNKNOWN evidence = %#v, want the source-backed error", result.Unknowns[0])
	}
	if len(result.Bindings) != 0 || len(result.Obligations) != 0 {
		t.Fatalf("rejected result retained partial records: %#v", result)
	}
}
