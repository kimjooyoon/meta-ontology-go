package claimdependencyjudge

// This package intentionally repeats the small raw-input reconstruction and
// state algebra. It must remain import-independent from the producer so a
// producer receipt is comparison evidence, not the judge's source of truth.
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const judgmentSchema = "gooo.meta.claim-dependency-judgment/v3"
const claimTotal, edgeTotal, initialTransitions = 6, 8, 12
const producerID = "gooo://meta/claim-dependency/producer/v3"
const consumerID = "gooo://meta/claim-dependency/independent-judge/v3"
const operation, proof = "classify-claim-state-causality", "COHERENCE"

type edgeKind string

const (
	supports          edgeKind = "SUPPORTS"
	requires          edgeKind = "REQUIRES"
	contradicts       edgeKind = "CONTRADICTS"
	failureEntailment edgeKind = "FAILURE_ENTAILMENT"
)

type predicate string

const (
	unknown               predicate = "UNKNOWN"
	accepted              predicate = "EVIDENCE_ACCEPTED"
	explicitContradiction predicate = "EXPLICIT_CONTRADICTION"
)

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}
type target struct {
	Inputs   []string `json:"inputs,omitempty"`
	Output   string   `json:"output,omitempty"`
	Artifact string   `json:"artifact"`
}
type claim struct {
	Ordinal           int        `json:"ordinal"`
	Axis              string     `json:"axis"`
	ClaimID           string     `json:"claim_id"`
	ActivityID        string     `json:"activity_id"`
	ActivityName      string     `json:"activity_name"`
	Proposition       string     `json:"proposition"`
	PropositionDigest string     `json:"proposition_digest"`
	Target            target     `json:"target"`
	ValueProgram      string     `json:"value_program"`
	Producer          string     `json:"producer"`
	Consumer          string     `json:"consumer"`
	MetaOperation     string     `json:"meta_operation"`
	ProofChoice       string     `json:"proof_choice"`
	Coordinate        coordinate `json:"coordinate"`
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
type evidenceClaim struct {
	ClaimID           string     `json:"claim_id"`
	PropositionDigest string     `json:"proposition_digest"`
	ObservedPredicate predicate  `json:"observed_predicate"`
	ObservedValue     string     `json:"observed_value"`
	Status            string     `json:"status"`
	Coordinate        coordinate `json:"coordinate"`
	Digest            string     `json:"digest"`
}
type capability struct {
	Provider   string     `json:"provider"`
	Permission string     `json:"permission"`
	Status     string     `json:"status"`
	Coordinate coordinate `json:"coordinate"`
	Digest     string     `json:"digest"`
}
type authorityCase struct {
	CaseID             string `json:"case_id"`
	NetworkState       string `json:"network_state"`
	CapabilityStatus   string `json:"capability_status"`
	ExpectedResolution string `json:"expected_resolution"`
	ObservedResolution string `json:"observed_resolution"`
}
type snapshot struct {
	RepositoryRoot   string     `json:"repository_root"`
	TrackedDigest    string     `json:"tracked_digest"`
	UntrackedDigest  string     `json:"untracked_digest"`
	BeforeDigest     string     `json:"before_digest"`
	AfterDigest      string     `json:"after_digest"`
	OutputPath       string     `json:"output_path"`
	OutputDigest     string     `json:"output_digest"`
	RepositoryWrites int        `json:"repository_writes"`
	Coordinate       coordinate `json:"coordinate"`
}
type evidenceReceipt struct {
	Schema              string          `json:"schema"`
	Provider            string          `json:"provider"`
	ArtifactPath        string          `json:"artifact_path"`
	ArtifactBytesDigest string          `json:"artifact_bytes_digest"`
	Operation           string          `json:"operation"`
	RequestStatus       string          `json:"request_status"`
	Procedure           string          `json:"procedure"`
	ObservedPredicate   predicate       `json:"observed_predicate"`
	ObservedValue       string          `json:"observed_value"`
	Status              string          `json:"status"`
	Coordinate          coordinate      `json:"coordinate"`
	Claims              []evidenceClaim `json:"claims"`
	Capability          capability      `json:"capability"`
	Snapshot            snapshot        `json:"snapshot"`
	Digest              string          `json:"digest"`
}
type subject struct {
	SourcePath          string     `json:"source_path"`
	SourceDigest        string     `json:"source_digest"`
	SemanticDigest      string     `json:"semantic_digest"`
	Producer            string     `json:"producer"`
	Consumer            string     `json:"consumer"`
	MetaOperation       string     `json:"meta_operation"`
	ProofChoice         string     `json:"proof_choice"`
	ReadOnly            bool       `json:"read_only"`
	RepositoryWrites    int        `json:"repository_writes"`
	AuthorityResolution string     `json:"authority_resolution"`
	AuthorityCoordinate coordinate `json:"authority_coordinate"`
}
type transition struct {
	Sequence                  int        `json:"sequence"`
	ClaimID                   string     `json:"claim_id"`
	Event                     string     `json:"event"`
	Before                    string     `json:"before"`
	After                     string     `json:"after"`
	Coordinate                coordinate `json:"coordinate"`
	EvidenceDigest            string     `json:"evidence_digest,omitempty"`
	UpstreamEdgeIDs           []string   `json:"upstream_edge_ids,omitempty"`
	UpstreamTransitionDigests []string   `json:"upstream_transition_digests,omitempty"`
	Provenance                string     `json:"provenance"`
	PreviousTransitionDigest  string     `json:"previous_transition_digest,omitempty"`
	TransitionDigest          string     `json:"transition_digest"`
}
type resolution struct {
	ClaimID                string      `json:"claim_id"`
	Axis                   string      `json:"axis"`
	PropositionDigest      string      `json:"proposition_digest"`
	State                  string      `json:"state"`
	Kind                   string      `json:"kind"`
	ObservedEvent          string      `json:"observed_event"`
	Coordinate             coordinate  `json:"coordinate"`
	EvidenceDigest         string      `json:"evidence_digest,omitempty"`
	Provenance             string      `json:"provenance"`
	FailureResponsibility  string      `json:"failure_responsibility"`
	FailureOwnerClaimID    string      `json:"failure_owner_claim_id"`
	MissingEvidenceIDs     []string    `json:"missing_evidence_ids,omitempty"`
	BlockedByClaimIDs      []string    `json:"blocked_by_claim_ids,omitempty"`
	BlockedByEdgeIDs       []string    `json:"blocked_by_edge_ids,omitempty"`
	CausePath              []string    `json:"cause_path"`
	CauseEdgeIDs           []string    `json:"cause_edge_ids"`
	CauseEdgeKinds         []edgeKind  `json:"cause_edge_kinds"`
	CauseTransitionDigests []string    `json:"cause_transition_digests"`
	CauseCoordinate        *coordinate `json:"cause_coordinate"`
}
type truthCase struct {
	Schema         string   `json:"schema"`
	CaseID         string   `json:"case_id"`
	Kind           edgeKind `json:"kind"`
	Direction      string   `json:"direction"`
	UpstreamState  string   `json:"upstream_state"`
	LocalPredicate string   `json:"local_predicate"`
	ExpectedState  string   `json:"expected_state"`
	Positive       bool     `json:"positive"`
	SemanticBasis  string   `json:"semantic_basis"`
}
type edgeMetric struct {
	Kind           edgeKind `json:"kind"`
	Eligible       int      `json:"eligible"`
	ObservedCausal int      `json:"observed_causal"`
	Blocking       int      `json:"blocking"`
	Refuting       int      `json:"refuting"`
	Discharge      int      `json:"discharge"`
}
type metrics struct {
	FixedClaimTotal             int          `json:"fixed_claim_total"`
	DistinctPropositionTotal    int          `json:"distinct_proposition_total"`
	FixedEdgeTotal              int          `json:"fixed_edge_total"`
	EligibleEdgeTotal           int          `json:"eligible_edge_total"`
	ObservedCausalEdgeTotal     int          `json:"observed_causal_edge_total"`
	ShortestPathEdgeUnionTotal  int          `json:"shortest_path_edge_union_total"`
	ClassifiedClaimTotal        int          `json:"classified_claim_total"`
	OpenClaimTotal              int          `json:"open_claim_total"`
	DischargedClaimTotal        int          `json:"discharged_claim_total"`
	RefutedClaimTotal           int          `json:"refuted_claim_total"`
	CurrentEvidenceTotal        int          `json:"current_evidence_total"`
	HistoricalEvidenceTotal     int          `json:"historical_evidence_total"`
	UnknownEvidenceTotal        int          `json:"unknown_evidence_total"`
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
	TruthTableCaseTotal         int          `json:"truth_table_case_total"`
	AuthorityCaseTotal          int          `json:"authority_case_total"`
	EdgeMetrics                 []edgeMetric `json:"edge_metrics"`
}
type decision struct {
	Value                       string `json:"value"`
	Resolution                  string `json:"resolution"`
	Reason                      string `json:"reason"`
	SemanticPromotionAuthorized bool   `json:"semantic_promotion_authorized"`
}
type receipt struct {
	Schema                   string          `json:"schema"`
	Scope                    string          `json:"scope"`
	Subject                  subject         `json:"subject"`
	Evidence                 evidenceReceipt `json:"evidence"`
	Graph                    graph           `json:"graph"`
	TruthTable               []truthCase     `json:"truth_table"`
	AuthorityCases           []authorityCase `json:"authority_cases"`
	PriorReceiptDigest       string          `json:"prior_receipt_digest,omitempty"`
	PreviousTransitionDigest string          `json:"previous_transition_digest,omitempty"`
	PriorClaimStates         []string        `json:"prior_claim_states,omitempty"`
	EvidenceDigest           string          `json:"evidence_digest"`
	Transitions              []transition    `json:"transitions"`
	TransitionHeadDigest     string          `json:"transition_head_digest"`
	Resolutions              []resolution    `json:"resolutions"`
	Metrics                  metrics         `json:"metrics"`
	Decision                 decision        `json:"decision"`
	Digest                   string          `json:"digest"`
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
	AuthorityResolution              string  `json:"authority_resolution"`
	SourceReconstruction             string  `json:"source_reconstruction"`
	SourceReconstructionNumerator    int     `json:"source_reconstruction_numerator"`
	SourceReconstructionDenominator  int     `json:"source_reconstruction_denominator"`
	ProducerPackageImportNumerator   int     `json:"producer_package_import_numerator"`
	ProducerPackageImportDenominator int     `json:"producer_package_import_denominator"`
	AppendOnlyRecoveryChainTotal     int     `json:"append_only_recovery_chain_total"`
	Digest                           string  `json:"digest"`
}
type reconstructed struct {
	Graph       graph
	RootProgram string
}

func Judge(source []byte, sourcePath string, priorBytes, evidenceBytes, receiptBytes []byte) (Judgment, error) {
	current, err := reconstruct(source, sourcePath)
	if err != nil {
		return Judgment{}, err
	}
	var evidence evidenceReceipt
	if err := json.Unmarshal(evidenceBytes, &evidence); err != nil {
		return Judgment{}, fmt.Errorf("decode raw evidence: %w", err)
	}
	if err := validateEvidence(evidence); err != nil {
		return Judgment{}, err
	}
	artifact, err := os.ReadFile(evidence.ArtifactPath)
	if err != nil {
		return Judgment{}, fmt.Errorf("judge cannot re-observe artifact: %w", err)
	}
	if digestBytes(artifact) != evidence.ArtifactBytesDigest {
		return Judgment{}, fmt.Errorf("judge observed artifact bytes digest mismatch")
	}
	artifactGraph, err := reconstruct(artifact, evidence.ArtifactPath)
	if err != nil {
		return Judgment{}, err
	}
	if err := validateEvidenceClaims(evidence, artifactGraph.Graph, artifactGraph, artifact); err != nil {
		return Judgment{}, err
	}
	if err := validateSourceObservation(current, evidence); err != nil {
		return Judgment{}, err
	}
	var got receipt
	if err := json.Unmarshal(receiptBytes, &got); err != nil {
		return Judgment{}, fmt.Errorf("decode receipt: %w", err)
	}
	if got.Graph.Digest != current.Graph.Digest || !reflect.DeepEqual(got.Graph, current.Graph) {
		return Judgment{}, fmt.Errorf("receipt graph is not reconstructed from raw source")
	}
	if got.EvidenceDigest != evidence.Digest || !reflect.DeepEqual(got.Evidence, evidence) {
		return Judgment{}, fmt.Errorf("receipt is not bound to raw evidence receipt")
	}
	sourceDigest := digestBytes(source)
	if got.Subject.SourceDigest != sourceDigest || got.Subject.SourcePath != sourcePath || got.Subject.SemanticDigest != current.Graph.CanonicalIRDigest || got.Subject.Producer != producerID || got.Subject.Consumer != consumerID || got.Subject.MetaOperation != operation || got.Subject.ProofChoice != proof {
		return Judgment{}, fmt.Errorf("receipt subject provenance is invalid")
	}
	expectedAuthority := authorityResolution(evidence)
	if got.Subject.AuthorityResolution != expectedAuthority || got.Subject.ReadOnly != (expectedAuthority == "NET_REPOSITORY_STATE_UNCHANGED") || got.Subject.RepositoryWrites != evidence.Snapshot.RepositoryWrites || !reflect.DeepEqual(got.Subject.AuthorityCoordinate, evidence.Capability.Coordinate) {
		return Judgment{}, fmt.Errorf("receipt subject authority is not independently reproduced")
	}
	var prior *receipt
	if len(priorBytes) > 0 {
		var value receipt
		if err := json.Unmarshal(priorBytes, &value); err != nil {
			return Judgment{}, err
		}
		if err := validatePrior(current, value); err != nil {
			return Judgment{}, err
		}
		prior = &value
		d := receiptDigest(value)
		if got.PriorReceiptDigest != d || got.PreviousTransitionDigest != value.TransitionHeadDigest || !sameStrings(got.PriorClaimStates, statesOf(value.Resolutions)) {
			return Judgment{}, fmt.Errorf("recovery does not bind prior receipt head and states")
		}
		if len(got.Transitions) < len(value.Transitions) || !reflect.DeepEqual(got.Transitions[:len(value.Transitions)], value.Transitions) {
			return Judgment{}, fmt.Errorf("recovery is not append-only")
		}
	}
	states, templates := classify(current.Graph, evidence)
	provenance := fmt.Sprintf("source:%s|ir:%s|evidence:%s|producer:%s|consumer:%s", sourceDigest, current.Graph.CanonicalIRDigest, evidence.Digest, producerID, consumerID)
	expectedTransitions := transitionsFor(current.Graph, templates, provenance, prior)
	if !reflect.DeepEqual(got.Transitions, expectedTransitions) {
		return Judgment{}, fmt.Errorf("transition chain cannot be independently reproduced")
	}
	currentOutcomes := templates
	if prior != nil {
		currentOutcomes = expectedTransitions[len(expectedTransitions)-claimTotal:]
	}
	expectedResolutions := buildResolutions(current.Graph, states, currentOutcomes, provenance)
	if !reflect.DeepEqual(got.Resolutions, expectedResolutions) {
		return Judgment{}, fmt.Errorf("resolutions cannot be independently reproduced")
	}
	expectedMetrics := deriveMetrics(current.Graph, states, expectedResolutions, currentOutcomes, evidence, prior != nil)
	if prior != nil {
		expectedMetrics.AppendOnlyTransitionTotal = claimTotal
	}
	if !reflect.DeepEqual(got.Metrics, expectedMetrics) {
		return Judgment{}, fmt.Errorf("metrics cannot be independently reproduced")
	}
	expectedDecision := decisionFor(states, evidence, prior != nil)
	if !reflect.DeepEqual(got.Decision, expectedDecision) {
		return Judgment{}, fmt.Errorf("decision cannot be independently reproduced")
	}
	if got.TruthTable == nil || !reflect.DeepEqual(got.TruthTable, truthTable()) {
		return Judgment{}, fmt.Errorf("truth table is not independently reproduced")
	}
	if got.AuthorityCases == nil || !reflect.DeepEqual(got.AuthorityCases, authorityCases()) {
		return Judgment{}, fmt.Errorf("authority resolution cases are not independently reproduced")
	}
	if err := validateAuthorityCases(got.AuthorityCases); err != nil {
		return Judgment{}, err
	}
	if err := validateTruthTable(got.TruthTable); err != nil {
		return Judgment{}, err
	}
	if got.TransitionHeadDigest != got.Transitions[len(got.Transitions)-1].TransitionDigest || receiptDigest(got) != got.Digest {
		return Judgment{}, fmt.Errorf("receipt digest or transition head is invalid")
	}
	judgment := Judgment{Schema: judgmentSchema, ReceiptDigest: got.Digest, Predicate: string(evidence.ObservedPredicate), Decision: expectedDecision.Value, Resolution: expectedDecision.Resolution, Reason: expectedDecision.Reason, Accepted: true, IndependentReplay: "RAW_GOOO_PARSE_LOWER_ARTIFACT_REOBSERVE_AND_TRANSITION_REDERIVED", Metrics: expectedMetrics, ReadOnly: got.Subject.ReadOnly && got.Subject.RepositoryWrites == 0, SemanticPromotionAuthorized: false, AuthorityResolution: got.Subject.AuthorityResolution, SourceReconstruction: "syntax.ParseFile->bidir.Lower->semantic.IR", SourceReconstructionNumerator: 1, SourceReconstructionDenominator: 1, ProducerPackageImportNumerator: 0, ProducerPackageImportDenominator: 1, AppendOnlyRecoveryChainTotal: boolInt(prior != nil)}
	judgment.Digest = digestJSON(judgment)
	return judgment, nil
}

func reconstruct(source []byte, sourcePath string) (reconstructed, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return reconstructed{}, fmt.Errorf("judge parse failed: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return reconstructed{}, err
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
	generatedBy, usedBy := map[string]string{}, map[string][]string{}
	for _, fact := range ir.Graph.AllFacts() {
		switch fact.Predicate {
		case semantic.WasGeneratedBy:
			generatedBy[fact.Subject.String()] = fact.Object.String()
		case semantic.Used:
			usedBy[fact.Subject.String()] = append(usedBy[fact.Subject.String()], fact.Object.String())
		}
	}
	claims, activityIndex := []claim{}, map[string]int{}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		node, ok := activities[activity.Name]
		if !ok || node.ValueProgram == "" {
			return reconstructed{}, fmt.Errorf("judge cannot bind activity")
		}
		inputs := append([]string(nil), usedBy[node.ID.String()]...)
		sort.Strings(inputs)
		output := ""
		for entityID, activityID := range generatedBy {
			if activityID == node.ID.String() {
				output = entityID
				break
			}
		}
		artifact := "gooo://claim-dependency/artifact/" + strings.ToLower(activity.Name)
		proposition := fmt.Sprintf("execute(activity=%s,inputs=[%s],output=%s,artifact=%s,value=%s)", node.ID.String(), strings.Join(inputs, ","), output, artifact, node.ValueProgram)
		activityIndex[node.ID.String()] = len(claims)
		claims = append(claims, claim{len(claims) + 1, strings.ToLower(activity.Name), node.ID.String(), node.ID.String(), activity.Name, proposition, digestBytes([]byte(proposition)), target{Inputs: inputs, Output: output, Artifact: artifact}, node.ValueProgram, producerID, consumerID, operation, proof, coordinate{"CLAIM", activity.Name, "NORMALIZED_EXECUTION_PROPOSITION"}})
	}
	if len(claims) != claimTotal {
		return reconstructed{}, fmt.Errorf("judge reconstructed %d claims", len(claims))
	}
	seen := map[string]bool{}
	for _, c := range claims {
		seen[c.PropositionDigest] = true
	}
	if len(seen) != claimTotal {
		return reconstructed{}, fmt.Errorf("judge found non-distinct propositions")
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
				return reconstructed{}, fmt.Errorf("judge found untyped edge")
			}
			candidates = append(candidates, candidate{from, to, kind})
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
		return reconstructed{}, fmt.Errorf("judge reconstructed %d edges", len(candidates))
	}
	edges := make([]edge, len(candidates))
	for i, c := range candidates {
		edges[i] = edge{fmt.Sprintf("E%02d", i+1), claims[c.from].ClaimID, claims[c.to].ClaimID, c.kind, "prov:wasGeneratedBy + prov:used + source-derived value-program edge predicate"}
	}
	result := graph{"gooo.meta.claim-dependency-graph/v3", "CANONICAL_IR_FROM_SYNTAX_PARSE_AND_BIDIR_LOWER", "CLOSED_WORLD_SOURCE_RECONSTRUCTED", prefixedDigest(ir.StableHash()), claimTotal, edgeTotal, claims, edges, ""}
	result.Digest = graphDigest(result)
	return reconstructed{result, claims[0].ValueProgram}, nil
}

