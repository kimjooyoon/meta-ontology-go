package metricstrategy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
)

type conceptOperation struct {
	subject, carrier, family, trilemma string
}

// ConceptOperationSpec is the typed operation cohort emitted by the strategy
// producer. The cohort is derived from upstream intervention indicators and
// the two registered terminal bindings; it is not derived from the observed
// concept bindings in the sealed plan.
type ConceptOperationSpec struct {
	Subject  string
	Carrier  string
	Family   string
	Trilemma string
}

// ConceptOperationCohort reconstructs the expected concept-operation cohort
// from the non-concept intervention indicators. The terminal rule is fixed:
// both registered terminal bindings are always part of the cohort.
func ConceptOperationCohort(indicators []metric.Indicator) ([]ConceptOperationSpec, error) {
	source, err := buildBindings(indicators)
	if err != nil {
		return nil, err
	}
	return conceptOperationSpecs(source)
}

// ConceptIDForOperation returns the catalog concept bound to an operation.
func ConceptIDForOperation(operation string) (string, bool) {
	conceptID, ok := operationConceptIDs[operation]
	return conceptID, ok
}

// ConceptOperationIndicatorID returns the canonical indicator ID for an
// operation, including the explicit unresolved form for unmapped operations.
func ConceptOperationIndicatorID(operation string) string {
	if _, ok := operationConceptIDs[operation]; ok {
		return "gooo.concept.operation." + operation + ".v1"
	}
	return "gooo.concept.unresolved-operation." + operation + ".v1"
}

// IsConceptOperationBinding identifies the declared concept-operation
// binding family. This is a schema boundary, not an observed-count heuristic.
func IsConceptOperationBinding(indicatorID string) bool {
	return strings.HasPrefix(indicatorID, "gooo.concept.operation.") ||
		strings.HasPrefix(indicatorID, "gooo.concept.unresolved-operation.")
}

func conceptOperationBindings(value languageconcept.Artifact, source []Binding) ([]Binding, error) {
	known := make(map[string]bool, len(value.Report.Concepts))
	for _, concept := range value.Report.Concepts {
		known[concept.ID] = true
	}
	specs, err := conceptOperationSpecs(source)
	if err != nil {
		return nil, err
	}
	result := make([]Binding, 0, len(specs))
	for _, spec := range specs {
		binding, err := conceptOperationBinding(value.ArtifactDigest, conceptOperation{
			subject: spec.Subject, carrier: spec.Carrier, family: spec.Family, trilemma: spec.Trilemma,
		}, known)
		if err != nil {
			return nil, err
		}
		result = append(result, binding)
	}
	return result, nil
}

func conceptOperationSpecs(source []Binding) ([]ConceptOperationSpec, error) {
	byOperation := make(map[string]conceptOperation)
	for _, binding := range source {
		current := conceptOperation{binding.MetaOperation, binding.MetaOperation, binding.Family, binding.Trilemma}
		if previous, ok := byOperation[current.subject]; ok && previous != current {
			return nil, fmt.Errorf("operation %q spans proof families", current.subject)
		}
		byOperation[current.subject] = current
	}
	byOperation["terminate-at-fixed-point"] = conceptOperation{"terminate-at-fixed-point", "replay-counterfactual", "REGRESSION", "REGRESS"}
	byOperation["preserve-non-promoting-terminal"] = conceptOperation{"preserve-non-promoting-terminal", "preserve-non-promoting-terminal", "REGRESSION", "REGRESS"}
	keys := make([]string, 0, len(byOperation))
	for operation := range byOperation {
		keys = append(keys, operation)
	}
	sort.Strings(keys)
	result := make([]ConceptOperationSpec, 0, len(keys))
	for _, operation := range keys {
		value := byOperation[operation]
		result = append(result, ConceptOperationSpec{Subject: value.subject, Carrier: value.carrier, Family: value.family, Trilemma: value.trilemma})
	}
	return result, nil
}

func conceptOperationBinding(artifactDigest string, operation conceptOperation, known map[string]bool) (Binding, error) {
	conceptID, mapped := operationConceptIDs[operation.subject]
	id, expected, actual, status := "gooo.concept.operation."+operation.subject+".v1", conceptID, conceptID, "SATISFIED"
	if !mapped {
		id, expected, actual, status = "gooo.concept.unresolved-operation."+operation.subject+".v1", "REGISTERED_CONCEPT", "UNKNOWN", "UNSATISFIED"
	} else if !known[conceptID] {
		id, actual, status = "gooo.concept.unresolved-operation."+operation.subject+".v1", "MISSING", "UNSATISFIED"
	}
	digest, err := conceptEvidenceDigest(map[string]string{"artifact_digest": artifactDigest, "kind": "OPERATION", "subject_id": operation.subject, "proof_choice": operation.family, "carrier_operation": operation.carrier, "concept_id": conceptID, "expected": expected, "actual": actual, "status": status})
	return Binding{IndicatorID: id, Family: operation.family, Trilemma: operation.trilemma, MetaOperation: operation.carrier, Expected: expected, Actual: actual, Status: status, EvidenceDigest: digest}, err
}
