package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	explanation "github.com/kimjooyoon/meta-ontology-go/internal/meta/minimalcausalexplanation"
)

var subjectPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

const (
	requestEvidence = "evidence.request.accepted"
	policyEvidence  = "evidence.policy.allowed"
	resultEvidence  = "evidence.result.matches"
	noiseEvidence   = "evidence.audit.noise"
)

func Judge(receipt explanation.Receipt) (explanation.Judgment, error) {
	if err := validateReceipt(receipt); err != nil {
		return explanation.Judgment{}, err
	}
	judgment := explanation.Judgment{
		Schema: explanation.JudgmentSchema, Status: "VERIFIED", Decision: explanation.DecisionPass,
		Resolution: "INDEPENDENT_PATH_AND_COUNTERFACTUAL_JUDGE", PathSetVerified: true,
		CounterfactualsVerified: true, ClaimsPreserved: true,
		PromotionAuthorized: false, ReceiptDigest: independentReceiptDigest(receipt),
	}
	judgment.JudgmentDigest = digest(judgment)
	return judgment, nil
}

func validateReceipt(receipt explanation.Receipt) error {
	if receipt.Schema != explanation.ReceiptSchema || receipt.ReceiptDigest != independentReceiptDigest(receipt) {
		return fmt.Errorf("receipt schema or digest is invalid")
	}
	if receipt.Source.Schema != explanation.SourceSchema || !strings.HasSuffix(receipt.Source.Path, ".gooo") || receipt.Source.Digest == "" || receipt.Source.Lines == 0 {
		return fmt.Errorf("source binding is invalid")
	}
	if strings.TrimSpace(receipt.Subject.Repository) == "" || !subjectPattern.MatchString(receipt.Subject.SHA) {
		return fmt.Errorf("subject binding is invalid")
	}
	if err := validateGraph(receipt.Graph); err != nil {
		return err
	}
	if err := validateProgram(receipt.Program); err != nil {
		return err
	}
	if err := validateCases(receipt.Cases); err != nil {
		return err
	}
	if err := validateSummary(receipt.Summary); err != nil {
		return err
	}
	if err := validatePreservation(receipt); err != nil {
		return err
	}
	if err := validateIndicators(receipt.Indicators, receipt.Program); err != nil {
		return err
	}
	if receipt.Decision != explanation.DecisionPass || receipt.Resolution != "MINIMAL_CAUSAL_EXPLANATION_VERIFIED" {
		return fmt.Errorf("receipt decision is not verified")
	}
	if receipt.Authority.RepositoryWorkspaceWrites || receipt.Authority.PromotionAuthorized || receipt.Authority.SemanticMutationAuthorized {
		return fmt.Errorf("receipt claims mutation authority")
	}
	return nil
}

func validateGraph(graph explanation.CausalGraph) error {
	expected := explanation.CausalGraph{
		Schema:       explanation.GraphSchema,
		DecisionRule: "PASS iff request.accepted AND policy.allowed AND result.matches are present",
		Nodes: []explanation.CausalNode{
			{ID: requestEvidence, Role: "DECISION_INPUT", Producer: "request-observer", Consumer: "causal-evaluator"},
			{ID: policyEvidence, Role: "DECISION_INPUT", Producer: "policy-checker", Consumer: "causal-evaluator"},
			{ID: resultEvidence, Role: "DECISION_INPUT", Producer: "result-observer", Consumer: "causal-evaluator"},
			{ID: noiseEvidence, Role: "NON_CAUSAL_LOG", Producer: "audit-sampler", Consumer: "audit-archive"},
		},
		Edges: []explanation.CausalEdge{
			{ID: "edge.request.policy", From: requestEvidence, To: policyEvidence, Relation: "ENABLES"},
			{ID: "edge.policy.result", From: policyEvidence, To: resultEvidence, Relation: "CONSTRAINS"},
		},
	}
	expected.Digest = graphDigest(expected)
	if !equalJSON(graph, expected) {
		return fmt.Errorf("causal graph contract changed")
	}
	return nil
}

