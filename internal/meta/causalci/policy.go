package causalci

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	changedFileSemanticID = "gooo://causal-ci-selection/surface/changed-file"
	claimSemanticID       = "gooo://causal-ci-selection/claim/causal-selection"
	surfaceSemanticID     = "gooo://causal-ci-selection/surface/semantic-policy"
	checkSemanticPrefix   = "gooo://causal-ci-selection/check/"
	stateSemanticPrefix   = "gooo://causal-ci-selection/claim-state/"
)

// ReconstructPolicy is the producer's source authority boundary. It parses
// the real Gooo source, lowers it to semantic IR, and then reconstructs the
// causal policy from typed facts and semantic value programs.
func ReconstructPolicy(sourcePath string, source []byte) (PolicyGraph, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if diagnostics.HasErrors() || file == nil {
		return PolicyGraph{}, fmt.Errorf("%s: %s", ReasonMalformedPolicy, diagnostics.Error())
	}
	canonical, err := syntax.Format(file)
	if err != nil {
		return PolicyGraph{}, fmt.Errorf("%s: %w", ReasonMalformedPolicy, err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return PolicyGraph{}, fmt.Errorf("%s: lower: %w", ReasonMalformedPolicy, err)
	}

	policy := PolicyGraph{
		Source: SourceEvidence{
			Path:           sourcePath,
			RawDigest:      digestBytes(source),
			ParsedDigest:   digestBytes([]byte(canonical)),
			SemanticDigest: digestBytes([]byte(ir.SemanticCanonical())),
		},
		ClaimStateRules: map[string]string{},
	}
	if err := collectPolicyEntities(&policy, ir); err != nil {
		return PolicyGraph{}, err
	}
	if err := collectPolicyEdges(&policy, ir); err != nil {
		return PolicyGraph{}, err
	}
	validatePolicyShape(&policy)
	return policy, nil
}

func collectPolicyEntities(policy *PolicyGraph, ir semantic.IR) error {
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Entity {
			continue
		}
		id := string(node.ID)
		switch {
		case id == changedFileSemanticID:
			policy.ChangedFileID = id
		case id == claimSemanticID:
			policy.ClaimID = id
		case id == surfaceSemanticID:
			policy.SurfaceID = id
		case strings.HasPrefix(id, checkSemanticPrefix):
			check, err := checkFromSemanticID(id)
			if err != nil {
				return err
			}
			policy.Checks = append(policy.Checks, check)
		}
	}
	sort.Slice(policy.Checks, func(i, j int) bool { return policy.Checks[i].Ordinal < policy.Checks[j].Ordinal })
	if policy.ChangedFileID == "" || policy.ClaimID == "" || policy.SurfaceID == "" {
		return fmt.Errorf("%s: required causal entities are absent", ReasonMalformedPolicy)
	}
	if len(policy.Checks) != FixedCheckDenominator {
		return fmt.Errorf("%s: check denominator is %d, want %d", ReasonMalformedPolicy, len(policy.Checks), FixedCheckDenominator)
	}
	seen := map[string]struct{}{}
	for index, check := range policy.Checks {
		if check.Ordinal != index+1 || check.ID != fixedCheckIDs[index] {
			return fmt.Errorf("%s: check ordinal %d is %q", ReasonMalformedPolicy, index+1, check.ID)
		}
		if _, exists := seen[check.ID]; exists {
			return fmt.Errorf("%s: duplicate check %q", ReasonMalformedPolicy, check.ID)
		}
		seen[check.ID] = struct{}{}
	}
	return nil
}

func checkFromSemanticID(id string) (Check, error) {
	value := strings.TrimPrefix(id, checkSemanticPrefix)
	if len(value) < 3 || value[2] != '-' {
		return Check{}, fmt.Errorf("%s: malformed check identity %q", ReasonMalformedPolicy, id)
	}
	ordinal, err := strconv.Atoi(value[:2])
	if err != nil || ordinal < 1 {
		return Check{}, fmt.Errorf("%s: malformed check ordinal %q", ReasonMalformedPolicy, id)
	}
	return Check{ID: value[3:], Ordinal: ordinal, SemanticID: id}, nil
}

