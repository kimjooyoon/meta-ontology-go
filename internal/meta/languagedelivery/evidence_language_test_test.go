package languagedelivery

import (
	"encoding/json"
	"testing"
)

func TestCanonicalLanguageTestUsesExternalReceipt(t *testing.T) {
	for _, obligation := range CanonicalContract().Obligations {
		if obligation.ID != "USER-LANGUAGE-TEST" {
			continue
		}
		if obligation.Evidence.Source != SourceTest || obligation.Evidence.Kind != EvidenceTest || obligation.Evidence.Counter != "passed_tests" || obligation.Evidence.Target != 2 {
			t.Fatalf("language test evidence = %#v", obligation.Evidence)
		}
		return
	}
	t.Fatal("USER-LANGUAGE-TEST obligation missing")
}

func TestLanguageTestUnknownTopLowersResolution(t *testing.T) {
	receipt := exactLanguageTestReceipt("head")
	receipt.Decision = "FAIL_CLOSED"
	receipt.Resolution = "LOWER_RESOLUTION"
	receipt.Summary.Unknowns = 1
	observation := inspectLanguageTest(marshalLanguageTest(t, receipt), "head", &LanguageTestReceipt{}, ManifestEntry{})
	if observation.State != "UNKNOWN" || observation.Reason != "LANGUAGE_TEST_RECEIPT_UNKNOWN" {
		t.Fatalf("observation = %#v", observation)
	}
}

func TestLanguageTestRejectsGoodLookingPartialBoundary(t *testing.T) {
	receipt := exactLanguageTestReceipt("head")
	receipt.Summary.MissingTestRejections = 0
	observation := inspectLanguageTest(marshalLanguageTest(t, receipt), "head", &LanguageTestReceipt{}, ManifestEntry{})
	if observation.State != "FAIL" || observation.Reason != "LANGUAGE_TEST_BOUNDARY_NOT_EXACT" {
		t.Fatalf("observation = %#v", observation)
	}
}

func TestObserveRuleDispatchesLanguageTestCounter(t *testing.T) {
	decoded := decodedEvidence{LanguageTest: exactLanguageTestReceipt("head")}
	got, reason := observeRule(EvidenceRule{Kind: EvidenceTest, Counter: "passed_tests"}, decoded)
	if got != 2 || reason != "LANGUAGE_TESTS_PASSED" {
		t.Fatalf("observeRule() = (%d, %q)", got, reason)
	}
}

func exactLanguageTestReceipt(head string) LanguageTestReceipt {
	return LanguageTestReceipt{Schema: languageTestReportSchema, SubjectSHA: head, Decision: "PASS", Resolution: "EXACT", Summary: LanguageTestSummary{Coordinates: LanguageTestCoordinates{Satisfied: 12, Total: 12, BasisPoints: 10000}, DeclaredTests: 2, ExecutedTests: 2, PassedTests: 2, ReceiptDigestVariants: 1, ExecutionDigestVariants: 1, AssertionRejections: 1, MissingTestRejections: 1, NonClaims: 3, Compiler: LanguageTestCompiler{ExecutableDigest: "sha256:88fdc432c8cfec498cf58a4bc2a1072439dae099fcbc03412403ed78feeff26d", Go127Runtimes: 2}}, Views: []LanguageTestView{{Audience: "USER", Resolution: "USER_VISIBLE", Satisfied: 4, Total: 4, BasisPoints: 10000}, {Audience: "TOOL_AUTHOR", Resolution: "TOOL_CONTRACT", Satisfied: 8, Total: 8, BasisPoints: 10000}, {Audience: "GOVERNOR", Resolution: "FULL_RECEIPT", Satisfied: 12, Total: 12, BasisPoints: 10000}}}
}

func marshalLanguageTest(t *testing.T, receipt LanguageTestReceipt) []byte {
	t.Helper()
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
