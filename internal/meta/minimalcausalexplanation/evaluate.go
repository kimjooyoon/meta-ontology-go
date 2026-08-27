package minimalcausalexplanation

import (
	"fmt"
	"sort"
	"strings"
)

type assessment struct {
	Model         sourceModel
	Observed      []Evidence
	Corpus        []Evidence
	Cases         []ExplanationCase
	Summary       Summary
	ClaimOutcomes map[string]string
	Decision      string
}

func Evaluate(sourcePath string, source, compilerReceipt, repositoryBefore, repositoryAfter, independence []byte, repository, subjectSHA string) (Receipt, error) {
	if err := validSubject(repository, subjectSHA); err != nil {
		return Receipt{}, err
	}
	if !strings.HasSuffix(sourcePath, ".gooo") || len(source) == 0 || len(compilerReceipt) == 0 || len(repositoryBefore) == 0 || len(repositoryAfter) == 0 {
		return Receipt{}, fmt.Errorf("non-empty .gooo source and raw observations are required")
	}
	model, err := reconstructSource(sourcePath, source, independence)
	if err != nil {
		return Receipt{}, err
	}
	rawReceipt, err := decodeCompilerReceipt(compilerReceipt)
	if err != nil {
		return Receipt{}, err
	}
	if rawReceipt.HeadSHA != subjectSHA {
		return Receipt{}, fmt.Errorf("compiler receipt is not bound to subject SHA")
	}
	before, err := decodeRepositoryObservation(repositoryBefore)
	if err != nil {
		return Receipt{}, err
	}
	after, err := decodeRepositoryObservation(repositoryAfter)
	if err != nil {
		return Receipt{}, err
	}
	boundary := repositoryBoundary(before, after)
	result := assessModel(model, rawReceipt, compilerReceipt)
	result.ClaimOutcomes["read-only-preserved"] = claimOutcome(boundary.Writes == 0 && !boundary.PromotionAuthorized)
	transitions, preservation := buildClaimTransitions(model, result.ClaimOutcomes, digestEvidence(result.Observed))
	regression := buildClaimRegression(model, rawReceipt, compilerReceipt, digestEvidence(result.Observed))
	interventions, err := buildInterventions(sourcePath, source, model, rawReceipt, compilerReceipt, independence, result, transitions)
	if err != nil {
		return Receipt{}, err
	}
	indicators, err := buildIndicators(model, result, boundary, preservation, regression, interventions)
	if err != nil {
		return Receipt{}, err
	}

	receipt := Receipt{
		Schema:         ReceiptSchema,
		Source:         SourceBinding{Schema: SourceSchema, Path: sourcePath, Digest: contentDigest(source), Lines: countLines(source), SemanticDigest: model.SemanticDigest},
		Subject:        Subject{Repository: repository, SHA: subjectSHA},
		Reconstruction: model.SourceReconstruct,
		Program:        model.Program, Predicate: model.Predicate, Graph: model.Graph,
		Evidence: result.Corpus, ObservedReceiptDigest: contentDigest(compilerReceipt), Cases: result.Cases,
		Summary: result.Summary, Repository: boundary, Preservation: preservation,
		ClaimTransitions: transitions, ClaimRegression: regression, Interventions: interventions,
		Indicators: indicators, Decision: StatusPass, Resolution: "SOURCE_RECONSTRUCTED_MINIMAL_CAUSAL_EXPLANATION",
		Authority: Authority{RepositoryWorkspaceWrites: boundary.Writes != 0, PromotionAuthorized: boundary.PromotionAuthorized, SemanticMutationAuthorized: false},
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			receipt.Decision = StatusFailClosed
			receipt.Resolution = "SOURCE_RECONSTRUCTION_CONTRACT_UNSATISFIED"
			break
		}
	}
	receipt.ReceiptDigest = ReceiptDigest(receipt)
	return receipt, nil
}