func collectPolicyEdges(policy *PolicyGraph, ir semantic.IR) error {
	used := map[string][]string{}
	outputs := map[string][]string{}
	for _, fact := range ir.Graph.Facts() {
		subject, object := string(fact.Subject), string(fact.Object)
		switch fact.Predicate {
		case semantic.Used:
			used[subject] = append(used[subject], object)
		case semantic.WasGeneratedBy:
			outputs[object] = append(outputs[object], subject)
		}
	}
	for _, node := range ir.Graph.Nodes() {
		program := node.ValueProgram
		if program == "" {
			continue
		}
		kind, recognized := policyProgramKind(program)
		if !recognized {
			continue
		}
		inputs := append([]string(nil), used[string(node.ID)]...)
		results := append([]string(nil), outputs[string(node.ID)]...)
		sort.Strings(inputs)
		sort.Strings(results)
		if len(inputs) != 1 || len(results) != 1 {
			return fmt.Errorf("%s: activity %q does not have one input and output", ReasonMalformedPolicy, node.ID)
		}
		if kind == "claim-state" {
			state := strings.TrimPrefix(strings.TrimPrefix(program, "causal-ci.claim-transition/"), "causal-ci.prior-claim-state/")
			if strings.HasPrefix(state, "open/") {
				state = ClaimOpen
			} else if strings.HasPrefix(state, "discharged/") {
				state = ClaimDischarged
			} else if strings.HasPrefix(state, "refuted/") {
				state = ClaimRefuted
			} else if strings.HasPrefix(state, "open-lower-resolution/") {
				state = ClaimOpen
			}
			policy.PriorStates = append(policy.PriorStates, PriorStateRule{State: state, ActivityID: string(node.ID), ValueProgram: program})
			policy.ClaimStateRules[program] = state
			continue
		}
		edge := PolicyEdge{Kind: kind, From: inputs[0], To: results[0], ActivityID: string(node.ID), ValueProgram: program}
		canonical := edge.Kind + "\x00" + edge.From + "\x00" + edge.To + "\x00" + edge.ActivityID + "\x00" + edge.ValueProgram
		edge.ID = "policy-edge:" + strings.TrimPrefix(digestBytes([]byte(canonical)), "sha256:")
		policy.Edges = append(policy.Edges, edge)
	}
	sort.Slice(policy.Edges, func(i, j int) bool { return policy.Edges[i].ID < policy.Edges[j].ID })
	sort.Slice(policy.PriorStates, func(i, j int) bool { return policy.PriorStates[i].ValueProgram < policy.PriorStates[j].ValueProgram })
	return nil
}

func policyProgramKind(program string) (string, bool) {
	switch program {
	case programChangedFileToClaim:
		return "changed-file-to-claim", true
	case programClaimToSurface:
		return "claim-to-surface", true
	case programSurfaceToCheck:
		return "surface-to-check", true
	case programPriorClaimState, programDischarge, programLowerResolution, programRefute:
		return "claim-state", true
	default:
		return "", false
	}
}

func validatePolicyShape(policy *PolicyGraph) {
	for _, edge := range policy.Edges {
		if edge.Kind == "changed-file-to-claim" && (edge.From != policy.ChangedFileID || edge.To != policy.ClaimID) {
			policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: stageConformance, Step: stepValidatePolicy, Reason: "CHANGED_FILE_CLAIM_ENDPOINT_MISMATCH", Edges: []string{edge.ID}})
		}
		if edge.Kind == "claim-to-surface" && (edge.From != policy.ClaimID || edge.To != policy.SurfaceID) {
			policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: stageConformance, Step: stepValidatePolicy, Reason: "CLAIM_SURFACE_ENDPOINT_MISMATCH", Edges: []string{edge.ID}})
		}
		if edge.Kind == "surface-to-check" && !isKnownCheck(policy, edge.To) {
			policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: stageConformance, Step: stepValidatePolicy, Reason: "SURFACE_CHECK_ENDPOINT_UNREGISTERED", Edges: []string{edge.ID}})
		}
	}
	var surfaceEdges []string
	for _, edge := range policy.Edges {
		if edge.Kind == "surface-to-check" && edge.From == policy.SurfaceID {
			surfaceEdges = append(surfaceEdges, edge.ID)
		}
	}
	if len(surfaceEdges) > 1 {
		policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: stageConformance, Step: stepValidatePolicy, Reason: ReasonContradictoryPolicy, Edges: surfaceEdges})
	}
	if countEdges(policy, "changed-file-to-claim") != 1 || countEdges(policy, "claim-to-surface") != 1 || len(surfaceEdges) != 1 {
		policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: stageConformance, Step: stepValidatePolicy, Reason: "REQUIRED_CAUSAL_POLICY_VALUE_MISSING", Edges: nil})
	}
	requiredStatePrograms := []string{programPriorClaimState, programDischarge, programLowerResolution, programRefute}
	for _, program := range requiredStatePrograms {
		if _, exists := policy.ClaimStateRules[program]; !exists {
			policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: stageConformance, Step: stepValidatePolicy, Reason: "CLAIM_STATE_POLICY_INCOMPLETE", Edges: nil})
		}
	}
	if len(policy.PriorStates) < 4 {
		policy.Contradictions = append(policy.Contradictions, PolicyContradiction{Stage: stageConformance, Step: stepValidatePolicy, Reason: "CLAIM_STATE_POLICY_INCOMPLETE", Edges: nil})
	}
}

func countEdges(policy *PolicyGraph, kind string) int {
	count := 0
	for _, edge := range policy.Edges {
		if edge.Kind == kind {
			count++
		}
	}
	return count
}

func isKnownCheck(policy *PolicyGraph, value string) bool {
	for _, check := range policy.Checks {
		if check.SemanticID == value {
			return true
		}
	}
	return false
}

func policyEdge(policy PolicyGraph, kind, from string) []PolicyEdge {
	result := []PolicyEdge{}
	for _, edge := range policy.Edges {
		if edge.Kind == kind && edge.From == from {
			result = append(result, edge)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func checkIDBySemanticID(policy PolicyGraph, id string) string {
	for _, check := range policy.Checks {
		if check.SemanticID == id {
			return check.ID
		}
	}
	return ""
}
