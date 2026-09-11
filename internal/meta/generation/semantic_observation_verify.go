package generation

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

func verifyBoundSemanticObservation(manifest semanticOperationManifest, patch SemanticPatch, receipt SemanticOperationReceipt) error {
	observation := receipt.Observation
	if observation == nil {
		if manifest.ObservationDigest != "" || patch.Observation != nil {
			return errors.New("observation binding is incomplete")
		}
		return nil
	}
	if patch.Observation == nil || manifest.ObservationDigest == "" {
		return errors.New("observation binding is incomplete")
	}
	if manifest.ObservationDigest != envelopeDigestJSON(*observation) || envelopeDigestJSON(*patch.Observation) != envelopeDigestJSON(*observation) {
		return errors.New("observation binding digest mismatch")
	}
	if err := validateSemanticObservationContract(observation.Contract); err != nil {
		return err
	}
	if receipt.Metrics.ObservedOperations != observation.Metrics.ObservedOperations ||
		receipt.Metrics.DistinctInputDigests != observation.Metrics.DistinctInputDigests ||
		receipt.Metrics.DuplicateEvaluations != observation.Metrics.DuplicateEvaluations ||
		receipt.Metrics.CandidatesEmitted != observation.Metrics.CandidatesEmitted ||
		receipt.Metrics.BeforeOperationCount != observation.Metrics.BeforeOperationCount ||
		receipt.Metrics.AfterOperationCount != observation.Metrics.AfterOperationCount ||
		receipt.Metrics.AllocationCount != observation.Metrics.AllocationCount ||
		receipt.Metrics.AllocationBytes != observation.Metrics.AllocationBytes ||
		receipt.Metrics.WallMS != int(observation.Metrics.WallMS) ||
		receipt.Metrics.PeakRSSKib != int(observation.Metrics.PeakRSSKib) ||
		receipt.Metrics.BuildMS != observation.Metrics.BuildMS ||
		receipt.Metrics.TestMS != observation.Metrics.TestMS ||
		receipt.Metrics.ExecutedTests != observation.Metrics.ExecutedTests ||
		receipt.Metrics.ReusedTests != observation.Metrics.ReusedTests {
		return errors.New("observation metrics are not bound into the envelope receipt")
	}
	decision, reason, unknown, err := independentlyClassifySemanticObservation(*observation)
	if err != nil {
		return err
	}
	if decision != observation.Decision || reason != observation.Reason {
		return fmt.Errorf("independent observation decision mismatch: got %s/%s, report %s/%s", decision, reason, observation.Decision, observation.Reason)
	}
	if decision == "UNKNOWN" {
		if !sameEnvelopeUnknown(observation.Unknown, unknown) {
			return errors.New("observation UNKNOWN evidence is not causal and exact")
		}
	} else if observation.Unknown != nil {
		return errors.New("non-unknown observation contains unknown evidence")
	}
	if observation.Adoption != nil {
		if err := ValidateSemanticAdoptionEvidence(*observation.Adoption); err != nil {
			return err
		}
		if len(observation.Candidates) != 1 || observation.Adoption.CandidateStableID != observation.Candidates[0].StableID ||
			observation.Adoption.InputDigest != observation.Candidates[0].InputDigest {
			return errors.New("semantic adoption evidence is not bound to the observed candidate")
		}
		if err := ValidateBoundSemanticAdoption(*observation, SemanticAdoptionProposal{Candidate: observation.Candidates[0]}, *observation.Adoption); err != nil {
			return err
		}
	}
	return nil
}

