package metricstrategyverify

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

type replayConceptOperation struct {
	subject, carrier, family, trilemma string
}

func replayConceptOperations(value languageconcept.Artifact, source []strategy.Binding) ([]strategy.Binding, error) {
	known := make(map[string]bool, len(value.Report.Concepts))
	for _, concept := range value.Report.Concepts {
		known[concept.ID] = true
	}
	byOperation := make(map[string]replayConceptOperation)
	for _, binding := range source {
		current := replayConceptOperation{binding.MetaOperation, binding.MetaOperation, binding.Family, binding.Trilemma}
		if previous, ok := byOperation[current.subject]; ok && previous != current {
			return nil, fmt.Errorf("independent operation %q spans proof families", current.subject)
		}
		byOperation[current.subject] = current
	}
	byOperation["terminate-at-fixed-point"] = replayConceptOperation{"terminate-at-fixed-point", "replay-counterfactual", "REGRESSION", "REGRESS"}
	byOperation["preserve-non-promoting-terminal"] = replayConceptOperation{"preserve-non-promoting-terminal", "preserve-non-promoting-terminal", "REGRESSION", "REGRESS"}
	keys := make([]string, 0, len(byOperation))
	for operation := range byOperation {
		keys = append(keys, operation)
	}
	sort.Strings(keys)
	result := make([]strategy.Binding, 0, len(keys))
	for _, operation := range keys {
		binding, err := replayConceptOperationBinding(value.ArtifactDigest, byOperation[operation], known)
		if err != nil {
			return nil, err
		}
		result = append(result, binding)
	}
	return result, nil
}

func replayConceptOperationBinding(artifactDigest string, operation replayConceptOperation, known map[string]bool) (strategy.Binding, error) {
	conceptID, mapped := replayOperationConceptIDs[operation.subject]
	id, expected, actual, status := "gooo.concept.operation."+operation.subject+".v1", conceptID, conceptID, "SATISFIED"
	if !mapped {
		id, expected, actual, status = "gooo.concept.unresolved-operation."+operation.subject+".v1", "REGISTERED_CONCEPT", "UNKNOWN", "UNSATISFIED"
	} else if !known[conceptID] {
		id, actual, status = "gooo.concept.unresolved-operation."+operation.subject+".v1", "MISSING", "UNSATISFIED"
	}
	digest, err := replayConceptEvidenceDigest(map[string]string{"artifact_digest": artifactDigest, "kind": "OPERATION", "subject_id": operation.subject, "proof_choice": operation.family, "carrier_operation": operation.carrier, "concept_id": conceptID, "expected": expected, "actual": actual, "status": status})
	return strategy.Binding{IndicatorID: id, Family: operation.family, Trilemma: operation.trilemma, MetaOperation: operation.carrier, Expected: expected, Actual: actual, Status: status, EvidenceDigest: digest}, err
}
