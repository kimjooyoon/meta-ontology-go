package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

func TestObservedAdmissionDocumentMutationFresh1895(t *testing.T) {
	_, baseline := admissionFixture1894(t)
	for _, id := range []string{"wrong-head", "wrong-program", "wrong-cohort", "known-binding", "outside-binding"} {
		t.Run(id, func(t *testing.T) {
			raw := baseline.Inputs.Observation.Document
			var want co.ObservationDocument
			if err := json.Unmarshal(raw, &want); err != nil {
				t.Fatal(err)
			}
			switch id {
			case "wrong-head":
				want.SubjectSHA = "0" + strings.Repeat("a", 39)
			case "wrong-program":
				want.ProgramDigest = artifact.ContentDigest([]byte("admission-other-program"))
			case "wrong-cohort":
				want.DeclaredCohortDigest = artifact.ContentDigest([]byte("admission-other-cohort"))
			case "known-binding":
				want.Observations[0].Binding.Actual = "original-actual-admission-changed"
			case "outside-binding":
				want.Observations[0].Operation = "urn:cohort:unregistered-observation-1889"
			}
			got, err := admissionDocument(raw, id)
			if err != nil || !bytes.Equal(got, marshalAdmission1894(t, want)) || bytes.Equal(got, raw) {
				t.Fatalf("document mutation did not target exactly its declared field: %s %v", id, err)
			}
		})
	}
}

func TestObservedAdmissionReceiptMutationFresh1896(t *testing.T) {
	_, baseline := admissionFixture1894(t)
	for _, id := range []string{"repository-write-authority", "promotion-authority", "wrong-digest", "resealed-source-value"} {
		t.Run(id, func(t *testing.T) {
			var want co.ObservedReceipt
			if err := json.Unmarshal(baseline.Receipt, &want); err != nil {
				t.Fatal(err)
			}
			switch id {
			case "repository-write-authority":
				want.RepositoryWorkspaceWrites = true
			case "promotion-authority":
				want.PromotionAuthorized = true
			case "wrong-digest":
				want.Digest = artifact.ContentDigest([]byte("admission-forged-digest"))
			case "resealed-source-value":
				want.Decision.Value = json.RawMessage(`{"readiness_credit":true,"marker":"test-only"}`)
			}
			if id != "wrong-digest" {
				want.Digest = ""
				digest, err := artifact.Digest(want)
				if err != nil {
					t.Fatal(err)
				}
				want.Digest = digest
			}
			got, err := admissionReceipt(baseline.Receipt, id)
			if err != nil || !bytes.Equal(got, marshalAdmission1894(t, want)) || bytes.Equal(got, baseline.Receipt) {
				t.Fatalf("receipt mutation changed undeclared content or lost its seal: %s %v", id, err)
			}
		})
	}
}

func TestObservedAdmissionVariantIsolationFresh1897(t *testing.T) {
	f, baseline := admissionFixture1894(t)
	before := marshalAdmission1894(t, baseline)
	for _, planned := range observedAdmissionPlans("urn:admission:declared") {
		t.Run(planned.ID, func(t *testing.T) {
			v, err := observedAdmissionVariant(baseline, f, planned)
			if err != nil || !bytes.Equal(before, marshalAdmission1894(t, baseline)) {
				t.Fatal("countercase changed the shared baseline", err)
			}
			if planned.Kind != "MISSING_OBSERVATION_INPUTS" && v.Inputs.Observation == baseline.Inputs.Observation {
				t.Fatal("countercase shares a mutable observation pointer")
			}
			switch planned.Kind {
			case "POLICY_BYTES":
				want := "// admission policy bytes changed\npackage admission_fixture\n"
				if string(v.Inputs.Observation.PolicySource) != want {
					t.Fatal("policy byte mutation differs")
				}
			case "DOCUMENT_BYTES":
				if string(v.Inputs.Observation.Document) != string(baseline.Inputs.Observation.Document)+" " {
					t.Fatal("document byte mutation differs")
				}
			case "MISSING_OBSERVATION_INPUTS":
				if v.Inputs.Observation != nil {
					t.Fatal("missing observation pair retained")
				}
			case "LEGACY_DOWNGRADE":
				if !bytes.Equal(v.Receipt, f.Raw) || v.Inputs.Observation == nil {
					t.Fatal("downgrade did not preserve the observations it attempts to ignore")
				}
			}
		})
	}
}
