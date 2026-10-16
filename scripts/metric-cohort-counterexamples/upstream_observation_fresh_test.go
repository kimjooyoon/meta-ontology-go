package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func TestUpstreamObservationBoundaryFresh1745(t *testing.T) {
	for _, mode := range []string{"retained-cohort", "missing-original", "duplicate", "carrier-replacement", "invented-coverage",
		"program-output", "upstream-authority", "consumer-credit", "producer-credit"} {
		t.Run(mode, func(t *testing.T) {
			base := metricstrategy.ConceptOperationSpec{Subject: "original", Carrier: "original", Family: "REGRESSION"}
			added := metricstrategy.ConceptOperationSpec{Subject: unregisteredUpstreamOperation,
				Carrier: unregisteredUpstreamOperation, Family: "FOUNDATION", Trilemma: "AXIOM"}
			refusal := &unresolvedpolicy.OperationRefusal{
				Origin:         unresolvedpolicy.OperationOrigin{RequestedOperation: unregisteredUpstreamOperation},
				OriginRecorded: true, Reason: "OPERATION_NOT_IN_SOURCE_REGISTRY", Resolution: "invariant_only",
				NextOperation: "DECLARE_SOURCE_OPERATION",
			}
			consumer := &programverify.RefusalReport{Origin: refusal.Origin,
				Status: "VERIFIED_SOURCE_ORIGIN_REFUSAL", UpstreamVerification: "UNASSESSED"}
			result := upstreamObservation{RequestedOperation: unregisteredUpstreamOperation, Refusal: refusal, Consumer: consumer,
				DeclaredCohort: []metricstrategy.ConceptOperationSpec{base, added}, ExpectedCount: 2, AddedExpectedCount: 1}
			switch mode {
			case "missing-original":
				result.DeclaredCohort = result.DeclaredCohort[1:]
			case "duplicate":
				result.DeclaredCohort = append(result.DeclaredCohort, added)
			case "carrier-replacement":
				result.DeclaredCohort[1].Carrier = "original"
			case "invented-coverage":
				full := 10000
				result.CoverageBPS = &full
			case "program-output":
				result.ProgramProduced = true
			case "upstream-authority":
				consumer.UpstreamVerification = "VERIFIED"
			case "consumer-credit":
				consumer.ReadinessCredit = true
			case "producer-credit":
				refusal.ReadinessCredit = true
			}
			if validUpstreamObservation(result, fixture{Contract: []metricstrategy.ConceptOperationSpec{base}}) !=
				(mode == "retained-cohort") {
				t.Fatal("diagnostic conflated with cohort execution or readiness credit")
			}
		})
	}
}
