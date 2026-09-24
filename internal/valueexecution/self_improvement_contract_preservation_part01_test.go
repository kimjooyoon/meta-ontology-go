package valueexecution

import (
	"strings"
	"testing"
)

func TestObserveSelfImprovementContractPreservationAcceptsSameContract(t *testing.T) {
	observation := ObserveSelfImprovementContractPreservation(
		SelfImprovementContractPreservationInput{
			Scope:                   "COUNTEREXAMPLE_RECOVERY",
			CounterexampleRecovered: true,
			BaselineContractDigest:  "contract-a",
			CandidateContractDigest: "contract-a",
			CandidateDigest:         "candidate-a",
			RegressionEvidenceDigest: "regression-a",
			RegressionStatus:         "PASSED",
		},
	)
	if observation.Status != SelfImprovementContractPreservationAccepted {
		t.Fatalf("status = %q, want ACCEPTED", observation.Status)
	}
	if !observation.ContractPreserved || !observation.RegressionEvidenceSeen {
		t.Fatal("accepted observation must record contract and regression evidence")
	}
	if observation.AdoptionAuthorized {
		t.Fatal("evaluation must not authorize adoption")
	}
	if !observation.NonAuthorizing {
		t.Fatal("observation must remain non-authorizing")
	}
	if len(observation.Digest) != 64 || strings.Trim(observation.Digest, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want lowercase sha256", observation.Digest)
	}
}

func TestObserveSelfImprovementContractPreservationRejectsContractChange(t *testing.T) {
	observation := ObserveSelfImprovementContractPreservation(
		SelfImprovementContractPreservationInput{
			Scope:                   "COUNTEREXAMPLE_RECOVERY",
			CounterexampleRecovered: true,
			BaselineContractDigest:  "contract-a",
			CandidateContractDigest: "contract-b",
			CandidateDigest:         "candidate-b",
			RegressionEvidenceDigest: "regression-b",
			RegressionStatus:         "PASSED",
		},
	)
	if observation.Status != SelfImprovementContractPreservationRejected {
		t.Fatalf("status = %q, want REJECTED", observation.Status)
	}
	if observation.Reason != "CONTRACT_CHANGED_REQUIRES_EXPLICIT_SCOPE" {
		t.Fatalf("reason = %q, want explicit-scope rejection", observation.Reason)
	}
}

func TestObserveSelfImprovementContractPreservationKeepsMissingEvidenceUnknown(t *testing.T) {
	observation := ObserveSelfImprovementContractPreservation(
		SelfImprovementContractPreservationInput{
			Scope:                   "COUNTEREXAMPLE_RECOVERY",
			CounterexampleRecovered: true,
			BaselineContractDigest:  "contract-a",
			CandidateContractDigest: "contract-a",
			CandidateDigest:         "candidate-a",
			RegressionStatus:        "UNKNOWN",
		},
	)
	if observation.Status != SelfImprovementContractPreservationUnknown {
		t.Fatalf("status = %q, want UNKNOWN", observation.Status)
	}
	if observation.Reason != "REGRESSION_EVIDENCE_MISSING_OR_UNKNOWN" {
		t.Fatalf("reason = %q, want missing-evidence reason", observation.Reason)
	}
}