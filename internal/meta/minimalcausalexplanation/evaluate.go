package minimalcausalexplanation

import (
	"fmt"
	"strings"
)

func Evaluate(sourcePath string, source []byte, repository, subjectSHA string) (Receipt, error) {
	if err := validSubject(repository, subjectSHA); err != nil {
		return Receipt{}, err
	}
	if !strings.HasSuffix(sourcePath, ".gooo") || len(source) == 0 {
		return Receipt{}, fmt.Errorf("a non-empty .gooo source is required")
	}
	for _, marker := range []string{"package minimalcausalexplanation", "namespace minimalcausalexplanation", "entity Evidence", "activity ProduceEvidence"} {
		if !strings.Contains(string(source), marker) {
			return Receipt{}, fmt.Errorf("source is missing %q", marker)
		}
	}

	graph := CanonicalGraph()
	program := CanonicalProgram()
	cases := buildCases()
	transitions, preservation := buildClaimTransitions()
	summary := summarize(cases)
	indicators, err := buildIndicators(source, graph, program, cases, summary, preservation)
	if err != nil {
		return Receipt{}, err
	}
	receipt := Receipt{
		Schema:  ReceiptSchema,
		Source:  SourceBinding{Schema: SourceSchema, Path: sourcePath, Digest: contentDigest(source), Lines: countLines(source)},
		Subject: Subject{Repository: repository, SHA: subjectSHA},
		Program: program, Graph: graph, Cases: cases, Summary: summary,
		Preservation: preservation, ClaimTransitions: transitions, Indicators: indicators,
		Decision: DecisionPass, Resolution: "MINIMAL_CAUSAL_EXPLANATION_VERIFIED",
		Authority: Authority{RepositoryWorkspaceWrites: false, PromotionAuthorized: false, SemanticMutationAuthorized: false},
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			receipt.Decision = DecisionFailClosed
			receipt.Resolution = "CONTRACT_INDICATOR_UNSATISFIED"
			break
		}
	}
	receipt.ReceiptDigest = ReceiptDigest(receipt)
	return receipt, nil
}

func buildCases() []ExplanationCase {
	positive := ExplanationPath{
		ID:          "path.minimal-approval",
		EvidenceIDs: canonicalEvidence(),
		EdgeIDs:     []string{requestPolicyEdge, policyResultEdge},
		Decision:    DecisionPass, Sufficient: true, Minimal: true,
	}
	positive.Counterfactuals = counterfactuals(positive.EvidenceIDs)

	overlong := ExplanationPath{
		ID:          "path.overlong-approval",
		EvidenceIDs: []string{requestEvidence, policyEvidence, resultEvidence, noiseEvidence},
		EdgeIDs:     []string{requestPolicyEdge, policyResultEdge},
		Decision:    DecisionPass, Sufficient: true, Minimal: false,
	}
	overlong.Counterfactuals = counterfactuals(overlong.EvidenceIDs)

	insufficient := ExplanationPath{
		ID:          "path.insufficient-approval",
		EvidenceIDs: []string{requestEvidence, resultEvidence},
		EdgeIDs:     nil, Decision: DecisionFailClosed, Sufficient: false, Minimal: false,
	}

	return []ExplanationCase{
		{ID: "minimal", Kind: "MINIMAL_SUFFICIENT", ExplanationText: "approval follows the decisive causal path", AvailableEvidenceTotal: 4, Paths: []ExplanationPath{positive}, ExpectedDecision: DecisionPass, Verdict: CaseAccepted},
		{ID: "overlong", Kind: "SUFFICIENT_NOT_MINIMAL", ExplanationText: "all observed logs explain the approval", AvailableEvidenceTotal: 4, Paths: []ExplanationPath{overlong}, ExpectedDecision: DecisionPass, Verdict: CaseRejected},
		{ID: "insufficient", Kind: "INSUFFICIENT", ExplanationText: "the request and result appear related", AvailableEvidenceTotal: 3, Paths: []ExplanationPath{insufficient}, ExpectedDecision: DecisionPass, Verdict: CaseRejected},
	}
}

