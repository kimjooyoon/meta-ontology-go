package main

import (
	"bytes"
	"encoding/json"
	"testing"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	metricverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention/verify"
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func TestUpstreamLedgerIdentityFresh1743(t *testing.T) {
	for _, mode := range []string{"extend-source-cohort", "already-present", "malformed-ledger", "malformed-verification"} {
		t.Run(mode, func(t *testing.T) {
			ledger := metric.Ledger{Indicators: []metric.Indicator{
				{ID: "source", Family: "FOUNDATION", Trilemma: "AXIOM", MetaOperation: "bind-exact-source-metrics",
					EvidenceDigest: "source-digest", Status: "SATISFIED"},
			}}
			if mode == "already-present" {
				ledger.Indicators[0].MetaOperation = unregisteredUpstreamOperation
			}
			original, err := metricstrategy.ConceptOperationCohort(ledger.Indicators)
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(ledger)
			verification, _ := json.Marshal(metricverify.Receipt{SourceMetricsDigest: "retained", ProjectionCount: 7})
			input := co.SourceInputs{Intervention: raw, InterventionVerification: verification,
				SourceMetrics: []byte("retained metrics"), Program: []byte("old program"),
				ProgramSource: []byte("old source"), ProgramVerification: []byte("old verification")}
			if mode == "malformed-ledger" {
				input.Intervention = []byte("{")
			}
			if mode == "malformed-verification" {
				input.InterventionVerification = []byte("{")
			}
			actual, cohort, err := extendUnregisteredLedger(input)
			if mode != "extend-source-cohort" {
				if err == nil {
					t.Fatal("invalid upstream input admitted")
				}
				return
			}
			if err != nil || len(cohort) != len(original)+1 || !bytes.Equal(input.Intervention, raw) ||
				!bytes.Equal(actual.SourceMetrics, input.SourceMetrics) || actual.Program != nil ||
				actual.ProgramSource != nil || actual.ProgramVerification != nil {
				t.Fatalf("upstream cohort or original/absent inputs changed: %v", err)
			}
			var changed metric.Ledger
			var proof metricverify.Receipt
			if json.Unmarshal(actual.Intervention, &changed) != nil || !metric.ValidLedger(changed) ||
				json.Unmarshal(actual.InterventionVerification, &proof) != nil || proof.LedgerDigest != changed.Digest ||
				proof.IndicatorCount != len(changed.Indicators) || proof.SourceMetricsDigest != "retained" || proof.ProjectionCount != 7 {
				t.Fatal("new upstream bytes lost exact ledger and input identity")
			}
			digest := proof.Digest
			proof.Digest = ""
			if expected, _ := artifact.Digest(proof); digest != expected {
				t.Fatal("adversarial receipt was not sealed consistently")
			}
		})
	}
}
