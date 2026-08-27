package claimdependency

import (
	"fmt"
	"sort"
	"strings"
)

// ObservationForSource binds an observation predicate to the root activity
// recovered from canonical semantic IR. It deliberately has no case-name
// input: the source value program and the observation predicate must agree.
func ObservationForSource(source []byte, sourcePath string, predicate ObservationPredicate, evidence string) (Observation, error) {
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return Observation{}, err
	}
	observation := Observation{
		Schema: ObservationSchema, Predicate: predicate, SubjectClaimID: parsed.Graph.Nodes[0].ClaimID,
		ReadOnly: true, RepositoryWrites: 0, MutationAuthority: false,
	}
	if evidence != "" {
		observation.EvidenceDigest = digestBytes([]byte(evidence))
	}
	observation.Digest, err = observationDigest(observation)
	if err != nil {
		return Observation{}, err
	}
	return observation, nil
}

// Evaluate classifies a source and, when prior is non-nil, appends recovery
// transitions to that exact prior ledger. It is the producer-side evaluator;
// the judge intentionally reconstructs all of this from raw inputs.
func Evaluate(source []byte, sourcePath string, observation Observation, prior *Receipt) (Receipt, error) {
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return Receipt{}, err
	}
	if err := validateObservation(parsed, observation); err != nil {
		return Receipt{}, err
	}
	if prior != nil {
		if observation.Predicate != ObservationEvidence {
			return Receipt{}, fmt.Errorf("a prior ledger is only valid for evidence recovery")
		}
		if err := validatePrior(parsed, *prior); err != nil {
			return Receipt{}, err
		}
	}

	sourceDigest := digestBytes(source)
	semanticDigest := parsed.Graph.CanonicalIRDigest
	provenance := fmt.Sprintf("source:%s|ir:%s|producer:%s|consumer:%s", sourceDigest, semanticDigest, ProducerID, ConsumerID)
	states, outcomes := classify(parsed.Graph, observation.Predicate, sourceDigest, semanticDigest, provenance, observation.EvidenceDigest)
	transitions := make([]Transition, 0, InitialTransitionTotal)
	if prior == nil {
		transitions, err = initialTransitions(parsed.Graph, outcomes, sourceDigest, semanticDigest, provenance)
	} else {
		transitions = append(transitions, prior.Transitions...)
		transitions, err = appendRecoveryTransitions(transitions, parsed.Graph, prior.Resolutions, sourceDigest, semanticDigest, provenance, observation.EvidenceDigest)
	}
	if err != nil {
		return Receipt{}, err
	}
	currentOutcomes := outcomes
	if prior != nil {
		currentOutcomes = transitions[len(transitions)-ClaimTotal:]
	}
	resolutions := buildResolutions(parsed.Graph, states, currentOutcomes, sourceDigest, semanticDigest, prior != nil)
	metrics := deriveMetrics(parsed.Graph, states, resolutions, prior != nil)
	decision := decisionFor(observation.Predicate, prior != nil, observation)
	receipt := Receipt{
		Schema: ReceiptSchema, Scope: Scope,
		Subject: Subject{
			SourcePath: sourcePath, SourceDigest: sourceDigest, SemanticDigest: semanticDigest,
			Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID,
			ProofChoice: ProofChoice, ReadOnly: observation.ReadOnly,
			RepositoryWrites: observation.RepositoryWrites,
		},
		Observation: observation, Graph: parsed.Graph,
		ObservationDigest: observation.Digest, Transitions: transitions,
		TransitionHeadDigest: transitions[len(transitions)-1].TransitionDigest,
		Resolutions:          resolutions, Metrics: metrics, Decision: decision,
	}
	if prior != nil {
		receipt.PriorReceiptDigest, err = receiptDigest(*prior)
		if err != nil {
			return Receipt{}, err
		}
		receipt.PreviousTransitionDigest = prior.TransitionHeadDigest
		receipt.PriorClaimStates = resolutionStates(prior.Resolutions)
		receipt.Metrics.AppendOnlyTransitionTotal = ClaimTotal
	}
	receipt.Digest, err = receiptDigest(receipt)
	if err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}

