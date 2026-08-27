package claimdependencyjudge

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const JudgmentSchema = "gooo.meta.claim-dependency-judgment/v2"

const (
	claimTotal = 6
	edgeTotal  = 8
	producerID = "gooo://meta/claim-dependency/producer/v2"
	consumerID = "gooo://meta/claim-dependency/independent-judge/v2"
	operation  = "classify-claim-state-causality"
	proof      = "COHERENCE"
	unknown    = "UNKNOWN"
	evidence   = "EVIDENCE_ACCEPTED"
	contradict = "EXPLICIT_CONTRADICTION"
)

type edgeKind string

const (
	supports          edgeKind = "SUPPORTS"
	requires          edgeKind = "REQUIRES"
	contradicts       edgeKind = "CONTRADICTS"
	failureEntailment edgeKind = "FAILURE_ENTAILMENT"
)

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}
type claim struct {
	Ordinal       int        `json:"ordinal"`
	Axis          string     `json:"axis"`
	ClaimID       string     `json:"claim_id"`
	ActivityID    string     `json:"activity_id"`
	ActivityName  string     `json:"activity_name"`
	Statement     string     `json:"statement"`
	ValueProgram  string     `json:"value_program"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	ProofChoice   string     `json:"proof_choice"`
	Coordinate    coordinate `json:"coordinate"`
}
type edge struct {
	EdgeID        string   `json:"edge_id"`
	FromClaimID   string   `json:"from_claim_id"`
	ToClaimID     string   `json:"to_claim_id"`
	Kind          edgeKind `json:"kind"`
	SemanticBasis string   `json:"semantic_basis"`
}
type graph struct {
	Schema            string  `json:"schema"`
	Authority         string  `json:"authority"`
	Completeness      string  `json:"completeness"`
	CanonicalIRDigest string  `json:"canonical_ir_digest"`
	NodeTotal         int     `json:"node_total"`
	EdgeTotal         int     `json:"edge_total"`
	Nodes             []claim `json:"nodes"`
	Edges             []edge  `json:"edges"`
	Digest            string  `json:"digest"`
}
type observation struct {
	Schema            string `json:"schema"`
	Predicate         string `json:"predicate"`
	SubjectClaimID    string `json:"subject_claim_id"`
	EvidenceDigest    string `json:"evidence_digest,omitempty"`
	ReadOnly          bool   `json:"read_only"`
	RepositoryWrites  int    `json:"repository_writes"`
	MutationAuthority bool   `json:"mutation_authority"`
	Digest            string `json:"digest"`
}
type subject struct {
	SourcePath       string `json:"source_path"`
	SourceDigest     string `json:"source_digest"`
	SemanticDigest   string `json:"semantic_digest"`
	Producer         string `json:"producer"`
	Consumer         string `json:"consumer"`
	MetaOperation    string `json:"meta_operation"`
	ProofChoice      string `json:"proof_choice"`
	ReadOnly         bool   `json:"read_only"`
	RepositoryWrites int    `json:"repository_writes"`
}
type transition struct {
	Sequence                 int        `json:"sequence"`
	ClaimID                  string     `json:"claim_id"`
	Event                    string     `json:"event"`
	Before                   string     `json:"before"`
	After                    string     `json:"after"`
	Coordinate               coordinate `json:"coordinate"`
	EvidenceDigest           string     `json:"evidence_digest,omitempty"`
	Provenance               string     `json:"provenance"`
	PreviousTransitionDigest string     `json:"previous_transition_digest,omitempty"`
	TransitionDigest         string     `json:"transition_digest"`
}
type resolution struct {
	ClaimID               string      `json:"claim_id"`
	Axis                  string      `json:"axis"`
	State                 string      `json:"state"`
	Kind                  string      `json:"kind"`
	ObservedEvent         string      `json:"observed_event"`
	Coordinate            coordinate  `json:"coordinate"`
	EvidenceDigest        string      `json:"evidence_digest,omitempty"`
	Provenance            string      `json:"provenance"`
	FailureResponsibility string      `json:"failure_responsibility"`
	FailureOwnerClaimID   string      `json:"failure_owner_claim_id"`
	MissingEvidenceIDs    []string    `json:"missing_evidence_ids,omitempty"`
	BlockedByClaimIDs     []string    `json:"blocked_by_claim_ids,omitempty"`
	BlockedByEdgeIDs      []string    `json:"blocked_by_edge_ids,omitempty"`
	CausePath             []string    `json:"cause_path"`
	CauseEdgeIDs          []string    `json:"cause_edge_ids"`
	CauseEdgeKinds        []edgeKind  `json:"cause_edge_kinds"`
	CauseTransitionDigest string      `json:"cause_transition_digest"`
	CauseCoordinate       *coordinate `json:"cause_coordinate"`
}
type edgeMetric struct {
	Kind     edgeKind `json:"kind"`
	Total    int      `json:"total"`
	Blocking int      `json:"blocking"`
	Refuting int      `json:"refuting"`
	Recovery int      `json:"recovery"`
}
type metrics struct {
	FixedClaimTotal             int          `json:"fixed_claim_total"`
	FixedEdgeTotal              int          `json:"fixed_edge_total"`
	ClassifiedClaimTotal        int          `json:"classified_claim_total"`
	OpenClaimTotal              int          `json:"open_claim_total"`
	DischargedClaimTotal        int          `json:"discharged_claim_total"`
	RefutedClaimTotal           int          `json:"refuted_claim_total"`
	UnknownClaimTotal           int          `json:"unknown_claim_total"`
	DirectUnknownClaimTotal     int          `json:"direct_unknown_claim_total"`
	DependencyBlockedClaimTotal int          `json:"dependency_blocked_claim_total"`
	DirectRefutedClaimTotal     int          `json:"direct_refuted_claim_total"`
	DependencyRefutedClaimTotal int          `json:"dependency_refuted_claim_total"`
	DirectDischargedClaimTotal  int          `json:"direct_discharged_claim_total"`
	DependencyDischargedTotal   int          `json:"dependency_discharged_claim_total"`
	ObservedBlockingEdgeTotal   int          `json:"observed_blocking_edge_total"`
	ObservedRefutingEdgeTotal   int          `json:"observed_refuting_edge_total"`
	ObservedRecoveryEdgeTotal   int          `json:"observed_recovery_edge_total"`
	MaximumCausePathDepth       int          `json:"maximum_cause_path_depth"`
	TransitionTotal             int          `json:"transition_total"`
	AppendOnlyTransitionTotal   int          `json:"append_only_transition_total"`
	ClassificationBasisPoints   int          `json:"classification_basis_points"`
	EdgeMetrics                 []edgeMetric `json:"edge_metrics"`
}
type decision struct {
	Value                       string `json:"value"`
	Resolution                  string `json:"resolution"`
	Reason                      string `json:"reason"`
	SemanticPromotionAuthorized bool   `json:"semantic_promotion_authorized"`
}
type receipt struct {
	Schema                   string       `json:"schema"`
	Scope                    string       `json:"scope"`
	Subject                  subject      `json:"subject"`
	Observation              observation  `json:"observation"`
	Graph                    graph        `json:"graph"`
	PriorReceiptDigest       string       `json:"prior_receipt_digest,omitempty"`
	PreviousTransitionDigest string       `json:"previous_transition_digest,omitempty"`
	PriorClaimStates         []string     `json:"prior_claim_states,omitempty"`
	ObservationDigest        string       `json:"observation_digest"`
	Transitions              []transition `json:"transitions"`
	TransitionHeadDigest     string       `json:"transition_head_digest"`
	Resolutions              []resolution `json:"resolutions"`
	Metrics                  metrics      `json:"metrics"`
	Decision                 decision     `json:"decision"`
	Digest                   string       `json:"digest"`
}

type Judgment struct {
	Schema                           string  `json:"schema"`
	ReceiptDigest                    string  `json:"receipt_digest"`
	Predicate                        string  `json:"predicate"`
	Decision                         string  `json:"decision"`
	Resolution                       string  `json:"resolution"`
	Reason                           string  `json:"reason"`
	Accepted                         bool    `json:"accepted"`
	IndependentReplay                string  `json:"independent_replay"`
	Metrics                          metrics `json:"metrics"`
	ReadOnly                         bool    `json:"read_only"`
	SemanticPromotionAuthorized      bool    `json:"semantic_promotion_authorized"`
	SourceReconstruction             string  `json:"source_reconstruction"`
	ProducerPackageImportNumerator   int     `json:"producer_package_import_numerator"`
	ProducerPackageImportDenominator int     `json:"producer_package_import_denominator"`
	AppendOnlyRecoveryChainTotal     int     `json:"append_only_recovery_chain_total"`
	Digest                           string  `json:"digest"`
}

type reconstructed struct {
	Graph       graph
	RootProgram string
}

// Judge is intentionally a raw-input consumer. It does not import the
// producer package or consume producer expectations; all claims, edges,
// states, and transitions are rebuilt from .gooo, observation, and ledger.
func Judge(source []byte, sourcePath string, priorBytes, observationBytes, receiptBytes []byte) (Judgment, error) {
	current, err := reconstruct(source, sourcePath)
	if err != nil {
		return Judgment{}, err
	}
	var got receipt
	if err := json.Unmarshal(receiptBytes, &got); err != nil {
		return Judgment{}, fmt.Errorf("decode receipt: %w", err)
	}
	var observed observation
	if err := json.Unmarshal(observationBytes, &observed); err != nil {
		return Judgment{}, fmt.Errorf("decode observation: %w", err)
	}
	if err := validateObservation(current, observed); err != nil {
		return Judgment{}, err
	}
	if !reflect.DeepEqual(got.Observation, observed) || got.ObservationDigest != observed.Digest {
		return Judgment{}, fmt.Errorf("receipt is not bound to the raw observation")
	}
	if got.Graph.Digest != current.Graph.Digest || !reflect.DeepEqual(got.Graph, current.Graph) {
		return Judgment{}, fmt.Errorf("receipt graph is not reconstructed from raw source")
	}
	sourceDigest := digestBytes(source)
	if got.Subject.SourceDigest != sourceDigest || got.Subject.SourcePath != sourcePath || got.Subject.SemanticDigest != current.Graph.CanonicalIRDigest || got.Subject.Producer != producerID || got.Subject.Consumer != consumerID || got.Subject.MetaOperation != operation || got.Subject.ProofChoice != proof || !got.Subject.ReadOnly || got.Subject.RepositoryWrites != 0 {
		return Judgment{}, fmt.Errorf("receipt subject provenance is invalid")
	}

	var prior *receipt
	if len(priorBytes) > 0 {
		var value receipt
		if err := json.Unmarshal(priorBytes, &value); err != nil {
			return Judgment{}, fmt.Errorf("decode prior receipt: %w", err)
		}
		prior = &value
		if err := validatePrior(current, value); err != nil {
			return Judgment{}, err
		}
		priorDigest := receiptDigest(value)
		if got.PriorReceiptDigest != priorDigest || got.PreviousTransitionDigest != value.TransitionHeadDigest || !sameStrings(got.PriorClaimStates, statesOf(value.Resolutions)) {
			return Judgment{}, fmt.Errorf("recovery does not append the prior receipt head and claim states")
		}
		if len(got.Transitions) <= len(value.Transitions) || !reflect.DeepEqual(got.Transitions[:len(value.Transitions)], value.Transitions) {
			return Judgment{}, fmt.Errorf("recovery transition prefix is not append-only")
		}
	}

	states, outcomeTemplates := classify(current.Graph, observed.Predicate, observed.EvidenceDigest)
	provenance := fmt.Sprintf("source:%s|ir:%s|producer:%s|consumer:%s", sourceDigest, current.Graph.CanonicalIRDigest, producerID, consumerID)
	expectedTransitions, err := expectedTransitionsFor(current.Graph, outcomeTemplates, provenance, prior)
	if err != nil {
		return Judgment{}, err
	}
	if !reflect.DeepEqual(got.Transitions, expectedTransitions) {
		return Judgment{}, fmt.Errorf("receipt transition chain is not independently reproducible")
	}
	currentOutcomes := expectedTransitions[claimTotal:]
	if prior != nil {
		currentOutcomes = expectedTransitions[len(expectedTransitions)-claimTotal:]
	}
	expectedResolutions := buildResolutions(current.Graph, states, currentOutcomes, sourceDigest, current.Graph.CanonicalIRDigest, prior != nil)
	if !reflect.DeepEqual(got.Resolutions, expectedResolutions) {
		return Judgment{}, fmt.Errorf("receipt resolutions are not independently reproducible")
	}
	expectedMetrics := deriveMetrics(current.Graph, states, expectedResolutions, prior != nil)
	if !reflect.DeepEqual(got.Metrics, expectedMetrics) {
		return Judgment{}, fmt.Errorf("receipt edge algebra metrics are not independently reproducible")
	}
	expectedDecision := decisionFor(observed.Predicate, prior != nil)
	if !reflect.DeepEqual(got.Decision, expectedDecision) {
		return Judgment{}, fmt.Errorf("receipt decision is not independently reproducible")
	}
	if got.TransitionHeadDigest != got.Transitions[len(got.Transitions)-1].TransitionDigest || receiptDigest(got) != got.Digest {
		return Judgment{}, fmt.Errorf("receipt digest or transition head is invalid")
	}
	judgment := Judgment{
		Schema: JudgmentSchema, ReceiptDigest: got.Digest, Predicate: observed.Predicate,
		Decision: expectedDecision.Value, Resolution: expectedDecision.Resolution, Reason: expectedDecision.Reason,
		Accepted: true, IndependentReplay: "RAW_GOOO_PARSE_LOWER_GRAPH_AND_TRANSITION_REDERIVED",
		Metrics: expectedMetrics, ReadOnly: got.Subject.ReadOnly && got.Subject.RepositoryWrites == 0,
		SemanticPromotionAuthorized: false, SourceReconstruction: "syntax.ParseFile->bidir.Lower->semantic.IR",
		ProducerPackageImportNumerator: 0, ProducerPackageImportDenominator: 1,
		AppendOnlyRecoveryChainTotal: boolInt(prior != nil),
	}
	judgment.Digest = digestJSON(judgment)
	return judgment, nil
}

func reconstruct(source []byte, sourcePath string) (reconstructed, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return reconstructed{}, fmt.Errorf("consumer parse failed: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return reconstructed{}, fmt.Errorf("consumer lower failed: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return reconstructed{}, err
	}
	activities := map[string]semantic.Node{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity {
			activities[node.Name] = node
		}
	}
	claims := []claim{}
	activityIndex := map[string]int{}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		node, ok := activities[activity.Name]
		if !ok || node.ValueProgram == "" {
			return reconstructed{}, fmt.Errorf("consumer cannot bind activity %q", activity.Name)
		}
		activityIndex[node.ID.String()] = len(claims)
		claims = append(claims, claim{
			Ordinal: len(claims) + 1, Axis: strings.ToLower(activity.Name), ClaimID: node.ID.String(), ActivityID: node.ID.String(),
			ActivityName: activity.Name, Statement: fmt.Sprintf("activity %s declares value claim %q", activity.Name, node.ValueProgram), ValueProgram: node.ValueProgram,
			Producer: producerID, Consumer: consumerID, MetaOperation: operation, ProofChoice: proof,
			Coordinate: coordinate{Stage: "CLAIM", Step: activity.Name, Reason: "SEMANTIC_ACTIVITY_VALUE"},
		})
	}
	if len(claims) != claimTotal {
		return reconstructed{}, fmt.Errorf("consumer reconstructed %d claims, want %d", len(claims), claimTotal)
	}
	generatedBy := map[string]string{}
	usedBy := map[string][]string{}
	for _, fact := range ir.Graph.AllFacts() {
		switch fact.Predicate {
		case semantic.WasGeneratedBy:
			generatedBy[fact.Subject.String()] = fact.Object.String()
		case semantic.Used:
			usedBy[fact.Subject.String()] = append(usedBy[fact.Subject.String()], fact.Object.String())
		}
	}
	type candidate struct {
		from, to int
		kind     edgeKind
	}
	candidates := []candidate{}
	for downstreamID, entities := range usedBy {
		to, ok := activityIndex[downstreamID]
		if !ok {
			continue
		}
		for _, entityID := range entities {
			upstreamID, ok := generatedBy[entityID]
			if !ok {
				continue
			}
			from, ok := activityIndex[upstreamID]
			if !ok || from == to {
				continue
			}
			kind, ok := edgeKind(claims[to].ValueProgram)
			if !ok {
				return reconstructed{}, fmt.Errorf("consumer found untyped dependency")
			}
			candidates = append(candidates, candidate{from: from, to: to, kind: kind})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].from != candidates[j].from {
			return candidates[i].from < candidates[j].from
		}
		if candidates[i].to != candidates[j].to {
			return candidates[i].to < candidates[j].to
		}
		return candidates[i].kind < candidates[j].kind
	})
	if len(candidates) != edgeTotal {
		return reconstructed{}, fmt.Errorf("consumer reconstructed %d edges, want %d", len(candidates), edgeTotal)
	}
	edges := make([]edge, len(candidates))
	for index, candidate := range candidates {
		edges[index] = edge{EdgeID: fmt.Sprintf("E%02d", index+1), FromClaimID: claims[candidate.from].ClaimID, ToClaimID: claims[candidate.to].ClaimID, Kind: candidate.kind, SemanticBasis: "prov:wasGeneratedBy + prov:used + activity.value-program"}
	}
	result := graph{Schema: "gooo.meta.claim-dependency-graph/v2", Authority: "CANONICAL_IR_FROM_SYNTAX_PARSE_AND_BIDIR_LOWER", Completeness: "CLOSED_WORLD_SOURCE_RECONSTRUCTED", CanonicalIRDigest: prefixedDigest(ir.StableHash()), NodeTotal: claimTotal, EdgeTotal: edgeTotal, Nodes: claims, Edges: edges}
	result.Digest = graphDigest(result)
	return reconstructed{Graph: result, RootProgram: claims[0].ValueProgram}, nil
}

func edgeKind(program string) (edgeKind, bool) {
	const prefix = "claim.edge:"
	if !strings.HasPrefix(program, prefix) {
		return "", false
	}
	switch strings.TrimPrefix(program, prefix) {
	case "supports":
		return supports, true
	case "requires":
		return requires, true
	case "contradicts":
		return contradicts, true
	case "failure-entailment":
		return failureEntailment, true
	default:
		return "", false
	}
}
func validateObservation(current reconstructed, observed observation) error {
	if observed.Schema != "gooo.meta.claim-dependency-observation/v1" || observed.SubjectClaimID != current.Graph.Nodes[0].ClaimID || !observed.ReadOnly || observed.RepositoryWrites != 0 || observed.MutationAuthority || observationDigest(observed) != observed.Digest {
		return fmt.Errorf("consumer observation is invalid")
	}
	switch observed.Predicate {
	case unknown, evidence:
		if current.RootProgram != "claim.observe:recoverable" {
			return fmt.Errorf("observation does not match source recoverable predicate")
		}
	case contradict:
		if current.RootProgram != "claim.observe:contradiction" {
			return fmt.Errorf("observation does not match source contradiction predicate")
		}
	default:
		return fmt.Errorf("consumer predicate is unknown")
	}
	if (observed.Predicate == unknown) != (observed.EvidenceDigest == "") {
		return fmt.Errorf("observation evidence presence is inconsistent with predicate")
	}
	return nil
}

func classify(g graph, predicate, evidenceDigest string) ([]string, []transition) {
	states := make([]string, len(g.Nodes))
	outcomes := make([]transition, len(g.Nodes))
	for index := range g.Nodes {
		state, event, reason := "OPEN", "DEPENDENCY_BLOCKED", "UPSTREAM_UNKNOWN_OR_NON_REFUTING"
		if index == 0 {
			switch predicate {
			case contradict:
				state, event, reason = "REFUTED", "EXPLICIT_CONTRADICTION", "OBSERVATION_PREDICATE_EXPLICITLY_CONTRADICTS"
			case evidence:
				state, event, reason = "DISCHARGED", "EVIDENCE_ACCEPTED", "OBSERVATION_EVIDENCE_PREDICATE_SATISFIED"
			default:
				state, event, reason = "OPEN", "OBSERVATION_UNKNOWN", "OBSERVATION_PREDICATE_UNKNOWN"
			}
		} else if predicate == evidence {
			state, event, reason = "DISCHARGED", "DEPENDENCY_DISCHARGED", "EVIDENCE_PREDICATE_SATISFIED"
		} else if predicate == contradict && hasExplicitRefutation(index, g, states) {
			state, event, reason = "REFUTED", "DEPENDENCY_REFUTED", "EXPLICIT_REFUTING_EDGE"
		}
		states[index] = state
		outcomes[index] = transition{ClaimID: g.Nodes[index].ClaimID, Event: event, Before: "OPEN", After: state, Coordinate: coordinate{Stage: stage(index), Step: g.Nodes[index].ActivityName, Reason: reason}, EvidenceDigest: evidenceDigest}
	}
	return states, outcomes
}
func hasExplicitRefutation(index int, g graph, states []string) bool {
	for _, e := range g.Edges {
		from := indexOf(e.FromClaimID, g)
		if e.ToClaimID == g.Nodes[index].ClaimID && from >= 0 && states[from] == "REFUTED" && (e.Kind == contradicts || e.Kind == failureEntailment) {
			return true
		}
	}
	return false
}
func stage(index int) string {
	if index == 0 {
		return "OBSERVE"
	}
	return "PROPAGATE"
}

func expectedTransitionsFor(g graph, outcomes []transition, provenance string, prior *receipt) ([]transition, error) {
	result := []transition{}
	if prior != nil {
		result = append(result, prior.Transitions...)
		previous := result[len(result)-1].TransitionDigest
		for index, claim := range g.Nodes {
			reason := "EVIDENCE_PREDICATE_SATISFIED"
			event := "DEPENDENCY_DISCHARGED"
			if index == 0 {
				reason, event = "RECOVERY_EVIDENCE_PREDICATE_SATISFIED", "EVIDENCE_ACCEPTED"
			}
			value := transition{Sequence: len(result) + 1, ClaimID: claim.ClaimID, Event: event, Before: prior.Resolutions[index].State, After: "DISCHARGED", Coordinate: coordinate{Stage: "RECOVER", Step: claim.ActivityName, Reason: reason}, EvidenceDigest: outcomes[0].EvidenceDigest, Provenance: provenance, PreviousTransitionDigest: previous}
			value.TransitionDigest = transitionDigest(value)
			result = append(result, value)
			previous = value.TransitionDigest
		}
		return result, nil
	}
	previous := ""
	for _, claim := range g.Nodes {
		value := transition{Sequence: len(result) + 1, ClaimID: claim.ClaimID, Event: "CLAIM_REGISTERED", Before: "UNRECORDED", After: "OPEN", Coordinate: coordinate{Stage: "DECLARE", Step: claim.ActivityName, Reason: "CLAIM_REGISTERED"}, Provenance: provenance, PreviousTransitionDigest: previous}
		value.TransitionDigest = transitionDigest(value)
		result = append(result, value)
		previous = value.TransitionDigest
	}
	for _, outcome := range outcomes {
		outcome.Sequence = len(result) + 1
		outcome.PreviousTransitionDigest = previous
		outcome.Provenance = provenance
		outcome.TransitionDigest = transitionDigest(outcome)
		result = append(result, outcome)
		previous = outcome.TransitionDigest
	}
	return result, nil
}

func buildResolutions(g graph, states []string, outcomes []transition, sourceDigest, semanticDigest string, recovered bool) []resolution {
	root := g.Nodes[0].ClaimID
	rootCoordinate := outcomes[0].Coordinate
	provenance := fmt.Sprintf("source:%s|ir:%s|producer:%s|consumer:%s", sourceDigest, semanticDigest, producerID, consumerID)
	result := make([]resolution, len(g.Nodes))
	for index, claim := range g.Nodes {
		path, ids, kinds := shortestPath(index, g)
		coord := outcomes[index].Coordinate
		result[index] = resolution{ClaimID: claim.ClaimID, Axis: claim.Axis, State: states[index], Kind: resolutionKind(index, states[index], recovered), ObservedEvent: outcomes[index].Event, Coordinate: coord, EvidenceDigest: outcomes[index].EvidenceDigest, Provenance: provenance, FailureResponsibility: "LOCAL_PRODUCER", FailureOwnerClaimID: root, CausePath: idsForPath(path, g), CauseEdgeIDs: ids, CauseEdgeKinds: kinds, CauseTransitionDigest: outcomes[0].TransitionDigest, CauseCoordinate: &rootCoordinate}
		if index != 0 {
			result[index].FailureResponsibility = "UPSTREAM_CLAIM"
		}
		if states[index] == "OPEN" {
			result[index].MissingEvidenceIDs = []string{"evidence:" + root}
			result[index].BlockedByClaimIDs, result[index].BlockedByEdgeIDs = blockedFrontier(index, g, states)
		}
	}
	return result
}
func resolutionKind(index int, state string, recovered bool) string {
	if index == 0 {
		if state == "REFUTED" {
			return "DIRECT_REFUTED"
		}
		if state == "DISCHARGED" {
			return "DIRECT_DISCHARGED"
		}
		return "DIRECT_UNKNOWN"
	}
	if state == "REFUTED" {
		return "DEPENDENCY_REFUTED"
	}
	if state == "DISCHARGED" {
		if recovered {
			return "DEPENDENCY_RECOVERED"
		}
		return "DEPENDENCY_DISCHARGED"
	}
	return "DEPENDENCY_BLOCKED"
}
func shortestPath(index int, g graph) ([]int, []string, []edgeKind) {
	if index == 0 {
		return []int{0}, nil, nil
	}
	best := []int(nil)
	for _, e := range g.Edges {
		if e.ToClaimID != g.Nodes[index].ClaimID {
			continue
		}
		from := indexOf(e.FromClaimID, g)
		candidate, _, _ := shortestPath(from, g)
		candidate = append(candidate, index)
		if best == nil || len(candidate) < len(best) || (len(candidate) == len(best) && pathKey(candidate, g) < pathKey(best, g)) {
			best = candidate
		}
	}
	if best == nil {
		return []int{index}, nil, nil
	}
	ids, kinds := []string{}, []edgeKind{}
	for position := 1; position < len(best); position++ {
		for _, e := range g.Edges {
			if e.FromClaimID == g.Nodes[best[position-1]].ClaimID && e.ToClaimID == g.Nodes[best[position]].ClaimID {
				ids = append(ids, e.EdgeID)
				kinds = append(kinds, e.Kind)
				break
			}
		}
	}
	return best, ids, kinds
}
func pathKey(path []int, g graph) string {
	parts := make([]string, len(path))
	for i, value := range path {
		parts[i] = g.Nodes[value].ClaimID
	}
	return strings.Join(parts, "\x00")
}
func idsForPath(path []int, g graph) []string {
	result := make([]string, len(path))
	for i, value := range path {
		result[i] = g.Nodes[value].ClaimID
	}
	return result
}
func indexOf(id string, g graph) int {
	for i, c := range g.Nodes {
		if c.ClaimID == id {
			return i
		}
	}
	return -1
}
func blockedFrontier(index int, g graph, states []string) ([]string, []string) {
	claims, edges := []string{}, []string{}
	for _, e := range g.Edges {
		if e.ToClaimID != g.Nodes[index].ClaimID {
			continue
		}
		from := indexOf(e.FromClaimID, g)
		if from >= 0 && (states[from] == "OPEN" || (states[from] == "REFUTED" && (e.Kind == supports || e.Kind == requires))) {
			claims = append(claims, e.FromClaimID)
			edges = append(edges, e.EdgeID)
		}
	}
	return claims, edges
}

func deriveMetrics(g graph, states []string, resolutions []resolution, recovered bool) metrics {
	result := metrics{FixedClaimTotal: claimTotal, FixedEdgeTotal: edgeTotal, ClassifiedClaimTotal: len(resolutions), TransitionTotal: claimTotal * 2, ClassificationBasisPoints: 10000}
	for _, state := range states {
		switch state {
		case "OPEN":
			result.OpenClaimTotal++
		case "DISCHARGED":
			result.DischargedClaimTotal++
		case "REFUTED":
			result.RefutedClaimTotal++
		}
	}
	for _, r := range resolutions {
		switch r.Kind {
		case "DIRECT_UNKNOWN":
			result.UnknownClaimTotal++
			result.DirectUnknownClaimTotal++
		case "DEPENDENCY_BLOCKED":
			result.DependencyBlockedClaimTotal++
		case "DIRECT_REFUTED":
			result.DirectRefutedClaimTotal++
		case "DEPENDENCY_REFUTED":
			result.DependencyRefutedClaimTotal++
		case "DIRECT_DISCHARGED":
			result.DirectDischargedClaimTotal++
		case "DEPENDENCY_DISCHARGED", "DEPENDENCY_RECOVERED":
			result.DependencyDischargedTotal++
		}
		if len(r.CausePath) > result.MaximumCausePathDepth {
			result.MaximumCausePathDepth = len(r.CausePath)
		}
	}
	if recovered {
		result.TransitionTotal += claimTotal
		result.AppendOnlyTransitionTotal = claimTotal
	}
	for _, kind := range []edgeKind{supports, requires, contradicts, failureEntailment} {
		value := edgeMetric{Kind: kind}
		for _, e := range g.Edges {
			if e.Kind != kind {
				continue
			}
			value.Total++
			from, to := indexOf(e.FromClaimID, g), indexOf(e.ToClaimID, g)
			if recovered && states[to] == "DISCHARGED" {
				value.Recovery++
			}
			if states[to] == "OPEN" && (states[from] == "OPEN" || (states[from] == "REFUTED" && (e.Kind == supports || e.Kind == requires))) {
				value.Blocking++
			}
			if states[to] == "REFUTED" && states[from] == "REFUTED" && (e.Kind == contradicts || e.Kind == failureEntailment) {
				value.Refuting++
			}
		}
		result.ObservedBlockingEdgeTotal += value.Blocking
		result.ObservedRefutingEdgeTotal += value.Refuting
		result.ObservedRecoveryEdgeTotal += value.Recovery
		result.EdgeMetrics = append(result.EdgeMetrics, value)
	}
	return result
}
func decisionFor(predicate string, recovered bool) decision {
	if recovered {
		return decision{Value: "PASS", Resolution: "CAUSAL_RECOVERY_DISCHARGED", Reason: "APPEND_ONLY_EVIDENCE_RECOVERY"}
	}
	if predicate == contradict {
		return decision{Value: "FAIL_CLOSED", Resolution: "CAUSAL_REFUTATION", Reason: "EXPLICIT_CONTRADICTION_EDGE_ALGEBRA"}
	}
	return decision{Value: "FAIL_CLOSED", Resolution: "UNRESOLVED_CLAIM", Reason: "UNKNOWN_REMAINS_OPEN"}
}
func validatePrior(current reconstructed, prior receipt) error {
	if prior.Schema != "gooo.meta.claim-dependency-receipt/v2" || prior.Scope != "CLAIM_STATE_PROPAGATION_ONLY" || prior.Observation.Predicate != unknown || prior.Graph.Digest != current.Graph.Digest || len(prior.Resolutions) != claimTotal || receiptDigest(prior) != prior.Digest || !sameStrings(prior.PriorClaimStates, nil) && len(prior.PriorClaimStates) != 0 {
		return fmt.Errorf("prior ledger is not a valid UNKNOWN graph receipt")
	}
	if err := validateChain(prior.Transitions, prior.TransitionHeadDigest); err != nil {
		return err
	}
	for i, r := range prior.Resolutions {
		if r.State != "OPEN" || r.ClaimID != current.Graph.Nodes[i].ClaimID {
			return fmt.Errorf("prior claim state %d is not OPEN", i+1)
		}
	}
	return nil
}
func validateChain(transitions []transition, head string) error {
	if len(transitions) == 0 || transitions[len(transitions)-1].TransitionDigest != head {
		return fmt.Errorf("transition head mismatch")
	}
	previous := ""
	for i, value := range transitions {
		if value.Sequence != i+1 || value.PreviousTransitionDigest != previous || transitionDigest(value) != value.TransitionDigest {
			return fmt.Errorf("transition %d chain mismatch", i+1)
		}
		previous = value.TransitionDigest
	}
	return nil
}
func statesOf(values []resolution) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.State
	}
	return result
}
func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func receiptDigest(value receipt) string {
	value.Digest = ""
	return digestJSON(value)
}
func observationDigest(value observation) string {
	value.Digest = ""
	return digestJSON(value)
}
func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func graphDigest(value graph) string           { value.Digest = ""; return digestJSON(value) }
func transitionDigest(value transition) string { value.TransitionDigest = ""; return digestJSON(value) }
func prefixedDigest(value string) string {
	if strings.HasPrefix(value, "sha256:") {
		return value
	}
	return "sha256:" + value
}
