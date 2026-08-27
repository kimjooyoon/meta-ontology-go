package claimdependencyjudge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/claimdependency"
)

const JudgmentSchema = "gooo.meta.claim-dependency-judgment/v1"

type Judgment struct {
	Schema                      string                  `json:"schema"`
	ReceiptDigest               string                  `json:"receipt_digest"`
	Case                        string                  `json:"case"`
	Decision                    string                  `json:"decision"`
	Resolution                  string                  `json:"resolution"`
	Reason                      string                  `json:"reason"`
	Accepted                    bool                    `json:"accepted"`
	IndependentReplay           string                  `json:"independent_replay"`
	Metrics                     claimdependency.Metrics `json:"metrics"`
	ReadOnly                    bool                    `json:"read_only"`
	SemanticPromotionAuthorized bool                    `json:"semantic_promotion_authorized"`
	Digest                      string                  `json:"digest"`
}

var expectedClaims = [...]claimdependency.Claim{
	{Ordinal: 1, Axis: "source-observed", ClaimID: "gooo.claim.dependency.source-observed.v1", Statement: "the Gooo source is the observed subject", Producer: claimdependency.ProducerID, Consumer: claimdependency.ConsumerID, MetaOperation: "observe-gooo-source", ProofChoice: "FOUNDATION", Coordinate: claimdependency.Coordinate{Stage: "READ", Step: "gooo-source", Reason: "SOURCE_READ"}},
	{Ordinal: 2, Axis: "producer-bound", ClaimID: "gooo.claim.dependency.producer-bound.v1", Statement: "the receipt identifies its deterministic producer", Producer: claimdependency.ProducerID, Consumer: claimdependency.ConsumerID, MetaOperation: "bind-producer", ProofChoice: "FOUNDATION", Coordinate: claimdependency.Coordinate{Stage: "BIND", Step: "producer", Reason: "PRODUCER_IDENTIFIED"}},
	{Ordinal: 3, Axis: "proof-choice-bound", ClaimID: "gooo.claim.dependency.proof-choice-bound.v1", Statement: "the state claim names a proof choice", Producer: claimdependency.ProducerID, Consumer: claimdependency.ConsumerID, MetaOperation: "choose-proof-route", ProofChoice: "COHERENCE", Coordinate: claimdependency.Coordinate{Stage: "BIND", Step: "proof-choice", Reason: "PROOF_CHOICE_DECLARED"}},
	{Ordinal: 4, Axis: "consumer-bound", ClaimID: "gooo.claim.dependency.consumer-bound.v1", Statement: "the receipt names an independent decision consumer", Producer: claimdependency.ProducerID, Consumer: claimdependency.ConsumerID, MetaOperation: "bind-consumer", ProofChoice: "COHERENCE", Coordinate: claimdependency.Coordinate{Stage: "BIND", Step: "consumer", Reason: "CONSUMER_IDENTIFIED"}},
	{Ordinal: 5, Axis: "read-only-bound", ClaimID: "gooo.claim.dependency.read-only-bound.v1", Statement: "the experiment cannot mutate the repository", Producer: claimdependency.ProducerID, Consumer: claimdependency.ConsumerID, MetaOperation: "deny-repository-mutation", ProofChoice: "REGRESSION", Coordinate: claimdependency.Coordinate{Stage: "GUARD", Step: "authority", Reason: "READ_ONLY"}},
	{Ordinal: 6, Axis: "decision-replay-bound", ClaimID: "gooo.claim.dependency.decision-replay-bound.v1", Statement: "an independent judge can replay the state decision", Producer: claimdependency.ProducerID, Consumer: claimdependency.ConsumerID, MetaOperation: claimdependency.MetaOperationID, ProofChoice: "REGRESSION", Coordinate: claimdependency.Coordinate{Stage: "JUDGE", Step: "replay-decision", Reason: "INDEPENDENT_REPLAY"}},
}