func validateObservation(parsed sourceGraph, observation Observation) error {
	if observation.Schema != ObservationSchema || observation.SubjectClaimID != parsed.Graph.Nodes[0].ClaimID {
		return fmt.Errorf("observation identity does not bind to the canonical root claim")
	}
	if !observation.ReadOnly || observation.RepositoryWrites != 0 || observation.MutationAuthority {
		return fmt.Errorf("observation violates read-only boundary")
	}
	if observation.Digest == "" {
		return fmt.Errorf("observation digest is missing")
	}
	computed, err := observationDigest(observation)
	if err != nil || computed != observation.Digest {
		return fmt.Errorf("observation digest is invalid")
	}
	root := parsed.RootProgram
	switch observation.Predicate {
	case ObservationUnknown:
		if root != "claim.observe:recoverable" || observation.EvidenceDigest != "" {
			return fmt.Errorf("UNKNOWN requires a recoverable source predicate and no evidence")
		}
	case ObservationEvidence:
		if root != "claim.observe:recoverable" || observation.EvidenceDigest == "" {
			return fmt.Errorf("EVIDENCE_ACCEPTED requires recoverable source semantics and evidence")
		}
	case ObservationContradiction:
		if root != "claim.observe:contradiction" || observation.EvidenceDigest == "" {
			return fmt.Errorf("EXPLICIT_CONTRADICTION requires a contradiction source predicate and evidence")
		}
	default:
		return fmt.Errorf("unknown observation predicate %q", observation.Predicate)
	}
	return nil
}

func validatePrior(parsed sourceGraph, prior Receipt) error {
	if prior.Schema != ReceiptSchema || prior.Scope != Scope || prior.Observation.Predicate != ObservationUnknown {
		return fmt.Errorf("prior ledger is not an UNKNOWN receipt")
	}
	if prior.Graph.Digest != parsed.Graph.Digest || prior.Graph.NodeTotal != ClaimTotal || prior.Graph.EdgeTotal != EdgeTotal {
		return fmt.Errorf("recovery graph does not preserve the prior semantic graph")
	}
	if prior.Subject.SourceDigest == "" || prior.TransitionHeadDigest == "" || len(prior.Resolutions) != ClaimTotal {
		return fmt.Errorf("prior ledger is incomplete")
	}
	computedReceipt, err := receiptDigest(prior)
	if err != nil || computedReceipt != prior.Digest {
		return fmt.Errorf("prior receipt digest is invalid")
	}
	if err := validateTransitionChain(prior.Transitions, prior.TransitionHeadDigest); err != nil {
		return fmt.Errorf("prior transition chain: %w", err)
	}
	if !sameStrings(prior.PriorClaimStates, nil) && len(prior.PriorClaimStates) != 0 {
		return fmt.Errorf("initial UNKNOWN ledger cannot have prior claim states")
	}
	for index, resolution := range prior.Resolutions {
		if resolution.State != "OPEN" || resolution.ClaimID != prior.Graph.Nodes[index].ClaimID {
			return fmt.Errorf("prior claim state %d is not preserved OPEN", index+1)
		}
	}
	return nil
}

func classify(graph Graph, predicate ObservationPredicate, sourceDigest, semanticDigest, provenance, evidence string) ([]string, []Transition) {
	states := make([]string, len(graph.Nodes))
	outcomes := make([]Transition, len(graph.Nodes))
	for index := range graph.Nodes {
		state := "OPEN"
		event := "DEPENDENCY_BLOCKED"
		reason := "UPSTREAM_UNKNOWN_OR_NON_REFUTING"
		if index == 0 {
			switch predicate {
			case ObservationContradiction:
				state, event, reason = "REFUTED", "EXPLICIT_CONTRADICTION", "OBSERVATION_PREDICATE_EXPLICITLY_CONTRADICTS"
			case ObservationEvidence:
				state, event, reason = "DISCHARGED", "EVIDENCE_ACCEPTED", "OBSERVATION_EVIDENCE_PREDICATE_SATISFIED"
			default:
				state, event, reason = "OPEN", "OBSERVATION_UNKNOWN", "OBSERVATION_PREDICATE_UNKNOWN"
			}
		} else if predicate == ObservationEvidence {
			state, event, reason = "DISCHARGED", "DEPENDENCY_DISCHARGED", "EVIDENCE_PREDICATE_SATISFIED"
		} else if predicate == ObservationContradiction && hasExplicitRefutation(index, graph, states) {
			state, event, reason = "REFUTED", "DEPENDENCY_REFUTED", "EXPLICIT_REFUTING_EDGE"
		}
		states[index] = state
		outcomes[index] = Transition{
			ClaimID: graph.Nodes[index].ClaimID, Event: event, Before: "OPEN", After: state,
			Coordinate:     Coordinate{Stage: outcomeStage(index), Step: graph.Nodes[index].ActivityName, Reason: reason},
			EvidenceDigest: evidence, Provenance: provenance,
		}
	}
	return states, outcomes
}