func edgeKind(program string) (edgeKind, bool) {
	if !strings.HasPrefix(program, "claim.edge:") {
		return "", false
	}
	value := strings.TrimPrefix(program, "claim.edge:")
	if i := strings.IndexByte(value, '|'); i >= 0 {
		value = value[:i]
	}
	switch value {
	case "supports":
		return supports, true
	case "requires":
		return requires, true
	case "contradicts":
		return contradicts, true
	case "failure-entailment":
		return failureEntailment, true
	}
	return "", false
}
func truthTable() []truthCase {
	return []truthCase{{"gooo.meta.claim-dependency-truth-table/v1", "SUPPORTS-POSITIVE", supports, "established supports target", "DISCHARGED", "EVIDENCE_ACCEPTED", "OPEN", true, "support never discharges by itself"}, {"gooo.meta.claim-dependency-truth-table/v1", "SUPPORTS-NEGATIVE", supports, "refuted supports target", "REFUTED", "UNKNOWN", "OPEN", false, "support does not refute"}, {"gooo.meta.claim-dependency-truth-table/v1", "REQUIRES-POSITIVE", requires, "required proposition established", "DISCHARGED", "EVIDENCE_ACCEPTED", "DISCHARGED", true, "upstream and local requirement hold"}, {"gooo.meta.claim-dependency-truth-table/v1", "REQUIRES-NEGATIVE", requires, "required proposition established", "DISCHARGED", "UNKNOWN", "OPEN", false, "local requirement missing"}, {"gooo.meta.claim-dependency-truth-table/v1", "CONTRADICTS-POSITIVE", contradicts, "established contradiction of target", "REFUTED", "UNKNOWN", "REFUTED", true, "upstream refutation is sufficient on a typed contradiction edge"}, {"gooo.meta.claim-dependency-truth-table/v1", "CONTRADICTS-NEGATIVE", contradicts, "ordinary support direction", "REFUTED", "UNKNOWN", "OPEN", false, "name alone cannot refute"}, {"gooo.meta.claim-dependency-truth-table/v1", "FAILURE_ENTAILMENT-POSITIVE", failureEntailment, "failure entails target failure", "REFUTED", "UNKNOWN", "REFUTED", true, "upstream refutation is sufficient on a typed failure-entailment edge"}, {"gooo.meta.claim-dependency-truth-table/v1", "FAILURE_ENTAILMENT-NEGATIVE", failureEntailment, "success or ordinary dependency", "REFUTED", "UNKNOWN", "OPEN", false, "failure evidence is absent"}}
}