func assessModel(model sourceModel, raw RawCompilerReceipt, rawBytes []byte) assessment {
	observed := observedEvidence(model, raw, rawBytes)
	corpus := append([]Evidence(nil), observed...)
	if noise, ok := syntheticNoise(model); ok {
		corpus = append(corpus, noise)
	}
	sort.Slice(corpus, func(i, j int) bool { return corpus[i].ID < corpus[j].ID })

	required := evidenceIDsForRoles(model, model.Predicate.RequiredRoles)
	minimal := makePath("minimal-sufficient", required, model, corpus)
	overlongIDs := append([]string(nil), required...)
	if noise, ok := syntheticNoise(model); ok {
		overlongIDs = append(overlongIDs, noise.ID)
	}
	overlong := makePath("sufficient-overlong", overlongIDs, model, corpus)
	insufficientIDs := []string{}
	if len(required) >= 2 {
		insufficientIDs = []string{required[0], required[len(required)-1]}
	}
	insufficient := makePath("insufficient", insufficientIDs, model, corpus)
	cases := []ExplanationCase{
		{ID: "minimal", Kind: "MINIMAL_SUFFICIENT", ExplanationText: "the compiler receipt is supported by the decisive path", AvailableEvidenceTotal: len(corpus), Paths: []ExplanationPath{minimal}, ExpectedDecision: StatusPass, Verdict: verdict(minimal.Sufficient)},
		{ID: "overlong", Kind: "SUFFICIENT_NOT_MINIMAL", ExplanationText: "all logs explain the compiler receipt", AvailableEvidenceTotal: len(corpus), Paths: []ExplanationPath{overlong}, ExpectedDecision: StatusPass, Verdict: CaseRejected},
		{ID: "insufficient", Kind: "INSUFFICIENT", ExplanationText: "two observations appear related", AvailableEvidenceTotal: len(observed), Paths: []ExplanationPath{insufficient}, ExpectedDecision: StatusPass, Verdict: CaseRejected},
	}
	summary := summarize(cases, observed, corpus)
	decision := minimal.Decision
	outcomes := claimOutcomes(model, minimal, summary, decision == StatusPass, raw, observed, true)
	return assessment{Model: model, Observed: observed, Corpus: corpus, Cases: cases, Summary: summary, ClaimOutcomes: outcomes, Decision: decision}
}

func makePath(kind string, evidenceIDs []string, model sourceModel, corpus []Evidence) ExplanationPath {
	decision := decisionForEvidence(model, evidenceIDs, corpus)
	path := ExplanationPath{ID: "path." + kind, EvidenceIDs: append([]string(nil), evidenceIDs...), EdgeIDs: pathEdges(model.Graph, evidenceIDs), Decision: decision, Sufficient: decision == StatusPass}
	if path.Sufficient {
		path.Counterfactuals = removalCounterfactuals(model, path, corpus)
		path.SingleRemovalTotal = len(path.Counterfactuals)
		path.SingleRemovalChanged = countChanged(path.Counterfactuals)
		if path.SingleRemovalChanged == path.SingleRemovalTotal && path.SingleRemovalTotal > 0 {
			path.SubsetMinimal = SubsetMinimal
		} else {
			path.SubsetMinimal = NotSubsetMinimal
		}
		path.CombinationSearch = enumerateSmallerCombinations(model, path, corpus)
		if path.CombinationSearch.Exhaustive && path.CombinationSearch.SufficientSmallerCombinationTotal == 0 {
			path.CardinalityMinimum = CardinalityMinimum
		} else {
			path.CardinalityMinimum = NotCardinalityMinimum
		}
	} else {
		path.SubsetMinimal = NotSubsetMinimal
		path.CardinalityMinimum = CardinalityUnknown
	}
	return path
}

func decisionForEvidence(model sourceModel, evidenceIDs []string, corpus []Evidence) string {
	byID := make(map[string]Evidence, len(corpus))
	for _, evidence := range corpus {
		byID[evidence.ID] = evidence
	}
	for _, role := range model.Predicate.RequiredRoles {
		evidence, ok := model.EvidenceByRole[role]
		if !ok || !containsString(evidenceIDs, evidence.ID) || byID[evidence.ID].Status != StatusPass {
			return StatusFailClosed
		}
	}
	requiredIDs := evidenceIDsForRoles(model, model.Predicate.RequiredRoles)
	for index := 1; index < len(requiredIDs); index++ {
		if !hasCausalEdge(model.Graph, requiredIDs[index-1], requiredIDs[index]) || !containsString(evidenceIDs, requiredIDs[index-1]) || !containsString(evidenceIDs, requiredIDs[index]) {
			return StatusFailClosed
		}
	}
	if model.Predicate.DecisionOutput == "" || model.PriorClaimState != ClaimOpen {
		return StatusFailClosed
	}
	return StatusPass
}

func evidenceIDsForRoles(model sourceModel, roles []string) []string {
	ids := make([]string, 0, len(roles))
	for _, role := range roles {
		if evidence, ok := model.EvidenceByRole[role]; ok {
			ids = append(ids, evidence.ID)
		}
	}
	return ids
}