var expectedEdges = [...]claimdependency.Edge{
	{EdgeID: "E01", FromClaimID: expectedClaims[0].ClaimID, ToClaimID: expectedClaims[1].ClaimID, Kind: "SOURCE_INFORMS_PRODUCER"},
	{EdgeID: "E02", FromClaimID: expectedClaims[1].ClaimID, ToClaimID: expectedClaims[2].ClaimID, Kind: "PRODUCER_SELECTS_PROOF"},
	{EdgeID: "E03", FromClaimID: expectedClaims[1].ClaimID, ToClaimID: expectedClaims[3].ClaimID, Kind: "PRODUCER_BINDS_CONSUMER"},
	{EdgeID: "E04", FromClaimID: expectedClaims[1].ClaimID, ToClaimID: expectedClaims[4].ClaimID, Kind: "PRODUCER_DENIES_MUTATION"},
	{EdgeID: "E05", FromClaimID: expectedClaims[2].ClaimID, ToClaimID: expectedClaims[5].ClaimID, Kind: "PROOF_SUPPORTS_DECISION"},
	{EdgeID: "E06", FromClaimID: expectedClaims[3].ClaimID, ToClaimID: expectedClaims[5].ClaimID, Kind: "CONSUMER_ACCEPTS_RECEIPT"},
	{EdgeID: "E07", FromClaimID: expectedClaims[4].ClaimID, ToClaimID: expectedClaims[5].ClaimID, Kind: "AUTHORITY_GUARDRAIL"},
	{EdgeID: "E08", FromClaimID: expectedClaims[1].ClaimID, ToClaimID: expectedClaims[5].ClaimID, Kind: "PRODUCER_TRACEABLE_DECISION"},
}

func Judge(receipt claimdependency.Receipt) (Judgment, error) {
	if err := validateReceipt(receipt); err != nil {
		return Judgment{}, err
	}
	states, outcomes := statesAndOutcomes(receipt)
	if err := validateCaseShape(receipt.Subject.Case, states, outcomes); err != nil {
		return Judgment{}, err
	}
	if err := validateResolutions(receipt, states, outcomes); err != nil {
		return Judgment{}, err
	}
	metrics := deriveMetrics(receipt.Resolutions)
	if !reflect.DeepEqual(metrics, receipt.Metrics) {
		return Judgment{}, fmt.Errorf("receipt metrics are not independently reproducible")
	}
	decision, resolution, reason := expectedDecision(states)
	if receipt.Decision.Value != decision || receipt.Decision.Resolution != resolution || receipt.Decision.Reason != reason || receipt.Decision.SemanticPromotionAuthorized {
		return Judgment{}, fmt.Errorf("receipt decision is not independently reproducible")
	}
	judgment := Judgment{
		Schema: JudgmentSchema, ReceiptDigest: receipt.Digest, Case: receipt.Subject.Case,
		Decision: decision, Resolution: resolution, Reason: reason, Accepted: true,
		IndependentReplay: "GRAPH_AND_TRANSITION_REDERIVED", Metrics: metrics,
		ReadOnly:                    receipt.Subject.ReadOnly && receipt.Subject.RepositoryWrites == 0,
		SemanticPromotionAuthorized: false,
	}
	judgment.Digest = digestJudgment(judgment)
	return judgment, nil
}

