package causalci

import (
	"fmt"
	"sort"
	"strings"
)

// Evaluate produces only a plan. Execution remains UNKNOWN until the
// independent consumer process writes an adjudication receipt.
func Evaluate(observationRaw []byte, sourcePath string, source []byte) (Receipt, error) {
	return evaluateWithBinding(observationRaw, sourcePath, source, "HEAD")
}

// EvaluateIntervention evaluates a supplied Gooo source as a semantic
// intervention. Its bytes are intentionally not asserted to be HEAD bytes;
// the receipt still records the observed HEAD coordinate for comparison.
func EvaluateIntervention(observationRaw []byte, sourcePath string, source []byte) (Receipt, error) {
	return evaluateWithBinding(observationRaw, sourcePath, source, "INTERVENTION")
}

func evaluateWithBinding(observationRaw []byte, sourcePath string, source []byte, bindingMode string) (Receipt, error) {
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
	policy.Source.BindingMode = bindingMode
	policy.Source.ObservedCheckoutSHA = observation.ObservedCheckoutSHA
	policy.Source.HeadPathObjectID = observation.HeadPathObjectID
	policy.Source.SourceBytesDigest = digestBytes(source)
	if bindingMode == "HEAD" && (observation.ObservedCheckoutSHA != observation.HeadSHA || observation.SourceBytesDigest != digestBytes(source) || observation.HeadPathObjectID != GitBlobObjectID(source)) {
		policy.Contradictions = append([]PolicyContradiction{{Stage: "SOURCE_BINDING", Step: "validate-exact-head", Reason: ReasonSourceBinding}}, policy.Contradictions...)
	}

	receipt := Receipt{
		Schema: ReceiptSchema, Scope: ReceiptScope, Source: policy.Source,
		ObservationDigest: digestBytes(observationRaw), Operation: deriveOperation(observation),
		ExecutionMode:       CapabilityPlanOnly,
		Execution:           ExecutionStatus{Result: ExecutionUnknown, Capability: CapabilityPlanOnly, Coordinate: Coordinate{Stage: "ADJUDICATION", Step: "await-consumer", Reason: "CONSUMER_PROCESS_NOT_RUN"}},
		CheckInventory:      ExactInventory{ExpectedIDs: fixedCheckIDSlice()},
		IndicatorInventory:  ExactInventory{ExpectedIDs: indicatorIDSlice()},
		IndependentVerifier: IndependentVerifier{ID: "gooo://consumer/causal-ci-selection", Mode: "INDEPENDENT_RECONSTRUCTION", Required: true, Capability: "SEPARATE_PROCESS"},
	}
	receipt.Conformance = conformanceFor(policy)
	receipt.Subjects = evaluateSubjects(observation, policy)
	receipt.ClaimTransitions = appendClaimTransitions(observation, policy, receipt.Subjects, receipt.ObservationDigest)
	receipt.Metrics = deriveMetrics(observation, policy, receipt)
	receipt.CheckInventory.ObservedIDs = checkIDs(policy.Checks)
	receipt.IndicatorInventory.ObservedIDs = indicatorIDsFromPolicy(receipt.Indicators)
	receipt.Indicators = deriveIndicators(observation, policy, receipt)
	receipt.IndicatorInventory.ObservedIDs = indicatorIDsFromPolicy(receipt.Indicators)
	receipt.Metrics.FixedIndicatorSatisfied = satisfiedIndicators(receipt.Indicators)
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
	return Operation{
		Producer: "gooo://producer/causal-ci-selection", Consumer: "gooo://consumer/causal-ci-selection",
		MetaOperation: "causal-ci-select", ProofChoice: ProofCausalPath,
		DeclaredPlanCapability: CapabilityPlanOnly, ObservedRepositoryState: repositoryState(observation.Isolation),
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
			result = append(result, SubjectResolution{Path: file.Path, Resolution: ResolutionFailClosed, Coordinate: Coordinate{Stage: contradiction.Stage, Step: contradiction.Step, Reason: contradiction.Reason}, SelectedChecks: []CheckChoice{}})
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
	pathEvidence := PathEvidence{SubjectPath: path, ClaimIDs: []string{policy.ClaimID}, Proposition: ReasonCompleteRoute, SurfaceID: policy.SurfaceID, CheckID: checkID, PolicyEdgeIDs: []string{first[0].ID, second[0].ID, third[0].ID}, SemanticDigest: policy.Source.SemanticDigest, Explanation: "changed-file observation traverses claim-to-surface and surface-to-check semantic policy", ProofChoice: ProofCausalPath}
	return SubjectResolution{Path: path, Resolution: ResolutionSelected, Coordinate: Coordinate{Stage: stageSubject, Step: stepSelectChecks, Reason: ReasonCompletePath}, Paths: []PathEvidence{pathEvidence}, SelectedChecks: []CheckChoice{{CheckID: checkID, ProofChoice: ProofCausalPath, Reason: ReasonCompletePath, ClaimIDs: []string{policy.ClaimID}, PathIDs: pathEvidence.PolicyEdgeIDs}}}
}

func unknownSubject(file ChangedFileObservation, sourcePath string) SubjectResolution {
	reason := ReasonUnknownSubject
	if file.Path == sourcePath && file.Status == "D" {
		reason = "SOURCE_OBJECT_NOT_AVAILABLE"
	}
	unknown := UnknownCause{SubjectPath: file.Path, Coordinate: Coordinate{Stage: stageSubject, Step: stepObserveSubject, Reason: reason}, Provenance: "git://pull-request/changed-file/" + file.Path}
	choices := make([]CheckChoice, 0, FixedCheckDenominator)
	for _, id := range fixedCheckIDs {
		choices = append(choices, CheckChoice{CheckID: id, ProofChoice: ProofFullDescent, Reason: reason})
	}
	return SubjectResolution{Path: file.Path, Resolution: ResolutionUnknown, Coordinate: Coordinate{Stage: stageSubject, Step: stepDescendFull, Reason: reason}, UnknownCauses: []UnknownCause{unknown}, SelectedChecks: choices}
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
		after, resolution, reason := claim.State, PlanNone, ReasonClaimLowered
		stage, step := stageClaimLedger, stepClaimTransition
		if len(policy.Contradictions) > 0 {
			if claim.Proposition == ReasonCompleteRoute {
				after, reason, stage, step = ClaimRefuted, ReasonClaimRefuted, stageConformance, stepValidatePolicy
			} else {
				reason = ReasonUnrelatedContradiction
			}
		} else if subject, exists := byPath[claim.SubjectPath]; exists {
			switch subject.Resolution {
			case ResolutionSelected:
				resolution = PlanSelective
				if claim.Proposition == ReasonCompleteRoute {
					after, reason = ClaimDischarged, ReasonClaimDischarged
				} else {
					after, reason = ClaimOpen, ReasonPlanOnlyOpen
				}
			case ResolutionUnknown:
				resolution = PlanFull
				switch claim.State {
				case ClaimOpen:
					reason = ReasonClaimLowered
				case ClaimDischarged:
					reason = ReasonUnknownDischarged
				case ClaimRefuted:
					reason = ReasonUnknownRefuted
				}
			case ResolutionFailClosed:
				if claim.Proposition == ReasonCompleteRoute {
					after, reason = ClaimRefuted, ReasonClaimRefuted
				} else {
					reason = ReasonUnrelatedContradiction
				}
			}
		}
		if claim.State == ClaimRefuted && after != ClaimRefuted {
			after = ClaimRefuted
		}
		evidence, _ := digestJSON(struct {
			Observation string                `json:"observation"`
			Source      string                `json:"source"`
			Claim       PriorClaimObservation `json:"claim"`
			After       string                `json:"after"`
			Resolution  string                `json:"resolution"`
		}{observationDigest, policy.Source.SemanticDigest, claim, after, resolution})
		transition := ClaimTransition{Sequence: index + 1, TemplateID: claim.TemplateID, ClaimID: claim.InstanceID, SubjectPath: claim.SubjectPath, Proposition: claim.Proposition, Before: claim.State, After: after, Resolution: resolution, Stage: stage, Step: step, Reason: reason, EvidenceDigest: evidence, Provenance: claim.Provenance, PreviousDigest: previous}
		transition.Digest, _ = transitionDigest(transition)
		result = append(result, transition)
		previous = transition.Digest
	}
	return result
}