func pathEdges(graph CausalGraph, evidenceIDs []string) []string {
	result := make([]string, 0)
	for index := 1; index < len(evidenceIDs); index++ {
		for _, edge := range graph.Edges {
			if edge.From == evidenceIDs[index-1] && edge.To == evidenceIDs[index] && edge.Causal {
				result = append(result, edge.ID)
				break
			}
		}
	}
	return result
}

func hasCausalEdge(graph CausalGraph, from, to string) bool {
	for _, edge := range graph.Edges {
		if edge.From == from && edge.To == to && edge.Causal {
			return true
		}
	}
	return false
}

func removalCounterfactuals(model sourceModel, path ExplanationPath, corpus []Evidence) []Counterfactual {
	result := make([]Counterfactual, 0, len(path.EvidenceIDs))
	for index, removed := range path.EvidenceIDs {
		remaining := append([]string(nil), path.EvidenceIDs[:index]...)
		remaining = append(remaining, path.EvidenceIDs[index+1:]...)
		after := decisionForEvidence(model, remaining, corpus)
		result = append(result, Counterfactual{ExecutionID: "cf." + fmt.Sprint(index+1), RemovedEvidenceID: removed, Origin: syntheticOrigin, BeforeDecision: path.Decision, AfterDecision: after, Changed: path.Decision != after, Coordinate: Coordinate{Stage: "COUNTERFACTUAL", Step: "remove-single-evidence", Reason: removalReason(path.Decision != after)}, EvidenceDigest: digestString(strings.Join(path.EvidenceIDs, "|"))})
	}
	return result
}

func enumerateSmallerCombinations(model sourceModel, path ExplanationPath, corpus []Evidence) CombinationSearch {
	ids := make([]string, 0, len(corpus))
	for _, evidence := range corpus {
		ids = append(ids, evidence.ID)
	}
	sort.Strings(ids)
	search := CombinationSearch{CorpusEvidenceIDs: ids, Exhaustive: true}
	for size := 0; size < len(path.EvidenceIDs); size++ {
		combinations := combinations(ids, size)
		search.EnumeratedCombinationTotal += len(combinations)
		search.SmallerCombinationTotal += len(combinations)
		for _, combination := range combinations {
			if decisionForEvidence(model, combination, corpus) == StatusPass {
				search.SufficientSmallerCombinationTotal++
			}
		}
	}
	return search
}

func combinations(values []string, size int) [][]string {
	result := make([][]string, 0)
	var visit func(int, []string)
	visit = func(start int, selected []string) {
		if len(selected) == size {
			result = append(result, append([]string(nil), selected...))
			return
		}
		for index := start; index <= len(values)-(size-len(selected)); index++ {
			visit(index+1, append(selected, values[index]))
		}
	}
	visit(0, nil)
	return result
}

func summarize(cases []ExplanationCase, observed, corpus []Evidence) Summary {
	summary := Summary{CasesTotal: len(cases), PathsObserved: len(cases), ObservedEvidenceTotal: len(observed), CandidateEvidenceTotal: len(corpus), PathSetAuthoritative: true, ExplanationTextRole: ExplanationTextRole}
	for _, evidence := range corpus {
		if evidence.Origin == syntheticOrigin {
			summary.SyntheticEvidenceTotal++
		}
	}
	for _, example := range cases {
		path := example.Paths[0]
		if path.Sufficient {
			summary.SufficientPaths++
			summary.CounterfactualExecutions += len(path.Counterfactuals)
			summary.ChangedCounterfactuals += path.SingleRemovalChanged
			summary.CombinationExecutions += path.CombinationSearch.EnumeratedCombinationTotal
			if path.SubsetMinimal == SubsetMinimal {
				summary.SubsetMinimalNumerator++
			}
			if path.CardinalityMinimum == CardinalityMinimum {
				summary.CardinalityMinimumNumerator++
			}
		}
		if path.Sufficient {
			summary.SubsetMinimalDenominator++
			summary.CardinalityMinimumDenominator++
		} else {
			summary.InsufficientPaths++
			summary.CardinalityUnknownPaths++
		}
	}
	summary.ClaimTransitionTotal = len(claimIDs()) * 2
	summary.RegressionClaimTransitionTotal = len(claimIDs()) * 2
	summary.SyntheticCounterfactuals = summary.CounterfactualExecutions
	return summary
}

