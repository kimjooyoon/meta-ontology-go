package main

import "testing"

func TestObservedAdmissionPlanFresh1894(t *testing.T) {
	want := []countercase{
		{"wrong-head", "OBSERVATION_DOCUMENT", "rebuild observed concept-operation receipt: operation observation source identity is not exact", false},
		{"wrong-program", "OBSERVATION_DOCUMENT", "rebuild observed concept-operation receipt: operation observation source identity is not exact", false},
		{"wrong-cohort", "OBSERVATION_DOCUMENT", "rebuild observed concept-operation receipt: operation observation source identity is not exact", false},
		{"known-binding", "OBSERVATION_DOCUMENT", "rebuild observed concept-operation receipt: observed binding differs from source operation \"urn:admission:declared\"", false},
		{"outside-binding", "OBSERVATION_DOCUMENT", "rebuild observed concept-operation receipt: outside-cohort observation cannot claim an executable binding", false},
		{"repository-write-authority", "OBSERVED_RECEIPT", "observed concept-operation receipt identity is invalid", false},
		{"promotion-authority", "OBSERVED_RECEIPT", "observed concept-operation receipt identity is invalid", false},
		{"wrong-digest", "OBSERVED_RECEIPT", "observed concept-operation receipt does not match source replay", false},
		{"resealed-source-value", "OBSERVED_RECEIPT", "observed concept-operation receipt does not match source replay", false},
		{"policy-bytes", "POLICY_BYTES", "observation policy differs from repository source", false},
		{"document-bytes", "DOCUMENT_BYTES", "observed concept-operation receipt does not match source replay", false},
		{"missing-observation-inputs", "MISSING_OBSERVATION_INPUTS", "observed concept-operation receipt identity is invalid", false},
		{"legacy-downgrade", "LEGACY_DOWNGRADE", "verify concept-operation evidence: observations require their own receipt", false},
	}
	got := observedAdmissionPlans("urn:admission:declared")
	if len(got) != 13 || len(got) != len(want) || len(countercases()) != 16 || len(observedPlans1889(2)) != 6 {
		t.Fatal("admission controls changed another case population")
	}
	for index, row := range want {
		t.Run(row.ID, func(t *testing.T) {
			if got[index] != row {
				t.Fatalf("declared admission contract differs: got %+v, want %+v", got[index], row)
			}
		})
	}
}
