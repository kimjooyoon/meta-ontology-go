package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/unresolvedpolicy"
	programverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/verify"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func TestUpstreamCohortCardinalityFresh1747(t *testing.T) {
	for _, mode := range []string{"claimed-count", "extra-unaccounted-subject", "added-count"} {
		t.Run(mode, func(t *testing.T) {
			base := metricstrategy.ConceptOperationSpec{Subject: "base", Carrier: "base"}
			added := metricstrategy.ConceptOperationSpec{Subject: unregisteredUpstreamOperation,
				Carrier: unregisteredUpstreamOperation, Family: "FOUNDATION", Trilemma: "AXIOM"}
			refusal := &unresolvedpolicy.OperationRefusal{
				Origin:         unresolvedpolicy.OperationOrigin{RequestedOperation: unregisteredUpstreamOperation},
				OriginRecorded: true, Reason: "OPERATION_NOT_IN_SOURCE_REGISTRY", Resolution: "invariant_only",
				NextOperation: "DECLARE_SOURCE_OPERATION",
			}
			result := upstreamObservation{RequestedOperation: unregisteredUpstreamOperation, Refusal: refusal,
				Consumer: &programverify.RefusalReport{Origin: refusal.Origin,
					Status: "VERIFIED_SOURCE_ORIGIN_REFUSAL", UpstreamVerification: "UNASSESSED"},
				DeclaredCohort: []metricstrategy.ConceptOperationSpec{base, added}, ExpectedCount: 2, AddedExpectedCount: 1}
			switch mode {
			case "claimed-count":
				result.ExpectedCount = 1
			case "extra-unaccounted-subject":
				result.DeclaredCohort = append(result.DeclaredCohort, metricstrategy.ConceptOperationSpec{Subject: "extra"})
			case "added-count":
				result.AddedExpectedCount = 0
			}
			if validUpstreamObservation(result, fixture{Contract: []metricstrategy.ConceptOperationSpec{base}}) {
				t.Fatal("expected population was changed after classification")
			}
		})
	}
}