func validateProgram(program explanation.MetaProgram) error {
	expected := []explanation.MetaOperation{
		{ID: "bind-source", Activity: "BindGoooSource", Producer: "source-reader", Consumer: "causal-evaluator", ProofChoice: "FOUNDATION"},
		{ID: "freeze-graph", Activity: "FreezeCausalGraph", Producer: "causal-evaluator", Consumer: "path-checker", ProofChoice: "FOUNDATION"},
		{ID: "evaluate-sufficiency", Activity: "EvaluatePathSufficiency", Producer: "path-checker", Consumer: "counterfactual-checker", ProofChoice: "COHERENCE"},
		{ID: "minimize-path", Activity: "MinimizeByRemoval", Producer: "counterfactual-checker", Consumer: "path-checker", ProofChoice: "COHERENCE"},
		{ID: "judge-receipt", Activity: "JudgeExplanationReceipt", Producer: "independent-judge", Consumer: "ci-minimal-causal-explanation", ProofChoice: "REGRESSION"},
		{ID: "preserve-claims", Activity: "PreserveClaimTransitions", Producer: "claim-ledger", Consumer: "ci-minimal-causal-explanation", ProofChoice: "REGRESSION"},
	}
	expectedProgram := explanation.MetaProgram{Schema: explanation.SourceSchema, Producer: "gooo://meta/minimal-causal-explanation/evaluator", Consumer: "gooo://meta/minimal-causal-explanation/ci-judge", IndicatorDenominator: explanation.IndicatorTotal, MetaOperations: expected}
	if !equalJSON(program, expectedProgram) {
		return fmt.Errorf("meta program contract changed")
	}
	return nil
}

func validateCases(cases []explanation.ExplanationCase) error {
	if len(cases) != explanation.CaseTotal {
		return fmt.Errorf("case denominator changed")
	}
	expected := []struct {
		id, kind, verdict   string
		pathID              string
		evidence            []string
		edges               []string
		decision            string
		sufficient, minimal bool
		available           int
	}{
		{"minimal", "MINIMAL_SUFFICIENT", explanation.CaseAccepted, "path.minimal-approval", []string{requestEvidence, policyEvidence, resultEvidence}, []string{"edge.request.policy", "edge.policy.result"}, explanation.DecisionPass, true, true, 4},
		{"overlong", "SUFFICIENT_NOT_MINIMAL", explanation.CaseRejected, "path.overlong-approval", []string{requestEvidence, policyEvidence, resultEvidence, noiseEvidence}, []string{"edge.request.policy", "edge.policy.result"}, explanation.DecisionPass, true, false, 4},
		{"insufficient", "INSUFFICIENT", explanation.CaseRejected, "path.insufficient-approval", []string{requestEvidence, resultEvidence}, nil, explanation.DecisionFailClosed, false, false, 3},
	}
	for index, want := range expected {
		if cases[index].ID != want.id || cases[index].Kind != want.kind || cases[index].Verdict != want.verdict || cases[index].AvailableEvidenceTotal != want.available || len(cases[index].Paths) != 1 {
			return fmt.Errorf("case %q identity changed", want.id)
		}
		path := cases[index].Paths[0]
		if path.ID != want.pathID || !equalStrings(path.EvidenceIDs, want.evidence) || !equalStrings(path.EdgeIDs, want.edges) || path.Decision != want.decision || path.Sufficient != want.sufficient || path.Minimal != want.minimal {
			return fmt.Errorf("case %q path classification changed", want.id)
		}
		computedSufficient := decisionForEvidence(path.EvidenceIDs) == explanation.DecisionPass
		if path.Sufficient != computedSufficient {
			return fmt.Errorf("case %q sufficiency is not derived from its evidence", want.id)
		}
		if err := validateCounterfactuals(path, want.decision); err != nil {
			return fmt.Errorf("case %q: %w", want.id, err)
		}
		computedMinimal := computedSufficient && allChanged(path.Counterfactuals)
		if path.Minimal != computedMinimal {
			return fmt.Errorf("case %q minimality is not derived from removal", want.id)
		}
	}
	return nil
}

func allChanged(counterfactuals []explanation.Counterfactual) bool {
	if len(counterfactuals) == 0 {
		return false
	}
	for _, counterfactual := range counterfactuals {
		if !counterfactual.Changed {
			return false
		}
	}
	return true
}