func independentlyClassifySemanticObservation(observation SemanticObservation) (string, string, *EnvelopeUnknownState, error) {
	if observation.Schema != SemanticObservationSchema {
		return "", "", nil, errors.New("semantic observation schema mismatch")
	}
	if observation.ContractDigest == "" || !knownEnvelopeDigest(observation.ContractDigest) {
		return "", "", nil, errors.New("semantic observation contract digest is unknown")
	}
	if observation.InputSourceDigest != "" && !knownEnvelopeDigest(observation.InputSourceDigest) {
		return "REFUTED", SemanticObservationContradiction, nil, nil
	}
	if observation.Metrics.RepositoryWrites != 0 || observation.Metrics.LocalTestExecutions != 0 {
		return "REFUTED", SemanticObservationContradiction, nil, nil
	}
	type group struct {
		phase       string
		operationID string
		inputDigest string
		count       int
		spans       []SemanticObservationSpan
	}
	groups := make(map[string]*group)
	digests := make(map[string]struct{})
	for index, event := range observation.Events {
		if event.Sequence != index+1 || event.Phase != observation.Contract.Phase || event.OperationID != observation.Contract.OperationID || !event.Pure || !knownEnvelopeDigest(event.InputDigest) || len(event.SourceSpans) == 0 || !subsetStrings(event.Effects, observation.Contract.AllowedEffects) {
			return "REFUTED", SemanticObservationContradiction, nil, nil
		}
		for _, span := range event.SourceSpans {
			if !validSemanticObservationSpan(span) {
				return "REFUTED", SemanticObservationContradiction, nil, nil
			}
		}
		key := event.Phase + "\x00" + event.OperationID + "\x00" + event.InputDigest
		current := groups[key]
		if current == nil {
			current = &group{phase: event.Phase, operationID: event.OperationID, inputDigest: event.InputDigest}
			groups[key] = current
		}
		current.count++
		current.spans = appendUniqueObservationSpans(current.spans, event.SourceSpans...)
		digests[event.InputDigest] = struct{}{}
	}
	duplicateEvaluations := 0
	expectedCandidates := make(map[string]SemanticObservationCandidate)
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		current := groups[key]
		if current.count < 2 {
			continue
		}
		duplicateEvaluations += current.count - 1
		spans := sortedObservationSpans(current.spans)
		expected := SemanticObservationCandidate{
			StableID:               semanticObservationCandidateID(current.phase, current.operationID, current.inputDigest, spans),
			Phase:                  current.phase,
			OperationID:            current.operationID,
			InputDigest:            current.inputDigest,
			SourceSpans:            spans,
			ObservedCount:          current.count,
			ExpectedReducibleCount: current.count - 1,
			SafetyAssessment:       "UNKNOWN_NOT_INFERRED",
			BenefitAssessment:      "UNKNOWN_NOT_INFERRED",
		}
		expectedCandidates[semanticObservationCandidateKey(expected)] = expected
	}
	if observation.ObservedOperations != len(observation.Events) ||
		observation.DistinctInputDigests != len(digests) ||
		observation.DuplicateEvaluations != duplicateEvaluations ||
		observation.CandidatesEmitted != len(observation.Candidates) ||
		observation.CandidatesEmitted != len(expectedCandidates) ||
		observation.Metrics.ObservedOperations != observation.ObservedOperations ||
		observation.Metrics.DistinctInputDigests != observation.DistinctInputDigests ||
		observation.Metrics.DuplicateEvaluations != observation.DuplicateEvaluations ||
		observation.Metrics.CandidatesEmitted != observation.CandidatesEmitted ||
		observation.Metrics.BeforeOperationCount != observation.PairEvidence.BeforeOperationCount ||
		observation.Metrics.AfterOperationCount != observation.PairEvidence.AfterOperationCount {
		return "REFUTED", SemanticObservationContradiction, nil, nil
	}
	for _, candidate := range observation.Candidates {
		expected, ok := expectedCandidates[semanticObservationCandidateKey(candidate)]
		if !ok || candidate.StableID != expected.StableID || candidate.Phase != expected.Phase || candidate.OperationID != expected.OperationID || candidate.InputDigest != expected.InputDigest || candidate.ObservedCount != expected.ObservedCount || candidate.ExpectedReducibleCount != expected.ExpectedReducibleCount || candidate.SafetyAssessment != expected.SafetyAssessment || candidate.BenefitAssessment != expected.BenefitAssessment || !sameObservationSpans(candidate.SourceSpans, expected.SourceSpans) {
			return "REFUTED", SemanticObservationContradiction, nil, nil
		}
		delete(expectedCandidates, semanticObservationCandidateKey(candidate))
	}
	if len(expectedCandidates) != 0 {
		return "REFUTED", SemanticObservationContradiction, nil, nil
	}
	if observation.PairEvidence.Contradiction != "" {
		return "REFUTED", SemanticObservationContradiction, nil, nil
	}
	if !observation.PairEvidence.EvidenceAvailable {
		return "UNKNOWN", SemanticObservationUnknownReason, semanticObservationUnknownState(), nil
	}
	if observation.PairEvidence.ChangeAdopted {
		if observation.PairEvidence.BeforeOperationCount <= 0 || observation.PairEvidence.AfterOperationCount <= 0 {
			return "UNKNOWN", SemanticObservationUnknownReason, semanticObservationUnknownState(), nil
		}
		if !observation.PairEvidence.BehaviorEqual || !observation.PairEvidence.DeterminismEqual {
			return "REFUTED", "BEHAVIOR_OR_DETERMINISM_MISMATCH", nil, nil
		}
	}
	return "CLOSED", observationReason(observation.CandidatesEmitted), nil, nil
}

func semanticObservationUnknownState() *EnvelopeUnknownState {
	return &EnvelopeUnknownState{
		Stage:         SemanticObservationUnknownStage,
		Step:          SemanticObservationUnknownStep,
		Reason:        SemanticObservationUnknownReason,
		UnknownClass:  SemanticObservationUnknownClass,
		NextOperation: SemanticObservationUnknownNext,
		BlockedBy:     []string{"behavior_determinism_pair"},
	}
}

func semanticObservationCandidateKey(candidate SemanticObservationCandidate) string {
	return candidate.Phase + "\x00" + candidate.OperationID + "\x00" + candidate.InputDigest
}

func sameObservationSpans(left, right []SemanticObservationSpan) bool {
	left = sortedObservationSpans(left)
	right = sortedObservationSpans(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameEnvelopeUnknown(left, right *EnvelopeUnknownState) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason && left.UnknownClass == right.UnknownClass && left.NextOperation == right.NextOperation && strings.Join(left.BlockedBy, "\x00") == strings.Join(right.BlockedBy, "\x00")
}