func hasExplicitRefutation(index int, graph Graph, states []string) bool {
	for _, edge := range graph.Edges {
		if edge.ToClaimID != graph.Nodes[index].ClaimID || states[indexOfClaim(edge.FromClaimID, graph)] != "REFUTED" {
			continue
		}
		if edge.Kind == Contradicts || edge.Kind == FailureEntailment {
			return true
		}
	}
	return false
}

func outcomeStage(index int) string {
	if index == 0 {
		return "OBSERVE"
	}
	return "PROPAGATE"
}

func initialTransitions(graph Graph, outcomes []Transition, sourceDigest, semanticDigest, provenance string) ([]Transition, error) {
	transitions := make([]Transition, 0, InitialTransitionTotal)
	previous := ""
	for _, claim := range graph.Nodes {
		transition := Transition{
			Sequence: len(transitions) + 1, ClaimID: claim.ClaimID, Event: "CLAIM_REGISTERED",
			Before: "UNRECORDED", After: "OPEN",
			Coordinate: Coordinate{Stage: "DECLARE", Step: claim.ActivityName, Reason: "CLAIM_REGISTERED"},
			Provenance: provenance, PreviousTransitionDigest: previous,
		}
		var err error
		transition.TransitionDigest, err = transitionDigest(transition)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, transition)
		previous = transition.TransitionDigest
	}
	for index, outcome := range outcomes {
		outcome.Sequence = len(transitions) + 1
		outcome.PreviousTransitionDigest = previous
		outcome.Provenance = provenance
		if outcome.EvidenceDigest == "" && outcome.After == "DISCHARGED" {
			outcome.EvidenceDigest = digestBytes([]byte(sourceDigest + semanticDigest))
		}
		var err error
		outcome.TransitionDigest, err = transitionDigest(outcome)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, outcome)
		previous = transitions[len(transitions)-1].TransitionDigest
		_ = index
	}
	return transitions, nil
}

func appendRecoveryTransitions(transitions []Transition, graph Graph, prior []Resolution, sourceDigest, semanticDigest, provenance, evidence string) ([]Transition, error) {
	previous := transitions[len(transitions)-1].TransitionDigest
	for index, claim := range graph.Nodes {
		before := prior[index].State
		event := "DEPENDENCY_DISCHARGED"
		reason := "EVIDENCE_PREDICATE_SATISFIED"
		if index == 0 {
			event, reason = "EVIDENCE_ACCEPTED", "RECOVERY_EVIDENCE_PREDICATE_SATISFIED"
		}
		transition := Transition{
			Sequence: len(transitions) + 1, ClaimID: claim.ClaimID, Event: event,
			Before: before, After: "DISCHARGED",
			Coordinate:     Coordinate{Stage: "RECOVER", Step: claim.ActivityName, Reason: reason},
			EvidenceDigest: evidence, Provenance: provenance, PreviousTransitionDigest: previous,
		}
		var err error
		transition.TransitionDigest, err = transitionDigest(transition)
		if err != nil {
			return nil, err
		}
		transitions = append(transitions, transition)
		previous = transition.TransitionDigest
	}
	return transitions, nil
}

func buildResolutions(graph Graph, states []string, outcomes []Transition, sourceDigest, semanticDigest string, recovered bool) []Resolution {
	root := graph.Nodes[0].ClaimID
	rootOutcome := outcomes[0]
	rootCoordinate := rootOutcome.Coordinate
	provenance := fmt.Sprintf("source:%s|ir:%s|producer:%s|consumer:%s", sourceDigest, semanticDigest, ProducerID, ConsumerID)
	resolutions := make([]Resolution, 0, len(graph.Nodes))
	for index, claim := range graph.Nodes {
		path, edgeIDs, edgeKinds := shortestPath(index, graph)
		outcome := outcomes[index]
		coordinate := outcome.Coordinate
		resolution := Resolution{
			ClaimID: claim.ClaimID, Axis: claim.Axis, State: states[index], Kind: resolutionKind(index, states[index], recovered),
			ObservedEvent: outcome.Event, Coordinate: coordinate, EvidenceDigest: outcome.EvidenceDigest,
			Provenance: provenance, FailureResponsibility: "LOCAL_PRODUCER", FailureOwnerClaimID: root,
			CausePath: claimIDs(path, graph), CauseEdgeIDs: edgeIDs, CauseEdgeKinds: edgeKinds,
			CauseTransitionDigest: rootOutcome.TransitionDigest, CauseCoordinate: &rootCoordinate,
		}
		if index != 0 {
			resolution.FailureResponsibility = "UPSTREAM_CLAIM"
		}
		if states[index] == "OPEN" {
			resolution.MissingEvidenceIDs = []string{"evidence:" + root}
			resolution.BlockedByClaimIDs, resolution.BlockedByEdgeIDs = blockedFrontier(index, graph, states)
		}
		resolutions = append(resolutions, resolution)
	}
	return resolutions
}