func validateReceipt(receipt claimdependency.Receipt) error {
	if receipt.Schema != claimdependency.ReceiptSchema || receipt.Scope != claimdependency.Scope {
		return fmt.Errorf("receipt identity is invalid")
	}
	if receipt.Subject.Producer != claimdependency.ProducerID || receipt.Subject.Consumer != claimdependency.ConsumerID || receipt.Subject.MetaOperation != claimdependency.MetaOperationID || receipt.Subject.ProofChoice != claimdependency.ProofChoice || !receipt.Subject.ReadOnly || receipt.Subject.RepositoryWrites != 0 {
		return fmt.Errorf("receipt provenance boundary is invalid")
	}
	if receipt.Subject.SourcePath == "" || len(receipt.Subject.SourceDigest) != 64 {
		return fmt.Errorf("receipt source provenance is invalid")
	}
	if receipt.Graph.Schema != claimdependency.GraphSchema || receipt.Graph.NodeTotal != claimdependency.ClaimTotal || receipt.Graph.EdgeTotal != claimdependency.EdgeTotal || len(receipt.Graph.Nodes) != claimdependency.ClaimTotal || len(receipt.Graph.Edges) != claimdependency.EdgeTotal {
		return fmt.Errorf("graph denominator is invalid")
	}
	for index, claim := range expectedClaims {
		if !reflect.DeepEqual(receipt.Graph.Nodes[index], claim) {
			return fmt.Errorf("claim node %d is not the fixed contract", index+1)
		}
	}
	for index, edge := range expectedEdges {
		if !reflect.DeepEqual(receipt.Graph.Edges[index], edge) {
			return fmt.Errorf("graph edge %d is not the fixed contract", index+1)
		}
	}
	graph := receipt.Graph
	graph.Digest = ""
	if digestValue(graph) != receipt.Graph.Digest {
		return fmt.Errorf("graph digest is invalid")
	}
	if len(receipt.Transitions) != claimdependency.TransitionTotal || len(receipt.Resolutions) != claimdependency.ClaimTotal {
		return fmt.Errorf("transition or resolution denominator is invalid")
	}
	previous := ""
	for index, transition := range receipt.Transitions {
		if transition.Sequence != index+1 || transition.PreviousTransitionDigest != previous || transition.TransitionDigest != digestTransition(transition) {
			return fmt.Errorf("transition %d chain is invalid", index+1)
		}
		if index < claimdependency.ClaimTotal {
			claim := expectedClaims[index]
			if transition.ClaimID != claim.ClaimID || transition.Event != "CLAIM_REGISTERED" || transition.Before != "UNRECORDED" || transition.After != "OPEN" || !reflect.DeepEqual(transition.Coordinate, claimdependency.Coordinate{Stage: "DECLARE", Step: claim.Axis, Reason: "CLAIM_REGISTERED"}) || transition.EvidenceDigest != "" {
				return fmt.Errorf("registration transition %d is invalid", index+1)
			}
		} else {
			claimIndex := index - claimdependency.ClaimTotal
			if transition.ClaimID != expectedClaims[claimIndex].ClaimID || transition.Before != "OPEN" || !validOutcomeCoordinate(receipt.Subject.Case, claimIndex, transition) {
				return fmt.Errorf("outcome transition %d is invalid", index+1)
			}
		}
		previous = transition.TransitionDigest
	}
	if receipt.Digest == "" || receipt.Digest != digestReceipt(receipt) {
		return fmt.Errorf("receipt digest is invalid")
	}
	return nil
}

func validOutcomeCoordinate(caseName string, index int, transition claimdependency.Transition) bool {
	want := claimdependency.Coordinate{Stage: "PROPAGATE", Step: expectedClaims[index].Axis}
	switch {
	case caseName == claimdependency.CaseDirectUnknown && index == 0:
		want = claimdependency.Coordinate{Stage: "RESOLVE", Step: "observe-gooo-source", Reason: "SOURCE_EVIDENCE_UNKNOWN"}
	case caseName == claimdependency.CaseDirectUnknown:
		want.Reason = "UPSTREAM_CLAIM_OPEN"
	case caseName == claimdependency.CaseRefuted && index == 0:
		want = claimdependency.Coordinate{Stage: "VERIFY", Step: "compare-gooo-source", Reason: "SOURCE_CONTRADICTS_EXPECTATION"}
	case caseName == claimdependency.CaseRefuted:
		want.Reason = "UPSTREAM_CLAIM_REFUTED"
	case caseName == claimdependency.CaseRecovered && index == 0:
		want = claimdependency.Coordinate{Stage: "VERIFY", Step: "compare-gooo-source", Reason: "SOURCE_MATCHES_EXPECTATION"}
	case caseName == claimdependency.CaseRecovered:
		want.Reason = "UPSTREAM_CLAIM_DISCHARGED"
	default:
		return false
	}
	return reflect.DeepEqual(transition.Coordinate, want)
}

