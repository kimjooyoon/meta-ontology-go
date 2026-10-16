package main

import "fmt"

const admissionIdentityReason = "observed concept-operation receipt identity is invalid"

const admissionReplayReason = "observed concept-operation receipt does not match source replay"

const admissionHeaderReason = "rebuild observed concept-operation receipt: " +
	"operation observation source identity is not exact"

// These controls belong to the observed receipt's admission contract. The
// existing 16 receipt controls and six source-classification cases stay separate.
func observedAdmissionPlans(sourceOperation string) []countercase {
	binding := fmt.Sprintf("rebuild observed concept-operation receipt: observed binding differs from source operation %q", sourceOperation)
	outside := "rebuild observed concept-operation receipt: " +
		"outside-cohort observation cannot claim an executable binding"
	return []countercase{
		{"wrong-head", "OBSERVATION_DOCUMENT", admissionHeaderReason, false},
		{"wrong-program", "OBSERVATION_DOCUMENT", admissionHeaderReason, false},
		{"wrong-cohort", "OBSERVATION_DOCUMENT", admissionHeaderReason, false},
		{"known-binding", "OBSERVATION_DOCUMENT", binding, false},
		{"outside-binding", "OBSERVATION_DOCUMENT", outside, false},
		{"repository-write-authority", "OBSERVED_RECEIPT", admissionIdentityReason, false},
		{"promotion-authority", "OBSERVED_RECEIPT", admissionIdentityReason, false},
		{"wrong-digest", "OBSERVED_RECEIPT", admissionReplayReason, false},
		{"resealed-source-value", "OBSERVED_RECEIPT", admissionReplayReason, false},
		{"policy-bytes", "POLICY_BYTES", "observation policy differs from repository source", false},
		{"document-bytes", "DOCUMENT_BYTES", admissionReplayReason, false},
		{"missing-observation-inputs", "MISSING_OBSERVATION_INPUTS", admissionIdentityReason, false},
		{"legacy-downgrade", "LEGACY_DOWNGRADE", evidenceReasonPrefix + "observations require their own receipt", false},
	}
}
