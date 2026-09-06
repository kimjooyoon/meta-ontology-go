package languagesyntax_test

import (
	"bytes"
	"encoding/json"
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
	assertSourceInventory(t, repository, report.Source, report.Summary.GoooLines)
	if err := languagesyntax.Validate(report, testHead); err != nil {
		t.Fatal(err)
	}
	if report.Decision != languagesyntax.DecisionPass || report.Resolution != languagesyntax.ResolutionExact ||
		report.Summary.Satisfied != 60 || report.Summary.ValidCases != 57 ||
		report.Summary.InvalidCases != 3 || report.Summary.Unresolved != 0 ||
		report.Summary.CapabilitySatisfied != languagesyntax.FixedCapabilityTotal ||
		report.Summary.CapabilityTotal != languagesyntax.FixedCapabilityTotal ||
		report.Summary.CapabilityExecuted != languagesyntax.FixedCapabilityTotal ||
		report.Summary.CapabilityUnresolved != 0 ||
		report.Summary.GovernanceSatisfied != languagesyntax.FixedGovernanceTotal ||
		report.Summary.GovernanceTotal != languagesyntax.FixedGovernanceTotal ||
		report.Summary.GovernanceExecuted != languagesyntax.FixedGovernanceTotal ||
		report.Summary.GovernanceUnresolved != 0 ||
		len(report.Source.GoooFiles) != 69 || len(report.Source.PackageUnits) != 4 ||
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
			report.Summary.Executed != 0 || report.Summary.Unresolved != 60 {
			t.Fatalf("unknown registry was not lowered: %#v", report)
		}
	}
}

func TestScopePartitionUsesFixedDenominatorsAndRejectsDrift(t *testing.T) {
	if languagesyntax.FixedTotal != 60 || languagesyntax.FixedCapabilityTotal != 58 ||
		languagesyntax.FixedGovernanceTotal != 2 ||
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

func TestRecordBindingRetainsEverySyntaxReplayObligation(t *testing.T) {
	repository, raw := fixture(t)
	report := languagesyntax.Evaluate(repository, testHead, raw, languageconcept.BuildArtifact(repository))
	found := 0
	for _, item := range report.Cases {
		if item.Definition.ID != "language-record-binding" {
			continue
		}
		found++
		if item.Definition.Path != "examples/language-record-binding/main.gooo" ||
			!item.Definition.EntityFields || item.Definition.Scope != languagesyntax.ScopeLanguageCapability ||
			item.Definition.Kind != languagesyntax.KindValid || item.Status != "SATISFIED" ||
			!item.Evidence.ASTReplayed || !item.Evidence.ByteReplayed || !item.Evidence.SemanticReplayed ||
			!item.Evidence.GetPut || !item.Evidence.PutGet {
			t.Fatalf("record syntax obligations not closed: %#v", item)
		}
	}
	if found != 1 {
		t.Fatalf("record syntax case count=%d want=1", found)
	}
}

func TestRemovingRecordBindingCannotLowerCorpusDenominator(t *testing.T) {
	repository, raw := fixture(t)
	var registry languagesyntax.Registry
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatal(err)
	}
	removed := 0
	for index, item := range registry.Cases {
		if item.ID == "language-record-binding" {
			registry.Cases = append(registry.Cases[:index], registry.Cases[index+1:]...)
			removed++
			break
		}
	}
	if removed != 1 {
		t.Fatal("record case missing from canonical fixture")
	}
	mutated, err := json.Marshal(registry)
	if err != nil {
		t.Fatal(err)
	}
	report := languagesyntax.Evaluate(repository, testHead, mutated, languageconcept.BuildArtifact(repository))
	if report.Decision != languagesyntax.DecisionClosed || report.Resolution != languagesyntax.ResolutionLower ||
		report.Summary.Total != 60 || report.Summary.Executed != 0 || report.Summary.Unresolved != 60 {
		t.Fatalf("removing a case reduced proof obligations: %#v", report.Summary)
	}
}
