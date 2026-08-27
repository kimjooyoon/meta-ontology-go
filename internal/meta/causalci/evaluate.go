package causalci

import (
	"fmt"
	"sort"
	"strings"
)

// Evaluate produces a plan-only receipt from raw CI observations and a real
// Gooo policy source. No check command is selected from the observation.
func Evaluate(observationRaw []byte, sourcePath string, source []byte) (Receipt, error) {
	observation, err := decodeObservation(observationRaw)
	if err != nil {
		return Receipt{}, err
	}
	if observation.SourcePath != sourcePath || len(source) == 0 {
		return Receipt{}, fmt.Errorf("%s: source binding", ReasonMalformedObservation)
	}
	policy, err := ReconstructPolicy(sourcePath, source)
	if err != nil {
		return Receipt{}, err
	}

	operation := deriveOperation(observation)
	receipt := Receipt{
		Schema:            ReceiptSchema,
		Scope:             ReceiptScope,
		Source:            policy.Source,
		ObservationDigest: digestBytes(observationRaw),
		Operation:         operation,
		ExecutionMode:     "PLAN_ONLY",
		IndependentVerifier: IndependentVerifier{
			ID: "gooo://consumer/causal-ci-selection", Mode: "INDEPENDENT_RECONSTRUCTION", Required: true, ReadOnly: true,
		},
	}
	receipt.Conformance = conformanceFor(policy)
	receipt.Subjects = evaluateSubjects(observation, policy)
	receipt.ClaimTransitions = appendClaimTransitions(observation, policy, receipt.Subjects, receipt.ObservationDigest)
	receipt.Metrics = deriveMetrics(observation, policy, receipt)
	receipt.Indicators = deriveIndicators(observation, policy, receipt)
	for _, indicator := range receipt.Indicators {
		if indicator.Satisfied {
			receipt.Metrics.FixedIndicatorSatisfied++
		}
	}
	receipt.PlanDigest, err = planDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	receipt.Digest, err = receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func deriveOperation(observation Observation) Operation {
	writes := repositoryWriteCount(observation.Isolation)
	return Operation{
		Producer:          "gooo://producer/causal-ci-selection",
		Consumer:          "gooo://consumer/causal-ci-selection",
		MetaOperation:     "causal-ci-select",
		ProofChoice:       ProofCausalPath,
		ReadOnly:          writes == 0,
		RepositoryWrites:  writes,
		MutationAuthority: writes != 0,
	}
}

func conformanceFor(policy PolicyGraph) Conformance {
	if len(policy.Contradictions) > 0 {
		value := policy.Contradictions[0]
		return Conformance{Decision: ConformanceFailClosed, Coordinate: Coordinate{Stage: value.Stage, Step: value.Step, Reason: value.Reason}}
	}
	return Conformance{Decision: ConformancePass, Coordinate: Coordinate{Stage: stageConformance, Step: stepLower, Reason: "GOOO_POLICY_SEMANTIC_GRAPH_RECONSTRUCTED"}}
}

func evaluateSubjects(observation Observation, policy PolicyGraph) []SubjectResolution {
	changed := sortedChangedFiles(observation.ChangedFiles)
	result := make([]SubjectResolution, 0, len(changed))
	for _, file := range changed {
		if len(policy.Contradictions) > 0 {
			contradiction := policy.Contradictions[0]
			result = append(result, SubjectResolution{
				Path: file.Path, Resolution: ResolutionFailClosed,
				Coordinate:     Coordinate{Stage: contradiction.Stage, Step: contradiction.Step, Reason: contradiction.Reason},
				SelectedChecks: []CheckChoice{},
			})
			continue
		}
		if file.Path != observation.SourcePath || file.Status == "D" {
			result = append(result, unknownSubject(file, observation.SourcePath))
			continue
		}
		result = append(result, selectedSubject(file.Path, policy))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func selectedSubject(path string, policy PolicyGraph) SubjectResolution {
	first := policyEdge(policy, "changed-file-to-claim", policy.ChangedFileID)
	second := policyEdge(policy, "claim-to-surface", policy.ClaimID)
	third := policyEdge(policy, "surface-to-check", policy.SurfaceID)
	if len(first) != 1 || len(second) != 1 || len(third) != 1 || first[0].To != policy.ClaimID || second[0].To != policy.SurfaceID {
		return unknownSubject(ChangedFileObservation{Path: path, Status: "M"}, policy.Source.Path)
	}
	checkID := checkIDBySemanticID(policy, third[0].To)
	if checkID == "" {
		return unknownSubject(ChangedFileObservation{Path: path, Status: "M"}, policy.Source.Path)
	}
	pathEvidence := PathEvidence{
		SubjectPath:    path,
		ClaimIDs:       []string{policy.ClaimID},
		SurfaceID:      policy.SurfaceID,
		CheckID:        checkID,
		PolicyEdgeIDs:  []string{first[0].ID, second[0].ID, third[0].ID},
		SemanticDigest: policy.Source.SemanticDigest,
		Explanation:    "changed-file observation traverses claim-to-surface and surface-to-check semantic policy",
		ProofChoice:    ProofCausalPath,
	}
	return SubjectResolution{
		Path: path, Resolution: ResolutionSelected,
		Coordinate:     Coordinate{Stage: stageSubject, Step: stepSelectChecks, Reason: ReasonCompletePath},
		Paths:          []PathEvidence{pathEvidence},
		SelectedChecks: []CheckChoice{{CheckID: checkID, ProofChoice: ProofCausalPath, Reason: ReasonCompletePath, ClaimIDs: []string{policy.ClaimID}, PathIDs: pathEvidence.PolicyEdgeIDs}},
	}
}

func unknownSubject(file ChangedFileObservation, sourcePath string) SubjectResolution {
	reason := ReasonUnknownSubject
	if file.Path == sourcePath && file.Status == "D" {
		reason = "SOURCE_OBJECT_NOT_AVAILABLE"
	}
	unknown := UnknownCause{
		SubjectPath: file.Path,
		Coordinate:  Coordinate{Stage: stageSubject, Step: stepObserveSubject, Reason: reason},
		Provenance:  "git://pull-request/changed-file/" + file.Path,
	}
	choices := make([]CheckChoice, 0, FixedCheckDenominator)
	for _, id := range fixedCheckIDs {
		choices = append(choices, CheckChoice{CheckID: id, ProofChoice: ProofFullDescent, Reason: reason})
	}
	return SubjectResolution{
		Path: file.Path, Resolution: ResolutionUnknown,
		Coordinate:    Coordinate{Stage: stageSubject, Step: stepDescendFull, Reason: reason},
		UnknownCauses: []UnknownCause{unknown}, SelectedChecks: choices,
	}
}

func appendClaimTransitions(observation Observation, policy PolicyGraph, subjects []SubjectResolution, observationDigest string) []ClaimTransition {
	byPath := map[string]SubjectResolution{}
	for _, subject := range subjects {
		byPath[subject.Path] = subject
	}
	prior := sortedPriorClaims(observation.PriorClaims)
	result := make([]ClaimTransition, 0, len(prior))
	previous := ""
	for index, claim := range prior {
		after := claim.State
		resolution := PlanNone
		reason := ReasonClaimLowered
		stage, step := stageClaimLedger, stepClaimTransition
		if len(policy.Contradictions) > 0 {
			after, resolution, reason = ClaimRefuted, PlanNone, ReasonClaimRefuted
			stage, step = stageConformance, stepValidatePolicy
		} else if subject, exists := byPath[claim.SubjectPath]; exists {
			switch subject.Resolution {
			case ResolutionSelected:
				after, resolution, reason = ClaimDischarged, PlanSelective, ReasonClaimDischarged
			case ResolutionUnknown:
				after, resolution, reason = claim.State, PlanFull, ReasonClaimLowered
			case ResolutionFailClosed:
				after, resolution, reason = ClaimRefuted, PlanNone, ReasonClaimRefuted
			}
		}
		if claim.State == ClaimRefuted {
			after = ClaimRefuted
		}
		evidence, _ := digestJSON(struct {
			Observation string                `json:"observation"`
			Source      string                `json:"source"`
			Claim       PriorClaimObservation `json:"claim"`
			After       string                `json:"after"`
			Resolution  string                `json:"resolution"`
		}{observationDigest, policy.Source.SemanticDigest, claim, after, resolution})
		transition := ClaimTransition{
			Sequence: index + 1, ClaimID: claim.ClaimID, SubjectPath: claim.SubjectPath,
			Before: claim.State, After: after, Resolution: resolution,
			Stage: stage, Step: step, Reason: reason, EvidenceDigest: evidence,
			Provenance: claim.Provenance, PreviousDigest: previous,
		}
		transition.Digest, _ = transitionDigest(transition)
		result = append(result, transition)
		previous = transition.Digest
	}
	return result
}

func deriveMetrics(observation Observation, policy PolicyGraph, receipt Receipt) Metrics {
	metrics := Metrics{
		ChangedFileNumerator: len(observation.ChangedFiles), ChangedFileDenominator: len(observation.ChangedFiles),
		SubjectTotal: len(receipt.Subjects), FullSuiteCheckDenominator: len(policy.Checks),
		ClaimTransitionTotal: len(receipt.ClaimTransitions), FixedIndicatorDenominator: FixedIndicatorDenominator,
		SourceReconstructionNumer: 1, SourceReconstructionDenom: 1,
	}
	for _, subject := range receipt.Subjects {
		metrics.SelectedCheckTotal += len(subject.SelectedChecks)
		switch subject.Resolution {
		case ResolutionSelected:
			metrics.SelectedSubjectTotal++
		case ResolutionUnknown:
			metrics.UnknownSubjectTotal++
		case ResolutionFailClosed:
			metrics.FailClosedSubjectTotal++
		}
	}
	for _, transition := range receipt.ClaimTransitions {
		switch transition.After {
		case ClaimDischarged:
			metrics.DischargedClaimTotal++
		case ClaimRefuted:
			metrics.RefutedClaimTotal++
		}
		if transition.Resolution == PlanFull {
			metrics.LowerResolutionClaimTotal++
		}
	}
	return metrics
}

func deriveIndicators(observation Observation, policy PolicyGraph, receipt Receipt) []Indicator {
	values := []bool{
		len(policy.Contradictions) == 0 && policy.Source.ParsedDigest != "" && policy.Source.SemanticDigest != "",
		len(receipt.Subjects) == len(observation.ChangedFiles),
		unknownSubjectsDescendFully(receipt.Subjects),
		validClaimTransitions(receipt.ClaimTransitions),
		repositoryWriteCount(observation.Isolation) == receipt.Operation.RepositoryWrites && receipt.Operation.ReadOnly == (receipt.Operation.RepositoryWrites == 0),
		receipt.ExecutionMode == "PLAN_ONLY",
	}
	result := make([]Indicator, 0, FixedIndicatorDenominator)
	for index, id := range indicatorIDs {
		observed := 0
		if values[index] {
			observed = 1
		}
		result = append(result, Indicator{ID: id, Observed: observed, Denominator: 1, Satisfied: values[index]})
	}
	return result
}

func unknownSubjectsDescendFully(subjects []SubjectResolution) bool {
	for _, subject := range subjects {
		if subject.Resolution != ResolutionUnknown {
			continue
		}
		if len(subject.SelectedChecks) != FixedCheckDenominator {
			return false
		}
		for index, choice := range subject.SelectedChecks {
			if choice.CheckID != fixedCheckIDs[index] || choice.ProofChoice != ProofFullDescent || len(choice.Reason) == 0 {
				return false
			}
		}
	}
	return true
}

func validClaimTransitions(values []ClaimTransition) bool {
	previous := ""
	for index, value := range values {
		if value.Sequence != index+1 || value.ClaimID == "" || value.SubjectPath == "" || value.EvidenceDigest == "" || value.Provenance == "" || value.PreviousDigest != previous {
			return false
		}
		if value.Before != ClaimOpen && value.Before != ClaimDischarged && value.Before != ClaimRefuted {
			return false
		}
		if value.After != ClaimOpen && value.After != ClaimDischarged && value.After != ClaimRefuted {
			return false
		}
		computed, err := transitionDigest(value)
		if err != nil || computed != value.Digest {
			return false
		}
		previous = value.Digest
	}
	return true
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func resolutionPlan(value SubjectResolution) string {
	switch value.Resolution {
	case ResolutionSelected:
		return PlanSelective
	case ResolutionUnknown:
		return PlanFull
	default:
		return PlanNone
	}
}

func subjectReason(value SubjectResolution) string {
	if value.Coordinate.Reason != "" {
		return value.Coordinate.Reason
	}
	return strings.Join([]string{value.Resolution, resolutionPlan(value)}, ":")
}