func statesAndOutcomes(receipt claimdependency.Receipt) ([]string, []claimdependency.Transition) {
	states := make([]string, claimdependency.ClaimTotal)
	outcomes := make([]claimdependency.Transition, claimdependency.ClaimTotal)
	for index := range claimdependency.ClaimTotal {
		registration := receipt.Transitions[index]
		outcome := receipt.Transitions[claimdependency.ClaimTotal+index]
		if registration.ClaimID != expectedClaims[index].ClaimID || registration.Event != "CLAIM_REGISTERED" || registration.Before != "UNRECORDED" || registration.After != "OPEN" || outcome.ClaimID != expectedClaims[index].ClaimID || outcome.Before != "OPEN" {
			states[index] = "INVALID"
			continue
		}
		states[index] = outcome.After
		outcomes[index] = outcome
	}
	return states, outcomes
}

func validateCaseShape(caseName string, states []string, outcomes []claimdependency.Transition) error {
	if caseName != claimdependency.CaseDirectUnknown && caseName != claimdependency.CaseRefuted && caseName != claimdependency.CaseRecovered {
		return fmt.Errorf("unknown case %q", caseName)
	}
	for index, state := range states {
		switch caseName {
		case claimdependency.CaseDirectUnknown:
			if state != "OPEN" || (index == 0 && outcomes[index].Event != "EVIDENCE_UNAVAILABLE") || (index > 0 && outcomes[index].Event != "DEPENDENCY_BLOCKED") {
				return fmt.Errorf("direct unknown case state %d is invalid", index+1)
			}
		case claimdependency.CaseRefuted:
			if state != "REFUTED" || (index == 0 && outcomes[index].Event != "EVIDENCE_REFUTED") || (index > 0 && outcomes[index].Event != "DEPENDENCY_REFUTED") {
				return fmt.Errorf("refuted case state %d is invalid", index+1)
			}
		case claimdependency.CaseRecovered:
			if state != "DISCHARGED" || (index == 0 && outcomes[index].Event != "EVIDENCE_ACCEPTED") || (index > 0 && outcomes[index].Event != "DEPENDENCY_RECOVERED") {
				return fmt.Errorf("recovered case state %d is invalid", index+1)
			}
		}
	}
	if caseName == claimdependency.CaseRecovered && outcomes[0].After != "DISCHARGED" {
		return fmt.Errorf("recovery did not discharge the root claim")
	}
	return nil
}