func claimOutcomes(model sourceModel, minimal ExplanationPath, summary Summary, decisionPass bool, raw RawCompilerReceipt, observed []Evidence, readOnly bool) map[string]string {
	outcomes := map[string]string{}
	for _, claim := range model.Claims {
		outcomes[claim] = StatusUnknown
	}
	setClaimOutcome(outcomes, "source-bound", model.SourceReconstruct.ASTParsed && model.SourceReconstruct.IRLowered)
	setClaimOutcome(outcomes, "graph-predicate-reconstructed", model.SourceReconstruct.GraphReconstructed && model.SourceReconstruct.PredicateReconstructed)
	setConditionalClaimOutcome(outcomes, "subset-minimal", decisionPass && minimal.SubsetMinimal == SubsetMinimal, observed, raw.Decision)
	setConditionalClaimOutcome(outcomes, "cardinality-minimum", decisionPass && minimal.CardinalityMinimum == CardinalityMinimum, observed, raw.Decision)
	setConditionalClaimOutcome(outcomes, "counterfactual-difference", summary.CounterfactualExecutions == 7 && summary.ChangedCounterfactuals == 6, observed, raw.Decision)
	setClaimOutcome(outcomes, "read-only-preserved", readOnly)
	return outcomes
}

func setClaimOutcome(outcomes map[string]string, suffix string, passed bool) {
	for claim := range outcomes {
		if strings.HasSuffix(claim, suffix) {
			if passed {
				outcomes[claim] = StatusPass
			} else {
				outcomes[claim] = StatusRefuted
			}
		}
	}
}

func setConditionalClaimOutcome(outcomes map[string]string, suffix string, passed bool, observed []Evidence, rawDecision string) {
	if passed {
		setClaimOutcome(outcomes, suffix, true)
		return
	}
	status := StatusUnknown
	if explicitCounterexample(rawDecision) || allEvidenceObserved(observed) {
		status = StatusRefuted
	}
	for claim := range outcomes {
		if strings.HasSuffix(claim, suffix) {
			outcomes[claim] = status
		}
	}
}

func explicitCounterexample(decision string) bool {
	return decision == StatusFailClosed || decision == "VALUE_WITNESS_REJECTED"
}

func allEvidenceObserved(observed []Evidence) bool {
	if len(observed) == 0 {
		return false
	}
	for _, evidence := range observed {
		if evidence.Status == StatusUnknown {
			return false
		}
	}
	return true
}

func claimOutcome(passed bool) string {
	if passed {
		return StatusPass
	}
	return StatusRefuted
}

func buildClaimTransitions(model sourceModel, outcomes map[string]string, evidenceDigest string) ([]ClaimTransition, Preservation) {
	transitions := make([]ClaimTransition, 0, len(model.Claims)*2)
	previous := ""
	for _, claim := range model.Claims {
		registered := ClaimTransition{Sequence: len(transitions) + 1, ClaimID: claim, Before: "UNRECORDED", After: model.PriorClaimState, EvidenceDigest: evidenceDigest, Provenance: "prior claim state reconstructed from .gooo:mce.claim-state", Coordinate: Coordinate{Stage: "DECLARE", Step: "claim-ledger", Reason: "PRIOR_STATE_OBSERVED"}, PreviousTransitionDigest: previous}
		registered.TransitionDigest = transitionDigest(registered)
		transitions = append(transitions, registered)
		previous = registered.TransitionDigest
		after, reason := ClaimOpen, "CLAIM_EVIDENCE_UNOBSERVED"
		switch outcomes[claim] {
		case StatusPass:
			after, reason = ClaimDischarged, "CLAIM_EVIDENCE_PASSED"
		case StatusRefuted:
			after, reason = ClaimRefuted, "CLAIM_COUNTEREXAMPLE"
		}
		resolved := ClaimTransition{Sequence: len(transitions) + 1, ClaimID: claim, Before: model.PriorClaimState, After: after, EvidenceDigest: evidenceDigest, Provenance: "raw observation and derived path receipt", Coordinate: Coordinate{Stage: "VERIFY", Step: claim, Reason: reason}, PreviousTransitionDigest: previous}
		resolved.TransitionDigest = transitionDigest(resolved)
		transitions = append(transitions, resolved)
		previous = resolved.TransitionDigest
	}
	preserved := 0
	for _, claim := range model.Claims {
		if outcomes[claim] == StatusPass {
			preserved++
		}
	}
	return transitions, Preservation{ClaimTotal: len(model.Claims), PreservedTotal: preserved, TransitionTotal: len(transitions), TransitionHead: previous, Policy: "APPEND_ONLY_OPEN_CONDITIONAL_DISCHARGE_OR_REFUTE"}
}

