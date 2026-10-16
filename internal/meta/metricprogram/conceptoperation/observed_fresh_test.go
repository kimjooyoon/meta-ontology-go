package conceptoperation

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func observedFixture1886(t *testing.T) (receiptFixture, ObservationDocument) {
	t.Helper()
	f := buildReceiptFixture(t)
	policy, err := os.ReadFile(filepath.Join(f.repositoryRoot, ObservationPolicyPath))
	if err != nil {
		t.Fatal(err)
	}
	document, err := DeclareObservations(f.receipt, f.inputs.Strategy)
	if err != nil {
		t.Fatal(err)
	}
	f.inputs.Observation = &ObservationInputs{PolicySource: policy}
	return f, document
}

func TestObservedCohortKeepsOriginalProgramFresh1886(t *testing.T) {
	f, document := observedFixture1886(t)
	declaredDigest := document.DeclaredCohortDigest
	document.Observations = append(document.Observations, OperationObservation{Operation: "urn:unregistered:observation1886"})
	f.inputs.Observation.Document, _ = json.Marshal(document)
	receipt, err := BuildObserved(context.Background(), f.inputs)
	if err != nil {
		t.Fatal(err)
	}
	assessment, err := observedAssessment(receipt)
	if err != nil || assessment.ExpectedCount != f.receipt.ExpectedCount ||
		assessment.BoundCount != f.receipt.ExpectedCount || assessment.UnknownCount != 1 ||
		assessment.CoverageBPS != 10000 || assessment.Status != "FAIL_CLOSED" ||
		assessment.Reason != "CONCEPT_OPERATION_BINDING_AMBIGUOUS" ||
		receipt.DeclaredCohortDigest != declaredDigest || receipt.Baseline.Digest != f.receipt.Digest {
		t.Fatalf("unknown observation changed the declared denominator: %v %+v", err, receipt)
	}
	raw, _ := json.Marshal(receipt)
	state, err := VerifyEvidence(raw, f.inputs, f.repositoryFS, f.repositoryRoot, f.repository, f.subjectSHA)
	if err != nil || state.Status != "FAIL_CLOSED" || state.Digest != receipt.Digest {
		t.Fatal("source replay did not preserve full coverage and the unknown guard", state, err)
	}
}

func TestObservedCohortInputIdentityFresh1887(t *testing.T) {
	for _, mode := range []string{"head", "program", "cohort", "known-binding", "outside-binding", "null-rows"} {
		t.Run(mode, func(t *testing.T) {
			f, document := observedFixture1886(t)
			switch mode {
			case "head":
				document.SubjectSHA = strings.Repeat("e", 40)
			case "program":
				document.ProgramDigest = fixtureDigest("different-program1887")
			case "cohort":
				document.DeclaredCohortDigest = fixtureDigest("different-cohort1887")
			case "known-binding":
				document.Observations[0].Binding.Actual = "forged1887"
			case "outside-binding":
				document.Observations[0].Operation = "urn:unregistered:with-forged-binding1887"
			case "null-rows":
				document.Observations = nil
			}
			f.inputs.Observation.Document, _ = json.Marshal(document)
			if _, err := BuildObserved(context.Background(), f.inputs); err == nil {
				t.Fatal("invalid observed source was accepted", mode)
			}
		})
	}
}

func TestObservedReceiptCannotUpgradeOrIgnoreInputsFresh1888(t *testing.T) {
	f, document := observedFixture1886(t)
	f.inputs.Observation.Document, _ = json.Marshal(document)
	receipt, err := BuildObserved(context.Background(), f.inputs)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"receipt-tamper", "digest-tamper", "policy-tamper", "document-tamper", "drop-inputs", "legacy-downgrade"} {
		t.Run(mode, func(t *testing.T) {
			inputs, altered := f.inputs, receipt
			observation := *inputs.Observation
			inputs.Observation = &observation
			switch mode {
			case "receipt-tamper":
				altered.PromotionAuthorized = true
			case "digest-tamper":
				altered.Digest = fixtureDigest("forged-observed-receipt1888")
			case "policy-tamper":
				inputs.Observation.PolicySource = append([]byte("// changed1888\n"), observation.PolicySource...)
			case "document-tamper":
				inputs.Observation.Document = append(append([]byte(nil), observation.Document...), ' ')
			case "drop-inputs":
				inputs.Observation = nil
			}
			raw, _ := json.Marshal(altered)
			if mode == "legacy-downgrade" {
				raw, _ = json.Marshal(f.receipt)
			}
			if _, err := VerifyEvidence(raw, inputs, f.repositoryFS, f.repositoryRoot, f.repository, f.subjectSHA); err == nil {
				t.Fatal("changed input or receipt bypassed source replay", mode)
			}
		})
	}
}