func validateResolutions(receipt claimdependency.Receipt, states []string, outcomes []claimdependency.Transition) error {
	root := expectedClaims[0].ClaimID
	for index, resolution := range receipt.Resolutions {
		if resolution.ClaimID != expectedClaims[index].ClaimID || resolution.Axis != expectedClaims[index].Axis || resolution.State != states[index] || resolution.ObservedEvent != outcomes[index].Event || !reflect.DeepEqual(resolution.Coordinate, outcomes[index].Coordinate) || resolution.CauseTransitionDigest != outcomes[0].TransitionDigest || resolution.CauseCoordinate == nil || !reflect.DeepEqual(*resolution.CauseCoordinate, outcomes[0].Coordinate) {
			return fmt.Errorf("resolution %d is not bound to the root transition", index+1)
		}
		wantPath := shortestPath(root, resolution.ClaimID)
		if !reflect.DeepEqual(resolution.CausePath, wantPath) || !reflect.DeepEqual(resolution.CauseEdgeIDs, edgesForPath(wantPath)) {
			return fmt.Errorf("resolution %d does not preserve the minimum cause path", index+1)
		}
		switch resolution.Kind {
		case "DIRECT_UNKNOWN":
			if index != 0 || states[index] != "OPEN" || resolution.FailureResponsibility != "LOCAL_PRODUCER" || !reflect.DeepEqual(resolution.MissingEvidenceIDs, []string{"evidence:" + root}) || len(resolution.BlockedByEdgeIDs) != 0 {
				return fmt.Errorf("direct unknown responsibility is invalid")
			}
		case "DEPENDENCY_BLOCKED":
			if index == 0 || states[index] != "OPEN" || resolution.FailureResponsibility != "UPSTREAM_CLAIM" || !reflect.DeepEqual(resolution.MissingEvidenceIDs, []string{"evidence:" + root}) || !frontierMatches(index, "OPEN", resolution) {
				return fmt.Errorf("dependency block responsibility is invalid")
			}
		case "DIRECT_REFUTED":
			if index != 0 || states[index] != "REFUTED" || resolution.FailureResponsibility != "LOCAL_PRODUCER" || len(resolution.MissingEvidenceIDs) != 0 || len(resolution.BlockedByEdgeIDs) != 0 {
				return fmt.Errorf("direct refutation responsibility is invalid")
			}
		case "DEPENDENCY_REFUTED":
			if index == 0 || states[index] != "REFUTED" || resolution.FailureResponsibility != "UPSTREAM_CLAIM" || len(resolution.MissingEvidenceIDs) != 0 || !frontierMatches(index, "REFUTED", resolution) {
				return fmt.Errorf("dependency refutation responsibility is invalid")
			}
		case "DIRECT_DISCHARGED":
			if index != 0 || states[index] != "DISCHARGED" || resolution.FailureResponsibility != "LOCAL_PRODUCER" || len(resolution.MissingEvidenceIDs) != 0 || len(resolution.BlockedByEdgeIDs) != 0 {
				return fmt.Errorf("direct discharge is invalid")
			}
		case "DEPENDENCY_RECOVERED":
			if index == 0 || states[index] != "DISCHARGED" || resolution.FailureResponsibility != "UPSTREAM_CLAIM" || len(resolution.MissingEvidenceIDs) != 0 || len(resolution.BlockedByEdgeIDs) != 0 {
				return fmt.Errorf("dependency recovery is invalid")
			}
		default:
			return fmt.Errorf("unknown resolution kind %q", resolution.Kind)
		}
	}
	return nil
}

func frontierMatches(index int, state string, resolution claimdependency.Resolution) bool {
	wantClaims, wantEdges := []string{}, []string{}
	for _, edge := range expectedEdges {
		if edge.ToClaimID != expectedClaims[index].ClaimID {
			continue
		}
		from := claimIndex(edge.FromClaimID)
		if from >= 0 && indexStateForFrontier(from, index, state) {
			wantClaims = append(wantClaims, edge.FromClaimID)
			wantEdges = append(wantEdges, edge.EdgeID)
		}
	}
	return reflect.DeepEqual(resolution.BlockedByClaimIDs, wantClaims) && reflect.DeepEqual(resolution.BlockedByEdgeIDs, wantEdges)
}

func indexStateForFrontier(from, target int, state string) bool {
	if from == 0 {
		return true
	}
	return target > from && (state == "OPEN" || state == "REFUTED")
}

func shortestPath(root, target string) []string {
	if root == target {
		return []string{root}
	}
	best := []string(nil)
	var visit func(string, []string, map[string]bool)
	visit = func(current string, path []string, seen map[string]bool) {
		for _, edge := range expectedEdges {
			if edge.FromClaimID != current || seen[edge.ToClaimID] {
				continue
			}
			next := append(append([]string(nil), path...), edge.ToClaimID)
			if edge.ToClaimID == target {
				if best == nil || len(next) < len(best) || (len(next) == len(best) && joinPath(next) < joinPath(best)) {
					best = next
				}
				continue
			}
			seen[edge.ToClaimID] = true
			visit(edge.ToClaimID, next, seen)
			delete(seen, edge.ToClaimID)
		}
	}
	visit(root, []string{root}, map[string]bool{root: true})
	return best
}

