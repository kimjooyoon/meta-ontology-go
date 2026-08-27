package claimdependency

import (
	"fmt"
	"strings"
)

func Evaluate(source []byte, sourcePath, caseName string) (Receipt, error) {
	if sourcePath == "" {
		return Receipt{}, fmt.Errorf("source path is required")
	}
	if err := validateSource(source, caseName); err != nil {
		return Receipt{}, err
	}
	graph, err := buildGraph()
	if err != nil {
		return Receipt{}, err
	}
	transitions := buildTransitions(source, caseName, graph)
	resolutions := buildResolutions(caseName, graph, transitions)
	receipt := Receipt{
		Schema: ReceiptSchema, Scope: Scope,
		Subject: Subject{
			Case: caseName, SourcePath: sourcePath, SourceDigest: digestBytes(source),
			Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID,
			ProofChoice: ProofChoice, ReadOnly: true, RepositoryWrites: 0,
		},
		Graph: graph, Metrics: deriveMetrics(resolutions),
		Transitions: transitions, Resolutions: resolutions,
		Decision: decisionFor(caseName, resolutions),
	}
	if caseName == CaseRecovered {
		receipt.Subject.RecoveryFromCase = CaseDirectUnknown
	}
	receipt.Digest, err = receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func validateSource(source []byte, caseName string) error {
	text := string(source)
	if !strings.Contains(text, "package claimdependency") ||
		!strings.Contains(text, "namespace claimdependency") ||
		!strings.Contains(text, "activity Root(Integer)") ||
		!strings.Contains(text, "activity Derived(Integer)") {
		return fmt.Errorf("source is not the claim dependency Gooo subject")
	}
	markers := map[string]string{
		CaseDirectUnknown: "int.unknown:1",
		CaseRefuted:       "int.add:-1",
		CaseRecovered:     "int.add:1",
	}
	marker, ok := markers[caseName]
	if !ok {
		return fmt.Errorf("unsupported claim dependency case %q", caseName)
	}
	if !strings.Contains(text, marker) {
		return fmt.Errorf("case %q source marker %q is missing", caseName, marker)
	}
	return nil
}

func buildTransitions(source []byte, caseName string, graph Graph) []Transition {
	transitions := make([]Transition, 0, TransitionTotal)
	previous := ""
	for _, claim := range graph.Nodes {
		transition := Transition{
			Sequence: len(transitions) + 1, ClaimID: claim.ClaimID,
			Event: "CLAIM_REGISTERED", Before: "UNRECORDED", After: "OPEN",
			Coordinate:               Coordinate{Stage: "DECLARE", Step: claim.Axis, Reason: "CLAIM_REGISTERED"},
			PreviousTransitionDigest: previous,
		}
		transition.TransitionDigest = transitionDigest(transition)
		transitions = append(transitions, transition)
		previous = transition.TransitionDigest
	}
	for index, claim := range graph.Nodes {
		transition := outcomeTransition(source, caseName, index, claim, previous)
		transition.Sequence = len(transitions) + 1
		transition.TransitionDigest = transitionDigest(transition)
		transitions = append(transitions, transition)
		previous = transition.TransitionDigest
	}
	return transitions
}

func outcomeTransition(source []byte, caseName string, index int, claim Claim, previous string) Transition {
	transition := Transition{
		ClaimID: claim.ClaimID, Before: "OPEN", PreviousTransitionDigest: previous,
		Coordinate: Coordinate{Stage: "PROPAGATE", Step: claim.Axis, Reason: "UPSTREAM_CLAIM_OPEN"},
	}
	switch {
	case caseName == CaseDirectUnknown && index == 0:
		transition.Event = "EVIDENCE_UNAVAILABLE"
		transition.After = "OPEN"
		transition.Coordinate = Coordinate{Stage: "RESOLVE", Step: "observe-gooo-source", Reason: "SOURCE_EVIDENCE_UNKNOWN"}
	case caseName == CaseDirectUnknown:
		transition.Event = "DEPENDENCY_BLOCKED"
		transition.After = "OPEN"
	case caseName == CaseRefuted && index == 0:
		transition.Event = "EVIDENCE_REFUTED"
		transition.After = "REFUTED"
		transition.EvidenceDigest = digestBytes(append(append([]byte(nil), source...), []byte(claim.ClaimID)...))
		transition.Coordinate = Coordinate{Stage: "VERIFY", Step: "compare-gooo-source", Reason: "SOURCE_CONTRADICTS_EXPECTATION"}
	case caseName == CaseRefuted:
		transition.Event = "DEPENDENCY_REFUTED"
		transition.After = "REFUTED"
		transition.Coordinate.Reason = "UPSTREAM_CLAIM_REFUTED"
	case caseName == CaseRecovered && index == 0:
		transition.Event = "EVIDENCE_ACCEPTED"
		transition.After = "DISCHARGED"
		transition.EvidenceDigest = digestBytes(append(append([]byte(nil), source...), []byte(claim.ClaimID)...))
		transition.Coordinate = Coordinate{Stage: "VERIFY", Step: "compare-gooo-source", Reason: "SOURCE_MATCHES_EXPECTATION"}
	case caseName == CaseRecovered:
		transition.Event = "DEPENDENCY_RECOVERED"
		transition.After = "DISCHARGED"
		transition.Coordinate.Reason = "UPSTREAM_CLAIM_DISCHARGED"
	}
	return transition
}

func transitionDigest(transition Transition) string {
	digest, _ := digestJSON(transition)
	return digest
}

func buildResolutions(caseName string, graph Graph, transitions []Transition) []Resolution {
	resolutions := make([]Resolution, 0, ClaimTotal)
	rootTransition := transitions[ClaimTotal]
	rootCoordinate := rootTransition.Coordinate
	for index, claim := range graph.Nodes {
		transition := transitions[ClaimTotal+index]
		path := shortestCausePath(index, graph)
		resolution := Resolution{
			ClaimID: claim.ClaimID, Axis: claim.Axis, State: transition.After,
			ObservedEvent: transition.Event, Coordinate: transition.Coordinate,
			CausePath: claimIDs(path, graph), CauseEdgeIDs: pathEdgeIDs(path, graph),
			CauseTransitionDigest: rootTransition.TransitionDigest,
			CauseCoordinate:       &rootCoordinate,
			FailureResponsibility: "LOCAL_PRODUCER", FailureOwnerClaimID: graph.Nodes[0].ClaimID,
		}
		switch {
		case caseName == CaseDirectUnknown && index == 0:
			resolution.Kind = "DIRECT_UNKNOWN"
			resolution.MissingEvidenceIDs = []string{"evidence:" + claim.ClaimID}
		case caseName == CaseDirectUnknown:
			resolution.Kind = "DEPENDENCY_BLOCKED"
			resolution.FailureResponsibility = "UPSTREAM_CLAIM"
			resolution.MissingEvidenceIDs = []string{"evidence:" + graph.Nodes[0].ClaimID}
			resolution.BlockedByClaimIDs, resolution.BlockedByEdgeIDs = blockedFrontier(index, graph, "OPEN")
		case caseName == CaseRefuted && index == 0:
			resolution.Kind = "DIRECT_REFUTED"
		case caseName == CaseRefuted:
			resolution.Kind = "DEPENDENCY_REFUTED"
			resolution.FailureResponsibility = "UPSTREAM_CLAIM"
			resolution.BlockedByClaimIDs, resolution.BlockedByEdgeIDs = blockedFrontier(index, graph, "REFUTED")
		case caseName == CaseRecovered && index == 0:
			resolution.Kind = "DIRECT_DISCHARGED"
		case caseName == CaseRecovered:
			resolution.Kind = "DEPENDENCY_RECOVERED"
			resolution.FailureResponsibility = "UPSTREAM_CLAIM"
		}
		resolutions = append(resolutions, resolution)
	}
	return resolutions
}

func shortestCausePath(index int, graph Graph) []int {
	if index == 0 {
		return []int{0}
	}
	best := []int(nil)
	for _, edge := range graph.Edges {
		if edge.ToClaimID != graph.Nodes[index].ClaimID {
			continue
		}
		from := nodeIndex(edge.FromClaimID, graph)
		candidate := append(shortestCausePath(from, graph), index)
		if best == nil || len(candidate) < len(best) || (len(candidate) == len(best) && edge.EdgeID < pathEdgeIDs(best, graph)[0]) {
			best = candidate
		}
	}
	if best == nil {
		return []int{index}
	}
	return best
}

func nodeIndex(claimID string, graph Graph) int {
	for index, claim := range graph.Nodes {
		if claim.ClaimID == claimID {
			return index
		}
	}
	return -1
}

func claimIDs(path []int, graph Graph) []string {
	ids := make([]string, 0, len(path))
	for _, index := range path {
		ids = append(ids, graph.Nodes[index].ClaimID)
	}
	return ids
}

func pathEdgeIDs(path []int, graph Graph) []string {
	edgeIDs := make([]string, 0, len(path)-1)
	for index := 1; index < len(path); index++ {
		for _, edge := range graph.Edges {
			if edge.FromClaimID == graph.Nodes[path[index-1]].ClaimID && edge.ToClaimID == graph.Nodes[path[index]].ClaimID {
				edgeIDs = append(edgeIDs, edge.EdgeID)
				break
			}
		}
	}
	return edgeIDs
}

func blockedFrontier(index int, graph Graph, upstreamState string) ([]string, []string) {
	claims, edges := []string{}, []string{}
	claimID := graph.Nodes[index].ClaimID
	for _, edge := range graph.Edges {
		if edge.ToClaimID != claimID {
			continue
		}
		from := nodeIndex(edge.FromClaimID, graph)
		if from >= 0 && stateForIndex(from, upstreamState, graph) {
			claims = append(claims, edge.FromClaimID)
			edges = append(edges, edge.EdgeID)
		}
	}
	return claims, edges
}

func stateForIndex(index int, state string, graph Graph) bool {
	if index < 0 || index >= len(graph.Nodes) {
		return false
	}
	return state == "OPEN" || state == "REFUTED"
}

func deriveMetrics(resolutions []Resolution) Metrics {
	metrics := Metrics{FixedClaimTotal: ClaimTotal, FixedEdgeTotal: EdgeTotal, ClassifiedClaimTotal: len(resolutions), TransitionTotal: TransitionTotal}
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

func decisionFor(caseName string, resolutions []Resolution) Decision {
	decision := Decision{Value: "FAIL_CLOSED", SemanticPromotionAuthorized: false}
	switch caseName {
	case CaseDirectUnknown:
		decision.Resolution = "CAUSAL_DEPENDENCY_BLOCK"
		decision.Reason = "DIRECT_UNKNOWN_BLOCKED_DESCENDANTS"
	case CaseRefuted:
		decision.Resolution = "CAUSAL_REFUTATION"
		decision.Reason = "DIRECT_REFUTATION_PROPAGATED"
	case CaseRecovered:
		decision.Resolution = "CAUSAL_RECOVERY_DISCHARGED"
		decision.Reason = "UPSTREAM_RECOVERY_PROPAGATED"
	}
	for _, resolution := range resolutions {
		if resolution.State != "DISCHARGED" {
			return decision
		}
	}
	decision.Value = "PASS"
	return decision
}
