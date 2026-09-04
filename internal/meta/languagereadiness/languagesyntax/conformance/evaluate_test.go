package languagesyntax_test

import (
	"bytes"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

const testHead = "0000000000000000000000000000000000000000"

const expectedInvalidCaseIDs = "unknown-keyword,unterminated-string,source-execution-invalid"

func fixture(t *testing.T) (fs.FS, []byte) {
	t.Helper()
	repository := os.DirFS("../../../../..")
	raw, err := fs.ReadFile(repository, "examples/language-syntax-roundtrip/corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	return repository, raw
}

func TestCompleteCorpusProvesSyntaxRoundTrip(t *testing.T) {
	repository, raw := fixture(t)
	report := languagesyntax.Evaluate(repository, testHead, raw, languageconcept.BuildArtifact(repository))
	if err := languagesyntax.Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != languagesyntax.DecisionPass || report.Resolution != languagesyntax.ResolutionExact ||
		report.Summary.Satisfied != 55 || report.Summary.ValidCases != 52 ||
		report.Summary.InvalidCases != 3 || report.Summary.Unresolved != 0 ||
		report.Summary.CapabilitySatisfied != languagesyntax.FixedCapabilityTotal ||
		report.Summary.CapabilityTotal != languagesyntax.FixedCapabilityTotal ||
		report.Summary.CapabilityExecuted != languagesyntax.FixedCapabilityTotal ||
		report.Summary.CapabilityUnresolved != 0 ||
		report.Summary.GovernanceSatisfied != languagesyntax.FixedGovernanceTotal ||
		report.Summary.GovernanceTotal != languagesyntax.FixedGovernanceTotal ||
		report.Summary.GovernanceExecuted != languagesyntax.FixedGovernanceTotal ||
		report.Summary.GovernanceUnresolved != 0 || report.Summary.GoooLines != 1370 ||
		len(report.Source.GoooFiles) != 62 || len(report.Source.PackageUnits) != 4 ||
		len(report.Source.PackageUnits[0].Members) != 2 || len(report.Source.PackageUnits[1].Members) != 3 ||
		len(report.Source.PackageUnits[2].Members) != 1 || len(report.Source.PackageUnits[3].Members) != 1 {
		invalidIDs := make([]string, 0, report.Summary.InvalidCases)
		for _, item := range report.Cases {
			if item.Definition.Kind == languagesyntax.KindInvalid {
				invalidIDs = append(invalidIDs, item.Definition.ID)
			}
		}
		t.Fatalf("report summary: decision=%q resolution=%q reason=%q gooo_lines=%d gooo_files=%d unregistered=%v missing=%v satisfied=%d valid=%d invalid=%d invalid_ids=%q expected_invalid_ids=%q unresolved=%d package_units=%d", report.Decision, report.Resolution, report.Reason, report.Summary.GoooLines, len(report.Source.GoooFiles), report.Source.UnregisteredGooo, report.Source.MissingRegistered, report.Summary.Satisfied, report.Summary.ValidCases, report.Summary.InvalidCases, strings.Join(invalidIDs, ","), expectedInvalidCaseIDs, report.Summary.Unresolved, len(report.Source.PackageUnits))
	}
	invalidIDs := make([]string, 0, report.Summary.InvalidCases)
	for _, item := range report.Cases {
		if item.Definition.Kind == languagesyntax.KindInvalid {
			invalidIDs = append(invalidIDs, item.Definition.ID)
		}
	}
	if strings.Join(invalidIDs, ",") != expectedInvalidCaseIDs {
		t.Fatalf("invalid corpus identities drifted: got=%q want=%q", strings.Join(invalidIDs, ","), expectedInvalidCaseIDs)
	}
}

func TestUnknownRegistryLowersResolution(t *testing.T) {
	repository, canonical := fixture(t)
	unknownField := bytes.Replace(canonical, []byte("{"), []byte(`{"unknown":true,`), 1)
	for _, raw := range [][]byte{[]byte(`{"schema":"UNKNOWN","cases":[]}`), unknownField} {
		report := languagesyntax.Evaluate(repository, testHead, raw, languageconcept.BuildArtifact(repository))
		if err := languagesyntax.Validate(report, testHead); err != nil {
			t.Fatal(err)
		}
		if report.Decision != languagesyntax.DecisionClosed || report.Resolution != languagesyntax.ResolutionLower ||
				report.Summary.Executed != 0 || report.Summary.Unresolved != 55 {
			t.Fatalf("unknown registry was not lowered: %#v", report)
		}
	}
}

func TestScopePartitionUsesFixedDenominatorsAndRejectsDrift(t *testing.T) {
	if languagesyntax.FixedTotal != 55 || languagesyntax.FixedCapabilityTotal != 54 ||
		languagesyntax.FixedGovernanceTotal != 1 ||
		languagesyntax.FixedCapabilityTotal+languagesyntax.FixedGovernanceTotal != languagesyntax.FixedTotal {
		t.Fatalf("scope denominators drifted: total=%d capability=%d governance=%d", languagesyntax.FixedTotal,
			languagesyntax.FixedCapabilityTotal, languagesyntax.FixedGovernanceTotal)
	}
	repository, canonical := fixture(t)
	cases := []struct {
		name string
		raw  []byte
	}{
		{"missing scope", bytes.Replace(canonical, []byte("      \"scope\": \"LANGUAGE_CAPABILITY\"\n"), nil, 1)},
		{"unknown scope", bytes.Replace(canonical, []byte(`"scope": "GOVERNANCE_OBSERVATION"`), []byte(`"scope": "UNKNOWN"`), 1)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			report := languagesyntax.Evaluate(repository, testHead, testCase.raw, languageconcept.BuildArtifact(repository))
			if report.Decision != languagesyntax.DecisionClosed || report.Resolution != languagesyntax.ResolutionLower ||
				report.Summary.Unresolved != languagesyntax.FixedTotal {
				t.Fatalf("scope drift was not rejected: %#v", report)
			}
		})
	}
}