func edgesForPath(path []string) []string {
	result := make([]string, 0, len(path)-1)
	for index := 1; index < len(path); index++ {
		for _, edge := range expectedEdges {
			if edge.FromClaimID == path[index-1] && edge.ToClaimID == path[index] {
				result = append(result, edge.EdgeID)
				break
			}
		}
	}
	return result
}

func claimIndex(claimID string) int {
	for index, claim := range expectedClaims {
		if claim.ClaimID == claimID {
			return index
		}
	}
	return -1
}

func joinPath(path []string) string {
	result := ""
	for _, claimID := range path {
		result += claimID + "\x00"
	}
	return result
}

func deriveMetrics(resolutions []claimdependency.Resolution) claimdependency.Metrics {
	metrics := claimdependency.Metrics{FixedClaimTotal: claimdependency.ClaimTotal, FixedEdgeTotal: claimdependency.EdgeTotal, ClassifiedClaimTotal: len(resolutions), TransitionTotal: claimdependency.TransitionTotal}
	recoveryEdges := map[string]bool{}
	for _, resolution := range resolutions {
		switch resolution.State {
		case "OPEN":
			metrics.OpenClaimTotal++
		case "DISCHARGED":
			metrics.DischargedClaimTotal++
		case "REFUTED":
			metrics.RefutedClaimTotal++
		}
		switch resolution.Kind {
		case "DIRECT_UNKNOWN":
			metrics.UnknownClaimTotal++
			metrics.DirectUnknownClaimTotal++
		case "DEPENDENCY_BLOCKED":
			metrics.UnknownClaimTotal++
			metrics.DependencyBlockedClaimTotal++
			metrics.ObservedBlockingEdgeTotal += len(resolution.BlockedByEdgeIDs)
		case "DIRECT_REFUTED":
			metrics.DirectRefutedClaimTotal++
		case "DEPENDENCY_REFUTED":
			metrics.DependencyRefutedClaimTotal++
			metrics.ObservedRefutingEdgeTotal += len(resolution.BlockedByEdgeIDs)
		case "DEPENDENCY_RECOVERED":
			metrics.DependencyRecoveredTotal++
			for _, edgeID := range resolution.CauseEdgeIDs {
				recoveryEdges[edgeID] = true
			}
		}
		if depth := len(resolution.CausePath) - 1; depth > metrics.MaximumCausePathDepth {
			metrics.MaximumCausePathDepth = depth
		}
	}
	metrics.ObservedRecoveryEdgeTotal = len(recoveryEdges)
	if metrics.FixedClaimTotal > 0 {
		metrics.ClassificationBasisPoints = metrics.ClassifiedClaimTotal * 10000 / metrics.FixedClaimTotal
	}
	return metrics
}

func expectedDecision(states []string) (string, string, string) {
	for _, state := range states {
		if state == "REFUTED" {
			return "FAIL_CLOSED", "CAUSAL_REFUTATION", "DIRECT_REFUTATION_PROPAGATED"
		}
	}
	for _, state := range states {
		if state == "OPEN" {
			return "FAIL_CLOSED", "CAUSAL_DEPENDENCY_BLOCK", "DIRECT_UNKNOWN_BLOCKED_DESCENDANTS"
		}
	}
	return "PASS", "CAUSAL_RECOVERY_DISCHARGED", "UPSTREAM_RECOVERY_PROPAGATED"
}

func digestValue(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func digestTransition(transition claimdependency.Transition) string {
	return digestValue(transition)
}

func digestReceipt(receipt claimdependency.Receipt) string {
	receipt.Digest = ""
	return digestValue(receipt)
}

func digestJudgment(judgment Judgment) string {
	judgment.Digest = ""
	return digestValue(judgment)
}
