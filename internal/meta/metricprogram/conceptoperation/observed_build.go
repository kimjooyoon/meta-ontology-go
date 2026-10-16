package conceptoperation

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

// BuildObserved evaluates separate observations against the unchanged program.
// It can restrict a verified baseline; it cannot repair or upgrade one.
func BuildObserved(ctx context.Context, inputs SourceInputs) (ObservedReceipt, error) {
	if inputs.Observation == nil {
		return ObservedReceipt{}, fmt.Errorf("concept-operation observation inputs are missing")
	}
	baseline, err := Build(inputs.Strategy, inputs.StrategyVerification, inputs.Intervention,
		inputs.Program, inputs.ProgramSource, inputs.ProgramVerification)
	if err != nil {
		return ObservedReceipt{}, err
	}
	if baseline.Status != "VERIFIED" {
		return ObservedReceipt{}, fmt.Errorf("observations require a verified canonical producer")
	}
	declaration, err := DeclareObservations(baseline, inputs.Strategy)
	if err != nil {
		return ObservedReceipt{}, err
	}
	var document ObservationDocument
	if err := completeness.CheckJSON(inputs.Observation.Document, 1<<20); err != nil {
		return ObservedReceipt{}, err
	}
	if err := decodeExact(inputs.Observation.Document, &document); err != nil {
		return ObservedReceipt{}, err
	}
	if document.Schema != ObservationDocumentSchema || document.Repository != baseline.Repository ||
		document.SubjectSHA != baseline.SubjectSHA || document.ProgramDigest != baseline.ProgramDigest ||
		document.DeclaredCohortDigest != declaration.DeclaredCohortDigest {
		return ObservedReceipt{}, fmt.Errorf("operation observation source identity is not exact")
	}
	expected, observed, err := validateObservationRows(document, baseline, inputs.Strategy)
	if err != nil {
		return ObservedReceipt{}, err
	}
	tuple, err := json.Marshal([]any{map[string]any{"ids": expected}, map[string]any{"ids": observed}})
	if err != nil {
		return ObservedReceipt{}, err
	}
	decision, err := bodycodegen.EvaluateFixedValue(ctx, ObservationPolicyPath,
		inputs.Observation.PolicySource, ObservationPolicyActivity, tuple)
	if err != nil {
		return ObservedReceipt{}, fmt.Errorf("evaluate source-owned observations: %w", err)
	}
	receipt := ObservedReceipt{Schema: ObservedReceiptSchema, MetricID: MetricID, Baseline: baseline,
		DeclaredCohortDigest: declaration.DeclaredCohortDigest,
		ObservationInputDigest: artifact.ContentDigest(inputs.Observation.Document), Decision: decision,
		Producer: "metricprogram/conceptoperation.BuildObserved", Consumer: "language-readiness"}
	assessment, err := observedAssessment(receipt)
	if err != nil {
		return ObservedReceipt{}, err
	}
	if !slices.Equal(assessment.ObservedIDs, observed) {
		return ObservedReceipt{}, fmt.Errorf("source-owned assessment did not preserve observed identities")
	}
	receipt.Digest, err = artifact.Digest(receipt)
	return receipt, err
}

func observedAssessment(receipt ObservedReceipt) (ObservationAssessment, error) {
	var result ObservationAssessment
	if err := decodeExact(receipt.Decision.Value, &result); err != nil {
		return result, err
	}
	if receipt.Decision.ActivityID != "urn:cohort:assess-declared" || result.ReadinessCredit ||
		(result.Status != "VERIFIED" && result.Status != "FAIL_CLOSED") ||
		result.ExpectedCount != receipt.Baseline.ExpectedCount || result.ExpectedCount == 0 ||
		result.BoundCount < 0 || result.BoundCount > result.ExpectedCount || result.UnknownCount < 0 ||
		result.CoverageBPS != result.BoundCount*10000/result.ExpectedCount {
		return result, fmt.Errorf("source-owned observation assessment is invalid")
	}
	if result.Status == "VERIFIED" && (result.BoundCount != result.ExpectedCount || result.UnknownCount != 0 ||
		result.CoverageBPS != 10000 || result.UnknownClass != "" || result.Reason != "" || result.NextOperation != "") {
		return result, fmt.Errorf("verified observation assessment is incomplete")
	}
	if result.Status == "FAIL_CLOSED" && (result.UnknownClass == "" || result.Reason == "" ||
		result.NextOperation == "" || (result.CoverageBPS == 10000 && result.UnknownCount == 0)) {
		return result, fmt.Errorf("incomplete observation assessment lacks an unresolved route")
	}
	return result, nil
}