func validateCounterfactuals(path explanation.ExplanationPath, before string) error {
	if before != explanation.DecisionPass {
		if len(path.Counterfactuals) != 0 {
			return fmt.Errorf("insufficient path must not claim counterfactual proof")
		}
		return nil
	}
	if len(path.Counterfactuals) != len(path.EvidenceIDs) {
		return fmt.Errorf("counterfactual denominator changed")
	}
	for index, item := range path.Counterfactuals {
		if item.RemovedEvidenceID != path.EvidenceIDs[index] || item.BeforeDecision != before {
			return fmt.Errorf("counterfactual %d does not bind path evidence", index+1)
		}
		remaining := append([]string(nil), path.EvidenceIDs[:index]...)
		remaining = append(remaining, path.EvidenceIDs[index+1:]...)
		after := decisionForEvidence(remaining)
		if item.AfterDecision != after || item.Changed != (before != after) || item.Coordinate.Stage != "COUNTERFACTUAL" || item.Coordinate.Step != "remove-evidence" {
			return fmt.Errorf("counterfactual %d is false", index+1)
		}
	}
	return nil
}

func validateSummary(summary explanation.Summary) error {
	if summary.CasesTotal != 3 || summary.CasesAccepted != 3 || summary.PathsObserved != 3 || summary.MinimalSufficientPaths != 1 || summary.SufficientNonminimalPaths != 1 || summary.InsufficientPaths != 1 || summary.CounterfactualTotal != 7 || summary.ChangedCounterfactualTotal != 6 || summary.CandidateEvidenceTotal != 11 {
		return fmt.Errorf("summary counts changed")
	}
	if summary.RepositoryWrites != 0 || summary.PromotionAuthorized || !summary.PathSetAuthoritative || summary.ExplanationTextRole != explanation.ExplanationTextRole {
		return fmt.Errorf("summary authority changed")
	}
	return nil
}

func validatePreservation(receipt explanation.Receipt) error {
	if receipt.Preservation.ClaimTotal != explanation.ClaimTotal || receipt.Preservation.PreservedTotal != explanation.ClaimTotal || receipt.Preservation.TransitionTotal != explanation.TransitionTotal || receipt.Preservation.Policy != "APPEND_ONLY_OPEN_TO_DISCHARGED" || len(receipt.ClaimTransitions) != explanation.TransitionTotal {
		return fmt.Errorf("preservation denominator changed")
	}
	ids := []string{"claim.source-bound", "claim.graph-closed", "claim.path-sufficient", "claim.path-minimal", "claim.counterfactual-difference", "claim.read-only-preserved"}
	previous := ""
	for index, transition := range receipt.ClaimTransitions {
		claim := ids[index/2]
		before, after := "UNRECORDED", "OPEN"
		if index%2 == 1 {
			before, after = "OPEN", "DISCHARGED"
		}
		if transition.Sequence != index+1 || transition.ClaimID != claim || transition.Before != before || transition.After != after || transition.PreviousTransitionDigest != previous || transition.TransitionDigest != transitionDigest(transition) {
			return fmt.Errorf("claim transition %d is not append-only", index+1)
		}
		previous = transition.TransitionDigest
	}
	if receipt.Preservation.TransitionHead != previous {
		return fmt.Errorf("claim transition head changed")
	}
	return nil
}

func validateIndicators(indicators []explanation.Indicator, program explanation.MetaProgram) error {
	if len(indicators) != explanation.IndicatorTotal {
		return fmt.Errorf("indicator denominator changed")
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied || indicator.ID == "" || indicator.MetaOperation == "" || indicator.Producer == "" || indicator.Consumer == "" || indicator.ProofChoice == "" || indicator.EvidenceDigest == "" {
			return fmt.Errorf("indicator %q is not independently bound", indicator.ID)
		}
		found := false
		for _, operation := range program.MetaOperations {
			if operation.ID == indicator.MetaOperation && operation.Producer == indicator.Producer && operation.Consumer == indicator.Consumer && operation.ProofChoice == indicator.ProofChoice {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("indicator %q has an unknown meta-operation", indicator.ID)
		}
	}
	return nil
}

func decisionForEvidence(evidence []string) string {
	for _, wanted := range []string{requestEvidence, policyEvidence, resultEvidence} {
		if !contains(evidence, wanted) {
			return explanation.DecisionFailClosed
		}
	}
	return explanation.DecisionPass
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func transitionDigest(transition explanation.ClaimTransition) string {
	transition.TransitionDigest = ""
	return digest(transition)
}

func independentReceiptDigest(receipt explanation.Receipt) string {
	receipt.ReceiptDigest = ""
	return digest(receipt)
}

func graphDigest(graph explanation.CausalGraph) string {
	graph.Digest = ""
	return digest(graph)
}

func digest(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func equalJSON(left, right any) bool {
	leftData, _ := json.Marshal(left)
	rightData, _ := json.Marshal(right)
	return string(leftData) == string(rightData)
}

func equalStrings(left, right []string) bool {
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