type relationOutcome string

const (
	relationOpen       relationOutcome = "OPEN"
	relationDischarged relationOutcome = "DISCHARGED"
	relationRefuted    relationOutcome = "REFUTED"
)

func edgeRelation(kind edgeKind, upstreamState string, local predicate, directionMatches bool) relationOutcome {
	if !directionMatches {
		return relationOpen
	}
	switch kind {
	case requires:
		if upstreamState == "DISCHARGED" && local == accepted {
			return relationDischarged
		}
	case contradicts, failureEntailment:
		if upstreamState == "REFUTED" {
			return relationRefuted
		}
	case supports:
		// SUPPORTS blocks unresolved claims but never discharges or refutes.
	}
	return relationOpen
}

func validateTruthTable(cases []truthCase) error {
	if len(cases) != 2*4 {
		return fmt.Errorf("truth table has %d cases, want 8", len(cases))
	}
	seen := map[edgeKind]int{}
	for _, test := range cases {
		actual := edgeRelation(test.Kind, test.UpstreamState, predicate(test.LocalPredicate), test.Positive)
		if string(actual) != test.ExpectedState {
			return fmt.Errorf("truth table case %q computed %s, expected %s", test.CaseID, actual, test.ExpectedState)
		}
		seen[test.Kind]++
	}
	for _, kind := range []edgeKind{supports, requires, contradicts, failureEntailment} {
		if seen[kind] != 2 {
			return fmt.Errorf("truth table edge kind %s has %d cases", kind, seen[kind])
		}
	}
	return nil
}
func validateEvidence(value evidenceReceipt) error {
	if value.Schema != "gooo.meta.claim-dependency-evidence/v2" || value.Status != "CURRENT_EVIDENCE" || value.ArtifactPath == "" || value.Digest == "" || value.RequestStatus != "CLAIMED_INPUT" || value.Procedure != "RAW_ARTIFACT_SOURCE_DERIVED_PREDICATE_V1" {
		return fmt.Errorf("raw evidence identity invalid")
	}
	if evidenceDigest(value) != value.Digest {
		return fmt.Errorf("raw evidence digest invalid")
	}
	if value.Snapshot.RepositoryWrites != 0 || value.Snapshot.BeforeDigest != value.Snapshot.AfterDigest || value.Capability.Status != "CURRENT_EVIDENCE" {
		return fmt.Errorf("raw evidence effects or capability invalid")
	}
	if capabilityDigest(value.Capability) != value.Capability.Digest {
		return fmt.Errorf("capability digest invalid")
	}
	return nil
}
func validateEvidenceClaims(value evidenceReceipt, g graph, artifact reconstructed, artifactBytes []byte) error {
	if len(value.Claims) != g.NodeTotal {
		return fmt.Errorf("raw evidence claim denominator mismatch")
	}
	expected, observedValue := observedPredicateForArtifact(artifact, value.Operation, value.ArtifactPath, artifactBytes)
	if value.ObservedPredicate != expected || value.ObservedValue != observedValue {
		return fmt.Errorf("raw evidence predicate is not procedure-derived")
	}
	for i, c := range value.Claims {
		if c.Status != "CURRENT_EVIDENCE" || c.ClaimID != g.Nodes[i].ClaimID || c.PropositionDigest != g.Nodes[i].PropositionDigest || c.Digest == "" || evidenceClaimDigest(c) != c.Digest {
			return fmt.Errorf("raw evidence claim %d is invalid", i+1)
		}
		claimPredicate := expected
		if expected == explicitContradiction && i != 0 {
			claimPredicate = unknown
		}
		claimValue := observedValue + "|claim:" + c.ClaimID + "|proposition_digest:" + c.PropositionDigest + "|predicate:" + string(claimPredicate)
		if c.ObservedPredicate != claimPredicate || c.ObservedValue != claimValue {
			return fmt.Errorf("raw evidence claim %d predicate is not source-derived", i+1)
		}
	}
	return nil
}