func resolutionKind(index int, state string, recovered bool) string {
	if index == 0 {
		switch state {
		case "REFUTED":
			return "DIRECT_REFUTED"
		case "DISCHARGED":
			return "DIRECT_DISCHARGED"
		default:
			return "DIRECT_UNKNOWN"
		}
	}
	switch state {
	case "REFUTED":
		return "DEPENDENCY_REFUTED"
	case "DISCHARGED":
		if recovered {
			return "DEPENDENCY_RECOVERED"
		}
		return "DEPENDENCY_DISCHARGED"
	default:
		return "DEPENDENCY_BLOCKED"
	}
}

func shortestPath(index int, graph Graph) ([]int, []string, []EdgeKind) {
	if index == 0 {
		return []int{0}, nil, nil
	}
	best := []int(nil)
	for _, edge := range graph.Edges {
		if edge.ToClaimID != graph.Nodes[index].ClaimID {
			continue
		}
		from := indexOfClaim(edge.FromClaimID, graph)
		candidate, _, _ := shortestPath(from, graph)
		candidate = append(candidate, index)
		if best == nil || len(candidate) < len(best) || (len(candidate) == len(best) && pathKey(candidate, graph) < pathKey(best, graph)) {
			best = candidate
		}
	}
	if best == nil {
		return []int{index}, nil, nil
	}
	ids := make([]string, 0, len(best)-1)
	kinds := make([]EdgeKind, 0, len(best)-1)
	for position := 1; position < len(best); position++ {
		for _, edge := range graph.Edges {
			if edge.FromClaimID == graph.Nodes[best[position-1]].ClaimID && edge.ToClaimID == graph.Nodes[best[position]].ClaimID {
				ids = append(ids, edge.EdgeID)
				kinds = append(kinds, edge.Kind)
				break
			}
		}
	}
	return best, ids, kinds
}

func pathKey(path []int, graph Graph) string {
	parts := make([]string, len(path))
	for index, value := range path {
		parts[index] = graph.Nodes[value].ClaimID
	}
	return strings.Join(parts, "\x00")
}
func claimIDs(path []int, graph Graph) []string {
	ids := make([]string, 0, len(path))
	for _, index := range path {
		ids = append(ids, graph.Nodes[index].ClaimID)
	}
	return ids
}
func indexOfClaim(claimID string, graph Graph) int {
	for index, claim := range graph.Nodes {
		if claim.ClaimID == claimID {
			return index
		}
	}
	return -1
}

func blockedFrontier(index int, graph Graph, states []string) ([]string, []string) {
	claims, edges := []string{}, []string{}
	for _, edge := range graph.Edges {
		if edge.ToClaimID != graph.Nodes[index].ClaimID {
			continue
		}
		from := indexOfClaim(edge.FromClaimID, graph)
		if from < 0 {
			continue
		}
		if states[from] == "OPEN" || (states[from] == "REFUTED" && (edge.Kind == Supports || edge.Kind == Requires)) {
			claims = append(claims, edge.FromClaimID)
			edges = append(edges, edge.EdgeID)
		}
	}
	return claims, edges
}