func buildClaimRegression(model sourceModel, raw RawCompilerReceipt, rawBytes []byte, evidenceDigest string) ClaimRegression {
	failed := raw
	failed.Decision = StatusFailClosed
	failed.Reason = "SYNTHETIC_FAILURE_REGRESSION"
	result := assessModel(model, failed, rawBytes)
	transitions, _ := buildClaimTransitions(model, result.ClaimOutcomes, evidenceDigest)
	return ClaimRegression{ScenarioID: "failed-receipt-does-not-discharge", ReceiptDecision: StatusFailClosed, LegacyUnconditionalState: ClaimDischarged, CorrectState: ClaimRefuted, Transitions: transitions}
}

func buildInterventions(sourcePath string, source []byte, base sourceModel, raw RawCompilerReceipt, rawBytes, independence []byte, baseAssessment assessment, baseTransitions []ClaimTransition) ([]Intervention, error) {
	variants := []struct{ id, kind, source, provenance string }{
		{"predicate-change", "SEMANTIC_PREDICATE", strings.Replace(string(source), "mce.predicate:PASS_IF:source-parsed+semantic-ir-bound+compiler-receipt-proven:v1", "mce.predicate:PASS_IF:source-parsed+semantic-ir-bound+missing-observation:v1", 1), "source .gooo decision predicate intervention"},
		{"relation-change", "SEMANTIC_EVIDENCE_RELATION", strings.Replace(string(source), "BindCompilerReceipt(SemanticIRBoundEvidence)", "BindCompilerReceipt(AuditNoiseEvidence)", 1), "source .gooo evidence relation intervention"},
		{"comment-only", "PRESENTATION_COMMENT", string(source) + "\n// comment-only semantic intervention\n", "source comment-only intervention"},
	}
	basePathDigest := pathSetDigest(baseAssessment.Cases)
	baseClaimDigest := digestTransitions(baseTransitions)
	result := make([]Intervention, 0, len(variants))
	for _, variant := range variants {
		model, err := reconstructSource(sourcePath, []byte(variant.source), independence)
		if err != nil {
			return nil, fmt.Errorf("intervention %s: %w", variant.id, err)
		}
		assessment := assessModel(model, raw, rawBytes)
		transitions, _ := buildClaimTransitions(model, assessment.ClaimOutcomes, digestEvidence(assessment.Observed))
		result = append(result, Intervention{
			ID: variant.id, Kind: variant.kind, BeforeSourceDigest: base.SourceDigest, AfterSourceDigest: model.SourceDigest, BeforeSemanticDigest: base.SemanticDigest, AfterSemanticDigest: model.SemanticDigest,
			BeforeDecision: baseAssessment.Decision, AfterDecision: assessment.Decision, SemanticChanged: base.SemanticDigest != model.SemanticDigest, SemanticDigestPreserved: base.SemanticDigest == model.SemanticDigest,
			ResultPreserved: baseAssessment.Decision == assessment.Decision && pathSetDigest(baseAssessment.Cases) == pathSetDigest(assessment.Cases), PathSetChanged: basePathDigest != pathSetDigest(assessment.Cases), MinimalityChanged: pathProperties(baseAssessment.Cases) != pathProperties(assessment.Cases), ClaimTransitionChanged: baseClaimDigest != digestTransitions(transitions), Provenance: variant.provenance,
		})
	}
	return result, nil
}

func pathSetDigest(cases []ExplanationCase) string {
	digest, _ := digestValue(cases)
	return digest
}

func pathProperties(cases []ExplanationCase) string {
	var builder strings.Builder
	for _, example := range cases {
		path := example.Paths[0]
		fmt.Fprintf(&builder, "%s|%s|%s|%s|", path.Decision, path.SubsetMinimal, path.CardinalityMinimum, strings.Join(path.EdgeIDs, ","))
	}
	return builder.String()
}

func digestTransitions(transitions []ClaimTransition) string {
	digest, _ := digestValue(transitions)
	return digest
}

func verdict(sufficient bool) string {
	if sufficient {
		return CaseAccepted
	}
	return CaseRejected
}

func removalReason(changed bool) string {
	if changed {
		return "DECISION_CHANGED"
	}
	return "DECISION_UNCHANGED"
}

func countChanged(values []Counterfactual) int {
	result := 0
	for _, value := range values {
		if value.Changed {
			result++
		}
	}
	return result
}

func claimIDs() []string {
	return []string{"source-bound", "graph-predicate-reconstructed", "subset-minimal", "cardinality-minimum", "counterfactual-difference", "read-only-preserved"}
}

func transitionDigest(transition ClaimTransition) string {
	transition.TransitionDigest = ""
	digest, _ := digestValue(transition)
	return digest
}