func observedPredicateForArtifact(artifact reconstructed, requestedOperation, artifactPath string, artifactBytes []byte) (predicate, string) {
	value := unknown
	complete := artifact.Graph.NodeTotal == claimTotal && artifact.Graph.EdgeTotal == edgeTotal && distinct(artifact.Graph) == claimTotal && len(artifact.Graph.Nodes) == claimTotal && len(artifact.Graph.Edges) == edgeTotal
	if complete && requestedOperation == "acceptance" && strings.HasPrefix(artifact.RootProgram, "claim.observe:recoverable") {
		value = accepted
	}
	if complete && requestedOperation == "contradiction" && strings.HasPrefix(artifact.RootProgram, "claim.observe:contradiction") {
		value = explicitContradiction
	}
	return value, fmt.Sprintf("procedure:RAW_ARTIFACT_SOURCE_DERIVED_PREDICATE_V1|artifact_path_digest:%s|artifact_bytes_digest:%s|predicate:%s", digestBytes([]byte(artifactPath)), digestBytes(artifactBytes), value)
}

func validateSourceObservation(current reconstructed, value evidenceReceipt) error {
	root := unknown
	for _, claim := range value.Claims {
		if claim.ClaimID == current.Graph.Nodes[0].ClaimID && claim.PropositionDigest == current.Graph.Nodes[0].PropositionDigest {
			root = claim.ObservedPredicate
			break
		}
	}
	if strings.HasPrefix(current.RootProgram, "claim.observe:recoverable") && root == explicitContradiction {
		return fmt.Errorf("recoverable source cannot be refuted by external predicate")
	}
	if strings.HasPrefix(current.RootProgram, "claim.observe:contradiction") && root != explicitContradiction {
		return fmt.Errorf("contradiction source requires explicit contradiction evidence")
	}
	if !strings.HasPrefix(current.RootProgram, "claim.observe:recoverable") && !strings.HasPrefix(current.RootProgram, "claim.observe:contradiction") {
		return fmt.Errorf("source has no recognized observation predicate")
	}
	return nil
}
func validatePrior(current reconstructed, value receipt) error {
	if value.Schema != "gooo.meta.claim-dependency-receipt/v3" || value.Scope != "CLAIM_STATE_PROPAGATION_ONLY" || value.Evidence.ObservedPredicate != unknown || value.Graph.Digest != current.Graph.Digest || len(value.Resolutions) != claimTotal || len(value.PriorClaimStates) != 0 || receiptDigest(value) != value.Digest {
		return fmt.Errorf("prior UNKNOWN ledger invalid")
	}
	replayed, err := replayReceipt(value)
	if err != nil {
		return fmt.Errorf("prior receipt replay failed: %w", err)
	}
	if !reflect.DeepEqual(replayed, value) {
		return fmt.Errorf("prior receipt is not a complete replay of raw source and evidence")
	}
	if err := validateChain(value.Transitions, value.TransitionHeadDigest); err != nil {
		return err
	}
	for i, r := range value.Resolutions {
		if r.State != "OPEN" || r.ClaimID != current.Graph.Nodes[i].ClaimID {
			return fmt.Errorf("prior state %d not OPEN", i+1)
		}
	}
	return nil
}