func deriveMetrics(observation Observation, policy PolicyGraph, receipt Receipt) Metrics {
	paths := make([]string, 0, len(observation.ChangedFiles))
	for _, file := range sortedChangedFiles(observation.ChangedFiles) {
		paths = append(paths, file.Path)
	}
	universe, _ := digestJSON(paths)
	metrics := Metrics{SubjectUniverseDigest: universe, SubjectUniverseCount: len(paths), SubjectCoverageNumerator: len(receipt.Subjects), SubjectCoverageDenominator: len(paths), SubjectTotal: len(receipt.Subjects), FullSuiteCheckDenominator: len(policy.Checks), ClaimTransitionTotal: len(receipt.ClaimTransitions), FixedIndicatorDenominator: FixedIndicatorDenominator}
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
	state := receipt.Operation.ObservedRepositoryState
	values := []bool{
		len(policy.Contradictions) == 0 && policy.Source.ParsedDigest != "" && policy.Source.SemanticDigest != "",
		len(receipt.Subjects) == len(observation.ChangedFiles),
		unknownSubjectsDescendFully(receipt.Subjects),
		validClaimTransitions(receipt.ClaimTransitions),
		state.NetState == ObservedStateUnchanged && state.ChangedPathCount == 0 && state.ChangedContentCount == 0 && state.TransientWrites == ObservedUnknown && state.GlobalMutationAuthority == ObservedUnknown,
		receipt.ExecutionMode == CapabilityPlanOnly && receipt.Execution.Result == ExecutionUnknown,
	}
	result := make([]Indicator, 0, FixedIndicatorDenominator)
	for index, id := range indicatorIDs {
		result = append(result, Indicator{ID: id, Observed: boolInt(values[index]), Denominator: 1, Satisfied: values[index]})
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
			if choice.CheckID != fixedCheckIDs[index] || choice.ProofChoice != ProofFullDescent || choice.Reason == "" {
				return false
			}
		}
	}
	return true
}

func validClaimTransitions(values []ClaimTransition) bool {
	previous := ""
	for index, value := range values {
		if value.Sequence != index+1 || value.ClaimID == "" || value.TemplateID == "" || value.SubjectPath == "" || value.Proposition == "" || value.EvidenceDigest == "" || value.Provenance == "" || value.PreviousDigest != previous || value.ClaimID != ClaimInstanceID(value.TemplateID, value.SubjectPath, value.Proposition) {
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

func fixedCheckIDSlice() []string {
	result := make([]string, len(fixedCheckIDs))
	copy(result, fixedCheckIDs[:])
	return result
}

func indicatorIDSlice() []string {
	result := make([]string, len(indicatorIDs))
	copy(result, indicatorIDs[:])
	return result
}

func checkIDs(values []Check) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func indicatorIDsFromPolicy(values []Indicator) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.ID)
	}
	return result
}

func satisfiedIndicators(values []Indicator) int {
	result := 0
	for _, value := range values {
		if value.Satisfied {
			result++
		}
	}
	return result
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
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
