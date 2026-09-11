package languagereadiness

import (
	"strings"
	"testing"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	conceptoperation "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func TestUseCaseCurrentCatalogIsSevenOfTwentyFour(t *testing.T) {
	snapshot, err := Evaluate(artifactFixture("PASS", currentConceptIDs...))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Decision != "PASS" {
		t.Fatalf("decision = %q", snapshot.Decision)
	}
	if snapshot.Summary.Completed != 6 || snapshot.Summary.Total != 24 {
		t.Fatalf("completion = %d/%d", snapshot.Summary.Completed, snapshot.Summary.Total)
	}
	if snapshot.Summary.ReadinessBPS != 2500 || snapshot.Summary.NotSatisfied != 18 {
		t.Fatalf("summary = %+v", snapshot.Summary)
	}
	if snapshot.Summary.RatioNumerator != 6 || snapshot.Summary.RatioDenominator != 24 {
		t.Fatalf("ratio = %d/%d", snapshot.Summary.RatioNumerator, snapshot.Summary.RatioDenominator)
	}
}

func TestUseCaseUnknownEvidenceLowersResolution(t *testing.T) {
	snapshot, err := Evaluate(artifactFixture("FUTURE_DECISION", currentConceptIDs...))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Decision != "LOWER_RESOLUTION" || snapshot.Summary.Unresolved != 24 {
		t.Fatalf("unexpected resolution: decision=%s summary=%+v", snapshot.Decision, snapshot.Summary)
	}
	if snapshot.Summary.Completed != 0 {
		t.Fatalf("unknown evidence completed %d obligations", snapshot.Summary.Completed)
	}
}

func TestUseCaseUnregisteredClaimsDoNotChangeTheCount(t *testing.T) {
	ids := append(append([]string(nil), currentConceptIDs...), "qualitative-language-progress")
	snapshot, err := Evaluate(artifactFixture("PASS", ids...))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Summary.Completed != 6 {
		t.Fatalf("unregistered claim changed completion to %d", snapshot.Summary.Completed)
	}
	replayed, err := Evaluate(artifactFixture("PASS", ids...))
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Digest != snapshot.Digest {
		t.Fatalf("replay digest mismatch: %s != %s", replayed.Digest, snapshot.Digest)
	}
}

func TestConceptOperationEvidenceReachesOnlyItsRegisteredObligation(t *testing.T) {
	without, err := Evaluate(artifactFixture("PASS", currentConceptIDs...))
	if err != nil {
		t.Fatal(err)
	}
	with, err := EvaluateWithConceptOperation(artifactFixture("PASS", currentConceptIDs...), conceptOperationEvidenceFixture())
	if err != nil {
		t.Fatal(err)
	}
	if without.Summary.Completed != 6 || with.Summary.Completed != 7 || with.Summary.Total != 24 {
		t.Fatalf("completion changed unexpectedly: without=%+v with=%+v", without.Summary, with.Summary)
	}
	for index, result := range with.Obligations {
		if result.ID == "META-CONCEPT-GOVERNED-REFACTORING" {
			if result.Status != "SATISFIED" || result.Reason != "CONCEPT_CONFORMANCE_EXPLICIT" || result.EvidenceDigest == "" {
				t.Fatalf("concept-operation result = %+v", result)
			}
			continue
		}
		if result.Status != without.Obligations[index].Status || result.Reason != without.Obligations[index].Reason {
			t.Fatalf("unrelated obligation changed: before=%+v after=%+v", without.Obligations[index], result)
		}
	}
}

func conceptOperationEvidenceFixture() conceptoperation.Receipt {
	receipt := conceptoperation.Receipt{
		Schema: conceptoperation.Schema, Repository: "owner/repository", SubjectSHA: strings.Repeat("a", 40),
		ExecutionPolicy: conceptoperation.ExecutionPolicy, MetricID: conceptoperation.MetricID,
		CohortRule: conceptoperation.CohortRule, SourceAuthority: conceptoperation.SourceAuthority,
		StrategyDigest: digestFixture(), StrategyVerificationDigest: digestFixture(),
		InterventionDigest: digestFixture(), ProgramDigest: digestFixture(),
		ProgramVerificationDigest: digestFixture(), ProgramSourcePath: metricprogram.ProgramSourceFilename,
		ProgramSourceDigest: digestFixture(), ProgramSemanticDigest: digestFixture(),
		ProgramRegistryDigest: digestFixture(), Expected: []conceptoperation.OperationBinding{{
			Operation: "terminate-at-fixed-point", CarrierOperation: "replay-counterfactual",
			IndicatorID: metricstrategy.ConceptOperationIndicatorID("terminate-at-fixed-point"),
			RegisteredActivity: "TerminateAtFixedPoint", RegisteredProofChoice: "REGRESSION",
			Activity: "TerminateAtFixedPoint", ProofChoice: "REGRESSION", Expected: "REGISTERED_CONCEPT",
			Actual: "REGISTERED_CONCEPT", Status: "SATISFIED", EvidenceDigest: digestFixture(), OperationDigest: digestFixture(),
		}}, ExpectedCount: 1, ObservedCount: 1, BoundCount: 1, CoverageBPS: 10000,
		Status: "VERIFIED", Producer: "metricprogram/conceptoperation.Build", Consumer: "language-readiness",
		MetaOperation: "bind-concept-operation-metric",
	}
	receipt.Digest, _ = artifact.Digest(receipt)
	return receipt
}

func digestFixture() string {
	return "sha256:0000000000000000000000000000000000000000000000000000000000000000"
}