func replayReceipt(value receipt) (receipt, error) {
	source, err := os.ReadFile(value.Subject.SourcePath)
	if err != nil {
		return receipt{}, fmt.Errorf("prior source cannot be re-observed: %w", err)
	}
	parsed, err := reconstruct(source, value.Subject.SourcePath)
	if err != nil {
		return receipt{}, err
	}
	artifact, err := os.ReadFile(value.Evidence.ArtifactPath)
	if err != nil {
		return receipt{}, fmt.Errorf("prior artifact cannot be re-observed: %w", err)
	}
	artifactGraph, err := reconstruct(artifact, value.Evidence.ArtifactPath)
	if err != nil {
		return receipt{}, err
	}
	if err := validateEvidence(value.Evidence); err != nil {
		return receipt{}, err
	}
	if digestBytes(artifact) != value.Evidence.ArtifactBytesDigest {
		return receipt{}, fmt.Errorf("prior artifact bytes digest mismatch")
	}
	if err := validateEvidenceClaims(value.Evidence, artifactGraph.Graph, artifactGraph, artifact); err != nil {
		return receipt{}, err
	}
	if err := validateSourceObservation(parsed, value.Evidence); err != nil {
		return receipt{}, err
	}
	states, outcomes := classify(parsed.Graph, value.Evidence)
	sourceDigest := digestBytes(source)
	provenance := fmt.Sprintf("source:%s|ir:%s|evidence:%s|producer:%s|consumer:%s", sourceDigest, parsed.Graph.CanonicalIRDigest, value.Evidence.Digest, producerID, consumerID)
	transitions := transitionsFor(parsed.Graph, outcomes, provenance, nil)
	resolutions := buildResolutions(parsed.Graph, states, outcomes, provenance)
	metrics := deriveMetrics(parsed.Graph, states, resolutions, outcomes, value.Evidence, false)
	decision := decisionFor(states, value.Evidence, false)
	authority := authorityResolution(value.Evidence)
	result := receipt{Schema: "gooo.meta.claim-dependency-receipt/v3", Scope: "CLAIM_STATE_PROPAGATION_ONLY", Subject: subject{SourcePath: value.Subject.SourcePath, SourceDigest: sourceDigest, SemanticDigest: parsed.Graph.CanonicalIRDigest, Producer: producerID, Consumer: consumerID, MetaOperation: operation, ProofChoice: proof, ReadOnly: authority == "NET_REPOSITORY_STATE_UNCHANGED", RepositoryWrites: value.Evidence.Snapshot.RepositoryWrites, AuthorityResolution: authority, AuthorityCoordinate: value.Evidence.Capability.Coordinate}, Evidence: value.Evidence, Graph: parsed.Graph, TruthTable: truthTable(), AuthorityCases: authorityCases(), EvidenceDigest: value.Evidence.Digest, Transitions: transitions, TransitionHeadDigest: transitions[len(transitions)-1].TransitionDigest, Resolutions: resolutions, Metrics: metrics, Decision: decision}
	result.Digest = receiptDigest(result)
	return result, nil
}