func deriveMetrics(graph Graph, states []string, resolutions []Resolution, recovered bool) Metrics {
	metrics := Metrics{
		FixedClaimTotal: ClaimTotal, FixedEdgeTotal: EdgeTotal, ClassifiedClaimTotal: len(resolutions),
		TransitionTotal: InitialTransitionTotal, ClassificationBasisPoints: 10000,
	}
	for _, state := range states {
		switch state {
		case "OPEN":
			metrics.OpenClaimTotal++
		case "DISCHARGED":
			metrics.DischargedClaimTotal++
		case "REFUTED":
			metrics.RefutedClaimTotal++
		}
	}
	for _, resolution := range resolutions {
		switch resolution.Kind {
		case "DIRECT_UNKNOWN":
			metrics.UnknownClaimTotal++
			metrics.DirectUnknownClaimTotal++
		case "DEPENDENCY_BLOCKED":
			metrics.DependencyBlockedClaimTotal++
		case "DIRECT_REFUTED":
			metrics.DirectRefutedClaimTotal++
		case "DEPENDENCY_REFUTED":
			metrics.DependencyRefutedClaimTotal++
		case "DIRECT_DISCHARGED":
			metrics.DirectDischargedClaimTotal++
		case "DEPENDENCY_DISCHARGED", "DEPENDENCY_RECOVERED":
			metrics.DependencyDischargedTotal++
		}
		if len(resolution.CausePath) > metrics.MaximumCausePathDepth {
			metrics.MaximumCausePathDepth = len(resolution.CausePath)
		}
	}
	if recovered {
		metrics.TransitionTotal = InitialTransitionTotal + ClaimTotal
	}
	for _, edgeKind := range []EdgeKind{Supports, Requires, Contradicts, FailureEntailment} {
		metric := EdgeMetric{Kind: edgeKind}
		for _, edge := range graph.Edges {
			if edge.Kind != edgeKind {
				continue
			}
			metric.Total++
			from, to := indexOfClaim(edge.FromClaimID, graph), indexOfClaim(edge.ToClaimID, graph)
			if recovered && states[to] == "DISCHARGED" {
				metric.Recovery++
			}
			if states[to] == "OPEN" && (states[from] == "OPEN" || (states[from] == "REFUTED" && (edge.Kind == Supports || edge.Kind == Requires))) {
				metric.Blocking++
			}
			if states[to] == "REFUTED" && states[from] == "REFUTED" && (edge.Kind == Contradicts || edge.Kind == FailureEntailment) {
				metric.Refuting++
			}
		}
		metrics.ObservedBlockingEdgeTotal += metric.Blocking
		metrics.ObservedRefutingEdgeTotal += metric.Refuting
		metrics.ObservedRecoveryEdgeTotal += metric.Recovery
		metrics.EdgeMetrics = append(metrics.EdgeMetrics, metric)
	}
	return metrics
}

func decisionFor(predicate ObservationPredicate, recovered bool, observation Observation) Decision {
	if !observation.ReadOnly || observation.RepositoryWrites != 0 || observation.MutationAuthority {
		return Decision{Value: "FAIL_CLOSED", Resolution: "EFFECTS_REJECTED", Reason: "READ_ONLY_OBSERVATION_REQUIRED"}
	}
	switch {
	case recovered:
		return Decision{Value: "PASS", Resolution: "CAUSAL_RECOVERY_DISCHARGED", Reason: "APPEND_ONLY_EVIDENCE_RECOVERY"}
	case predicate == ObservationContradiction:
		return Decision{Value: "FAIL_CLOSED", Resolution: "CAUSAL_REFUTATION", Reason: "EXPLICIT_CONTRADICTION_EDGE_ALGEBRA"}
	default:
		return Decision{Value: "FAIL_CLOSED", Resolution: "UNRESOLVED_CLAIM", Reason: "UNKNOWN_REMAINS_OPEN"}
	}
}

func resolutionStates(resolutions []Resolution) []string {
	states := make([]string, len(resolutions))
	for index, resolution := range resolutions {
		states[index] = resolution.State
	}
	return states
}
func validateTransitionChain(transitions []Transition, head string) error {
	if len(transitions) == 0 || transitions[len(transitions)-1].TransitionDigest != head {
		return fmt.Errorf("transition head does not match chain")
	}
	previous := ""
	for index, transition := range transitions {
		if transition.Sequence != index+1 || transition.PreviousTransitionDigest != previous {
			return fmt.Errorf("transition %d predecessor mismatch", index+1)
		}
		computed, err := transitionDigest(transition)
		if err != nil || computed != transition.TransitionDigest {
			return fmt.Errorf("transition %d digest mismatch", index+1)
		}
		previous = transition.TransitionDigest
	}
	return nil
}
func sameStrings(left, right []string) bool {
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

// Keep the edge algebra closed and its serialized order fixed for artifacts.
func EdgeKinds() []EdgeKind {
	result := []EdgeKind{Supports, Requires, Contradicts, FailureEntailment}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
