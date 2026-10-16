package main

import (
	"encoding/json"
	"strings"
	"testing"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

// This synthetic frame checks mutation and storage mechanics only. The CI
// admission cohort separately constructs a real receipt from original inputs.
func admissionFixture1894(t *testing.T) (fixture, variant) {
	t.Helper()
	document := co.ObservationDocument{Schema: co.ObservationDocumentSchema,
		Repository:           "example/gooo",
		SubjectSHA:           strings.Repeat("a", 40),
		ProgramDigest:        artifact.ContentDigest([]byte("program")),
		DeclaredCohortDigest: artifact.ContentDigest([]byte("declared")),
	}
	if err := json.Unmarshal([]byte(`[{"operation":"urn:admission:declared","binding":{"actual":"original-actual"}}]`),
		&document.Observations); err != nil {
		t.Fatal(err)
	}
	receipt := co.ObservedReceipt{Schema: co.ObservedReceiptSchema, MetricID: co.MetricID,
		Producer: "metricprogram/conceptoperation.BuildObserved", Consumer: "language-readiness"}
	receipt.Decision.Value = json.RawMessage(`{"readiness_credit":false,"marker":"test-only"}`)
	digest, err := artifact.Digest(receipt)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Digest = digest
	f := fixture{Raw: []byte(`{"schema":"legacy-test-only"}`), Inputs: co.SourceInputs{
		Strategy: []byte("strategy"), StrategyVerification: []byte("strategy verification"),
		SourceMetrics: []byte("metrics"), Intervention: []byte("ledger"),
		InterventionVerification: []byte("ledger verification"), Program: []byte("program"),
		ProgramSource: []byte("source"), ProgramVerification: []byte("program verification"),
	}}
	baseline := variant{Receipt: marshalAdmission1894(t, receipt), Inputs: f.Inputs}
	baseline.Inputs.Observation = &co.ObservationInputs{Document: marshalAdmission1894(t, document),
		PolicySource: []byte("package admission_fixture\n")}
	return f, baseline
}

func marshalAdmission1894(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