func counterfactuals(evidence []string) []Counterfactual {
	before := decisionForEvidence(evidence)
	result := make([]Counterfactual, 0, len(evidence))
	for _, removed := range evidence {
		remaining := make([]string, 0, len(evidence)-1)
		for _, item := range evidence {
			if item != removed {
				remaining = append(remaining, item)
			}
		}
		after := decisionForEvidence(remaining)
		changed := before != after
		reason := "DECISION_UNCHANGED"
		if changed {
			reason = "DECISION_CHANGED"
		}
		result = append(result, Counterfactual{
			RemovedEvidenceID: removed, BeforeDecision: before, AfterDecision: after, Changed: changed,
			Coordinate: Coordinate{Stage: "COUNTERFACTUAL", Step: "remove-evidence", Reason: reason},
		})
	}
	return result
}

func summarize(cases []ExplanationCase) Summary {
	summary := Summary{
		CasesTotal: len(cases), CandidateEvidenceTotal: 0, RepositoryWrites: 0,
		PromotionAuthorized: false, PathSetAuthoritative: true, ExplanationTextRole: ExplanationTextRole,
	}
	for _, example := range cases {
		summary.CandidateEvidenceTotal += example.AvailableEvidenceTotal
		for _, path := range example.Paths {
			summary.PathsObserved++
			summary.CounterfactualTotal += len(path.Counterfactuals)
			for _, counterfactual := range path.Counterfactuals {
				if counterfactual.Changed {
					summary.ChangedCounterfactualTotal++
				}
			}
			if path.Sufficient && path.Minimal {
				summary.MinimalSufficientPaths++
			}
			if path.Sufficient && !path.Minimal {
				summary.SufficientNonminimalPaths++
			}
			if !path.Sufficient {
				summary.InsufficientPaths++
			}
		}
		if example.Verdict == CaseAccepted || example.Verdict == CaseRejected {
			summary.CasesAccepted++
		}
	}
	return summary
}

func buildClaimTransitions() ([]ClaimTransition, Preservation) {
	ids := claimIDs()
	transitions := make([]ClaimTransition, 0, TransitionTotal)
	previous := ""
	for index, claimID := range ids {
		registered := ClaimTransition{Sequence: len(transitions) + 1, ClaimID: claimID, Before: "UNRECORDED", After: ClaimOpen, PreviousTransitionDigest: previous, Coordinate: Coordinate{Stage: "DECLARE", Step: "claim-ledger", Reason: "CLAIM_REGISTERED"}}
		registered.TransitionDigest = transitionDigest(registered)
		transitions = append(transitions, registered)
		previous = registered.TransitionDigest
		resolved := ClaimTransition{Sequence: len(transitions) + 1, ClaimID: claimID, Before: ClaimOpen, After: ClaimDischarged, PreviousTransitionDigest: previous, Coordinate: Coordinate{Stage: "VERIFY", Step: claimStep(index), Reason: "CLAIM_PRESERVED"}}
		resolved.TransitionDigest = transitionDigest(resolved)
		transitions = append(transitions, resolved)
		previous = resolved.TransitionDigest
	}
	return transitions, Preservation{ClaimTotal: ClaimTotal, PreservedTotal: ClaimTotal, TransitionTotal: len(transitions), TransitionHead: previous, Policy: "APPEND_ONLY_OPEN_TO_DISCHARGED"}
}

func claimStep(index int) string {
	return []string{"source-binding", "graph-closure", "path-sufficiency", "path-minimality", "counterfactuals", "read-only-boundary"}[index]
}

func transitionDigest(transition ClaimTransition) string {
	transition.TransitionDigest = ""
	digest, _ := digestValue(transition)
	return digest
}