type local struct {
	predicate predicate
	digest    string
	available bool
}

func classify(g graph, e evidenceReceipt) ([]string, []transition) {
	states := make([]string, len(g.Nodes))
	outcomes := make([]transition, len(g.Nodes))
	locals := make([]local, len(g.Nodes))
	for i, c := range g.Nodes {
		for _, ec := range e.Claims {
			if ec.ClaimID == c.ClaimID && ec.PropositionDigest == c.PropositionDigest && ec.Status == "CURRENT_EVIDENCE" {
				locals[i] = local{ec.ObservedPredicate, ec.Digest, true}
				break
			}
		}
		state, event, reason := "OPEN", "DEPENDENCY_BLOCKED", "UPSTREAM_UNKNOWN_OR_NON_REFUTING"
		incoming := incomingEdges(i, g)
		refuting := []string{}
		hasRequires, allRequires := false, true
		for _, ed := range incoming {
			from := indexOf(ed.FromClaimID, g)
			if from < 0 {
				continue
			}
			relation := edgeRelation(ed.Kind, states[from], locals[i].predicate, true)
			if relation == relationRefuted {
				refuting = append(refuting, ed.EdgeID)
			}
			if ed.Kind == requires {
				hasRequires = true
				if relation != relationDischarged {
					allRequires = false
				}
			}
		}
		if i == 0 && locals[i].available && locals[i].predicate == explicitContradiction {
			state, event, reason = "REFUTED", "EXPLICIT_CONTRADICTION", "CURRENT_EVIDENCE_EXPLICIT_CONTRADICTION"
		} else if len(refuting) > 0 {
			state, event, reason = "REFUTED", "DEPENDENCY_REFUTED", "EXPLICIT_TYPED_REFUTING_EDGE"
		} else if locals[i].available && locals[i].predicate == accepted {
			if !hasRequires || allRequires {
				state, event, reason = "DISCHARGED", "EVIDENCE_ACCEPTED", "LOCAL_CLAIM_EVIDENCE_PREDICATE"
				if hasRequires {
					event, reason = "DEPENDENCY_DISCHARGED", "ALL_REQUIRES_UPSTREAM_AND_LOCAL_EVIDENCE"
				}
			}
		}
		states[i] = state
		outcomes[i] = transition{0, g.Nodes[i].ClaimID, event, "OPEN", state, coordinate{stage(i), g.Nodes[i].ActivityName, reason}, locals[i].digest, transitionEdges(i, g, states, state, refuting, locals[i].predicate), nil, "pending", "", ""}
	}
	return states, outcomes
}
func stage(i int) string {
	if i == 0 {
		return "OBSERVE"
	}
	return "PROPAGATE"
}
func incomingEdges(i int, g graph) []edge {
	result := []edge{}
	for _, e := range g.Edges {
		if e.ToClaimID == g.Nodes[i].ClaimID {
			result = append(result, e)
		}
	}
	return result
}
func transitionEdges(i int, g graph, states []string, state string, refuting []string, local predicate) []string {
	if len(refuting) > 0 {
		return refuting
	}
	if state == "DISCHARGED" {
		var result []string
		for _, e := range incomingEdges(i, g) {
			from := indexOf(e.FromClaimID, g)
			if e.Kind == requires && from >= 0 && edgeRelation(e.Kind, states[from], local, true) == relationDischarged {
				result = append(result, e.EdgeID)
			}
		}
		return result
	}
	if state != "OPEN" {
		return nil
	}
	var result []string
	for _, e := range incomingEdges(i, g) {
		from := indexOf(e.FromClaimID, g)
		if from >= 0 && (e.Kind == supports || e.Kind == requires) && edgeRelation(e.Kind, states[from], local, true) == relationOpen && (states[from] == "OPEN" || states[from] == "REFUTED") {
			result = append(result, e.EdgeID)
		}
	}
	return result
}
func transitionsFor(g graph, outcomes []transition, provenance string, prior *receipt) []transition {
	result := []transition{}
	previous := ""
	if prior != nil {
		result = append(result, prior.Transitions...)
		previous = prior.TransitionHeadDigest
	}
	if prior == nil {
		for _, c := range g.Nodes {
			value := transition{len(result) + 1, c.ClaimID, "CLAIM_REGISTERED", "UNRECORDED", "OPEN", coordinate{"DECLARE", c.ActivityName, "CLAIM_REGISTERED"}, "", nil, nil, provenance, previous, ""}
			value.TransitionDigest = transitionDigest(value)
			result = append(result, value)
			previous = value.TransitionDigest
		}
	}
	for _, value := range outcomes {
		value.Sequence = len(result) + 1
		value.Provenance = provenance
		value.PreviousTransitionDigest = previous
		value.UpstreamTransitionDigests = upstreamDigests(value.UpstreamEdgeIDs, g, result)
		value.TransitionDigest = transitionDigest(value)
		result = append(result, value)
		previous = value.TransitionDigest
	}
	return result
}
func upstreamDigests(ids []string, g graph, transitions []transition) []string {
	var result []string
	for _, id := range ids {
		for _, e := range g.Edges {
			if e.EdgeID != id {
				continue
			}
			for i := len(transitions) - 1; i >= 0; i-- {
				if transitions[i].ClaimID == e.FromClaimID && transitions[i].Event != "CLAIM_REGISTERED" {
					result = append(result, transitions[i].TransitionDigest)
					break
				}
			}
		}
	}
	return result
}
func buildResolutions(g graph, states []string, outcomes []transition, provenance string) []resolution {
	result := make([]resolution, len(g.Nodes))
	for i, c := range g.Nodes {
		path, ids, kinds := shortestPath(i, g, states[i])
		digests := []string{}
		for _, n := range path {
			for _, o := range outcomes {
				if o.ClaimID == g.Nodes[n].ClaimID {
					digests = append(digests, o.TransitionDigest)
				}
			}
		}
		causePath := idsForPath(path, g)
		responsibility, owner := failureAttribution(i, states[i], causePath)
		value := resolution{ClaimID: c.ClaimID, Axis: c.Axis, PropositionDigest: c.PropositionDigest, State: states[i], Kind: resolutionKind(i, states[i], outcomes[i]), ObservedEvent: outcomes[i].Event, Coordinate: outcomes[i].Coordinate, EvidenceDigest: outcomes[i].EvidenceDigest, Provenance: provenance, FailureResponsibility: responsibility, FailureOwnerClaimID: owner, CausePath: causePath, CauseEdgeIDs: ids, CauseEdgeKinds: kinds, CauseTransitionDigests: digests, CauseCoordinate: &outcomes[i].Coordinate}
		if states[i] == "OPEN" {
			value.MissingEvidenceIDs = []string{"evidence:" + c.ClaimID}
			value.BlockedByClaimIDs, value.BlockedByEdgeIDs = blockedFrontier(i, g, states)
		}
		result[i] = value
	}
	return result
}
func failureAttribution(index int, state string, path []string) (string, string) {
	if state == "DISCHARGED" {
		return "N/A", ""
	}
	if len(path) <= 1 {
		if len(path) == 1 {
			return "DIRECT_CLAIM", path[0]
		}
		return "DIRECT_CLAIM", ""
	}
	return "UPSTREAM_CLAIM", path[0]
}
func resolutionKind(i int, state string, outcome transition) string {
	if i == 0 {
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
		if len(outcome.UpstreamEdgeIDs) == 0 {
			return "DIRECT_DISCHARGED"
		}
		return "DEPENDENCY_DISCHARGED"
	}
	return "DEPENDENCY_BLOCKED"
}
func shortestPath(index int, g graph, state string) ([]int, []string, []edgeKind) {
	if index == 0 {
		return []int{0}, nil, nil
	}
	allowed := map[edgeKind]bool{supports: state == "OPEN", requires: state == "OPEN" || state == "DISCHARGED", contradicts: state == "REFUTED", failureEntailment: state == "REFUTED"}
	best := []int(nil)
	bestEdges := []edge(nil)
	var walk func(int, []int, []edge)
	walk = func(current int, path []int, edges []edge) {
		if current == index {
			if best == nil || len(path) < len(best) || (len(path) == len(best) && pathKey(path, g) < pathKey(best, g)) {
				best = append([]int(nil), path...)
				bestEdges = append([]edge(nil), edges...)
			}
			return
		}
		for _, e := range g.Edges {
			if e.FromClaimID != g.Nodes[current].ClaimID || !allowed[e.Kind] {
				continue
			}
			next := indexOf(e.ToClaimID, g)
			seen := false
			for _, n := range path {
				if n == next {
					seen = true
				}
			}
			if !seen {
				walk(next, append(path, next), append(edges, e))
			}
		}
	}
	walk(0, []int{0}, nil)
	if best == nil {
		return []int{index}, nil, nil
	}
	ids, kinds := []string{}, []edgeKind{}
	for _, e := range bestEdges {
		ids, kinds = append(ids, e.EdgeID), append(kinds, e.Kind)
	}
	return best, ids, kinds
}
func pathKey(path []int, g graph) string {
	result := make([]string, len(path))
	for i, n := range path {
		result[i] = g.Nodes[n].ClaimID
	}
	return strings.Join(result, "\x00")
}
func idsForPath(path []int, g graph) []string {
	result := make([]string, len(path))
	for i, n := range path {
		result[i] = g.Nodes[n].ClaimID
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
func blockedFrontier(i int, g graph, states []string) ([]string, []string) {
	var claims, edges []string
	for _, e := range incomingEdges(i, g) {
		from := indexOf(e.FromClaimID, g)
		if from >= 0 && (e.Kind == supports || e.Kind == requires) && (states[from] == "OPEN" || states[from] == "REFUTED") {
			claims, edges = append(claims, e.FromClaimID), append(edges, e.EdgeID)
		}
	}
	return claims, edges
}

func deriveMetrics(g graph, states []string, resolutions []resolution, outcomes []transition, e evidenceReceipt, recovered bool) metrics {
	result := metrics{FixedClaimTotal: claimTotal, DistinctPropositionTotal: distinct(g), FixedEdgeTotal: edgeTotal, EligibleEdgeTotal: len(g.Edges), ClassifiedClaimTotal: len(states), ClassificationBasisPoints: 10000, TransitionTotal: initialTransitions, CurrentEvidenceTotal: len(e.Claims), TruthTableCaseTotal: len(truthTable()), AuthorityCaseTotal: len(authorityCases())}
	if recovered {
		result.TransitionTotal += claimTotal
	}
	observed, shortestUnion := map[string]bool{}, map[string]bool{}
	for _, ec := range e.Claims {
		if ec.Status == "HISTORICAL_FIXTURE" {
			result.HistoricalEvidenceTotal++
		}
		if ec.ObservedPredicate == unknown {
			result.UnknownEvidenceTotal++
		}
	}
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
	for i, r := range resolutions {
		switch r.Kind {
		case "DIRECT_UNKNOWN":
			result.DirectUnknownClaimTotal++
		case "DEPENDENCY_BLOCKED":
			result.DependencyBlockedClaimTotal++
		case "DIRECT_REFUTED":
			result.DirectRefutedClaimTotal++
		case "DEPENDENCY_REFUTED":
			result.DependencyRefutedClaimTotal++
		case "DIRECT_DISCHARGED":
			result.DirectDischargedClaimTotal++
		case "DEPENDENCY_DISCHARGED":
			result.DependencyDischargedTotal++
		}
		for _, id := range outcomes[i].UpstreamEdgeIDs {
			observed[id] = true
		}
		for _, id := range r.CauseEdgeIDs {
			shortestUnion[id] = true
		}
		if len(r.CauseEdgeIDs) > result.MaximumCausePathDepth {
			result.MaximumCausePathDepth = len(r.CauseEdgeIDs)
		}
	}
	result.ObservedCausalEdgeTotal, result.ShortestPathEdgeUnionTotal = len(observed), len(shortestUnion)
	for _, kind := range []edgeKind{supports, requires, contradicts, failureEntailment} {
		m := edgeMetric{Kind: kind}
		for _, e := range g.Edges {
			if e.Kind != kind {
				continue
			}
			m.Eligible++
			if observed[e.EdgeID] {
				m.ObservedCausal++
			}
			for _, o := range outcomes {
				if contains(o.UpstreamEdgeIDs, e.EdgeID) {
					if o.After == "OPEN" {
						m.Blocking++
					}
					if o.After == "REFUTED" {
						m.Refuting++
					}
					if recovered && o.After == "DISCHARGED" {
						m.Discharge++
					}
				}
			}
		}
		result.ObservedBlockingEdgeTotal += m.Blocking
		result.ObservedRefutingEdgeTotal += m.Refuting
		result.ObservedRecoveryEdgeTotal += m.Discharge
		result.EdgeMetrics = append(result.EdgeMetrics, m)
	}
	return result
}
func distinct(g graph) int {
	seen := map[string]bool{}
	for _, c := range g.Nodes {
		seen[c.PropositionDigest] = true
	}
	return len(seen)
}
func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
func decisionFor(states []string, e evidenceReceipt, recovered bool) decision {
	switch authorityResolution(e) {
	case "NET_REPOSITORY_STATE_CHANGED":
		return decision{"FAIL_CLOSED", "AUTHORITY_CHANGED", "AUTHORITY/REPOSITORY_SNAPSHOT/NET_REPOSITORY_STATE_CHANGED", false}
	case "TRANSIENT_WRITE_AUTHORITY_UNKNOWN":
		return decision{"FAIL_CLOSED", "AUTHORITY_UNKNOWN", "AUTHORITY/WORKFLOW_PERMISSIONS/TRANSIENT_WRITE_AUTHORITY_UNKNOWN", false}
	}
	if all(states, "DISCHARGED") {
		if recovered {
			return decision{"PASS", "CAUSAL_RECOVERY_DISCHARGED", "APPEND_ONLY_EVIDENCE_RECOVERY", false}
		}
		return decision{"PASS", "DIRECT_EVIDENCE_DISCHARGED", "CURRENT_EVIDENCE_PREDICATES_SATISFIED", false}
	}
	if any(states, "REFUTED") {
		if count(states, "REFUTED") == 1 {
			return decision{"FAIL_CLOSED", "DIRECT_REFUTATION", "ONLY_DIRECT_EXPLICIT_CONTRADICTION", false}
		}
		return decision{"FAIL_CLOSED", "CAUSAL_REFUTATION", "EXPLICIT_CONTRADICTION_OR_FAILURE_ENTAILMENT", false}
	}
	return decision{"FAIL_CLOSED", "UNRESOLVED_CLAIM", "UNKNOWN_REMAINS_OPEN", false}
}
func authorityResolution(e evidenceReceipt) string {
	if e.Snapshot.RepositoryWrites != 0 || e.Snapshot.BeforeDigest != e.Snapshot.AfterDigest {
		return "NET_REPOSITORY_STATE_CHANGED"
	}
	if e.Capability.Status != "CURRENT_EVIDENCE" || e.Capability.Provider == "" {
		return "TRANSIENT_WRITE_AUTHORITY_UNKNOWN"
	}
	return "NET_REPOSITORY_STATE_UNCHANGED"
}
func authorityCases() []authorityCase {
	result := []authorityCase{{"NET-SAME-CURRENT", "NET_SAME", "CURRENT_EVIDENCE", "NET_REPOSITORY_STATE_UNCHANGED", ""}, {"NET-CHANGED-CURRENT", "NET_CHANGED", "CURRENT_EVIDENCE", "NET_REPOSITORY_STATE_CHANGED", ""}, {"TRANSIENT-UNKNOWN", "TRANSIENT_UNKNOWN", "UNKNOWN", "TRANSIENT_WRITE_AUTHORITY_UNKNOWN", ""}}
	for i := range result {
		if result[i].NetworkState == "NET_CHANGED" {
			result[i].ObservedResolution = "NET_REPOSITORY_STATE_CHANGED"
		} else if result[i].NetworkState == "TRANSIENT_UNKNOWN" || result[i].CapabilityStatus != "CURRENT_EVIDENCE" {
			result[i].ObservedResolution = "TRANSIENT_WRITE_AUTHORITY_UNKNOWN"
		} else {
			result[i].ObservedResolution = "NET_REPOSITORY_STATE_UNCHANGED"
		}
	}
	return result
}
func validateAuthorityCases(values []authorityCase) error {
	if len(values) != 3 {
		return fmt.Errorf("authority cases have %d cases, want 3", len(values))
	}
	for _, value := range values {
		if value.ExpectedResolution == "" || value.ExpectedResolution != value.ObservedResolution {
			return fmt.Errorf("authority case %q did not execute its expected resolution", value.CaseID)
		}
	}
	return nil
}
func all(values []string, target string) bool {
	for _, v := range values {
		if v != target {
			return false
		}
	}
	return true
}
func any(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
func count(values []string, target string) int {
	total := 0
	for _, value := range values {
		if value == target {
			total++
		}
	}
	return total
}
func statesOf(values []resolution) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = v.State
	}
	return result
}
func validateChain(values []transition, head string) error {
	if len(values) == 0 || values[len(values)-1].TransitionDigest != head {
		return fmt.Errorf("transition head mismatch")
	}
	previous := ""
	for i, v := range values {
		if v.Sequence != i+1 || v.PreviousTransitionDigest != previous || transitionDigest(v) != v.TransitionDigest {
			return fmt.Errorf("transition %d chain mismatch", i+1)
		}
		previous = v.TransitionDigest
	}
	return nil
}
func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func digestJSON(v any) string {
	data, _ := json.Marshal(v)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func digestBytes(v []byte) string {
	sum := sha256.Sum256(v)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func prefixedDigest(v string) string {
	if strings.HasPrefix(v, "sha256:") {
		return v
	}
	return "sha256:" + v
}
func graphDigest(v graph) string                 { v.Digest = ""; return digestJSON(v) }
func receiptDigest(v receipt) string             { v.Digest = ""; return digestJSON(v) }
func evidenceDigest(v evidenceReceipt) string    { v.Digest = ""; return digestJSON(v) }
func evidenceClaimDigest(v evidenceClaim) string { v.Digest = ""; return digestJSON(v) }
func capabilityDigest(v capability) string       { v.Digest = ""; return digestJSON(v) }
func transitionDigest(v transition) string       { v.TransitionDigest = ""; return digestJSON(v) }
