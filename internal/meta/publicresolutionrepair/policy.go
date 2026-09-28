package publicresolutionrepair

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	PolicySchema                 = "gooo/public-semantic-resolution-repair-policy/v1"
	ReportSchema                 = "gooo/public-semantic-resolution-repair-report/v1"
	EvaluatorSchema              = "gooo/public-semantic-resolution-repair-evaluator/v1"
	PolicyName                   = "COUNTEREXAMPLE_DRIVEN_SEMANTIC_RESOLUTION_REPAIR"
	DecisionClosed               = "CLOSED"
	DecisionUnknown              = "UNKNOWN"
	DecisionRefuted              = "REFUTED"
	ResolutionSelective          = "PARTITION_SELECTIVE"
	ResolutionFallback           = "FULL_PROJECT_FALLBACK"
	ResolutionOverlay            = "AUTHORIZED_GRAPH_OVERLAY"
	ProofFoundation              = "FOUNDATION"
	ProofCoherence               = "COHERENCE"
	ProofRegression              = "REGRESSION"
	AuthorizationAuthorized      = "AUTHORIZED"
	AuthorizationRejected        = "REJECTED"
	RepairProposalCount          = 1
	AuthorizationDecisionCount   = 2
	ResolutionLevelCount         = 3
	ProofModeCount               = 3
	ProofObservationCount        = 6
	GraphEdgeCountBefore         = 1
	GraphEdgeCountAfter          = 2
	CanonicalGraphEdgeCount      = 2
	TestUnitCount                = 2
	FallbackExecuted             = 2
	FallbackReused               = 0
	OverlayExecuted              = 2
	OverlayReused                = 0
	SelectivityExecuted          = 1
	SelectivityReused            = 1
	ContinuityEdgeCount          = 2
	GeneratedArtifactCount       = 2
	EvidenceArtifactCount        = 30
	CaseCount                    = 6
	ProofFoundationCount         = 1
	ProofCoherenceCount          = 1
	ProofRegressionCount         = 4
	TransitionCount              = 4
	OriginalCounterexampleCaseID = "CHANGED_HIDDEN_DEPENDENCY"
)

const (
	UnknownClass  = "INCOMPLETE_OR_AMBIGUOUS_REPAIR_EVIDENCE"
	UnknownNext   = "PROVIDE_COMPLETE_REGRESSION_EVIDENCE_OR_AUTHORIZATION"
	RefutedReason = "FAIL_CLOSED_REPAIR_CONTRADICTION"
)

type ResolutionLevel struct {
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

type Partition struct {
	ID       string   `json:"id"`
	Activity string   `json:"activity"`
	TestName string   `json:"test_name"`
	Symbols  []string `json:"symbols"`
	Roots    []string `json:"roots"`
}

type Proposal struct {
	From      string `json:"from"`
	To        string `json:"to"`
	ProofMode string `json:"proof_mode"`
	Trigger   string `json:"trigger_case"`
	Reason    string `json:"reason"`
	Method    string `json:"method"`
	Digest    string `json:"digest"`
}

type Authorization struct {
	Decision string `json:"decision"`
	Method   string `json:"method"`
}

type Eligibility struct {
	ProofMode string `json:"proof_mode"`
	Evidence  string `json:"evidence"`
	Outcome   string `json:"outcome"`
}

type Transition struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Proof  string `json:"proof_mode"`
	Reason string `json:"reason"`
}

type Case struct {
	ID             string `json:"id"`
	Decision       string `json:"decision"`
	Changed        string `json:"changed"`
	Variant        string `json:"variant"`
	ResolutionFrom string `json:"resolution_from"`
	ResolutionTo   string `json:"resolution_to"`
	ProofMode      string `json:"proof_mode"`
	RepairVariant  string `json:"repair_variant"`
	Authorization  string `json:"authorization"`
}

type Policy struct {
	Schema                       string            `json:"schema"`
	EvaluatorSchema              string            `json:"evaluator_schema"`
	SourceDigest                 string            `json:"source_digest"`
	SemanticDigest               string            `json:"semantic_digest"`
	EvaluatorDigest              string            `json:"evaluator_digest"`
	Package                      string            `json:"package"`
	Namespace                    string            `json:"namespace"`
	Activity                     string            `json:"activity"`
	Name                         string            `json:"name"`
	Partitions                   []Partition       `json:"partitions"`
	ResolutionLevels             []ResolutionLevel `json:"resolution_levels"`
	ProofModes                   []string          `json:"proof_modes"`
	ProofModeObservationCount    int               `json:"proof_mode_observation_count"`
	ProofFoundationCount         int               `json:"proof_foundation_count"`
	ProofCoherenceCount          int               `json:"proof_coherence_count"`
	ProofRegressionCount         int               `json:"proof_regression_count"`
	Proposals                    []Proposal        `json:"proposals"`
	Authorizations               []Authorization   `json:"authorizations"`
	Eligibility                  []Eligibility     `json:"eligibility"`
	Transitions                  []Transition      `json:"transitions"`
	CanonicalEdges               []Edge            `json:"canonical_edges"`
	GraphEdgeCountBefore         int               `json:"graph_edge_count_before"`
	GraphEdgeCountAfter          int               `json:"graph_edge_count_after"`
	CanonicalGraphEdgeCount      int               `json:"canonical_graph_edge_count"`
	TestUnitCount                int               `json:"test_unit_count"`
	FallbackTestUnitsExecuted    int               `json:"fallback_test_units_executed"`
	FallbackTestUnitsReused      int               `json:"fallback_test_units_reused"`
	OverlayTestUnitsExecuted     int               `json:"overlay_test_units_executed"`
	OverlayTestUnitsReused       int               `json:"overlay_test_units_reused"`
	SelectivityTestUnitsExecuted int               `json:"selectivity_test_units_executed"`
	SelectivityTestUnitsReused   int               `json:"selectivity_test_units_reused"`
	ContinuityEdgeCount          int               `json:"continuity_edge_count"`
	GeneratedArtifactCount       int               `json:"generated_artifact_count"`
	EvidenceArtifactCount        int               `json:"evidence_artifact_count"`
	RuntimeRule                  string            `json:"runtime_rule"`
	RefutedDominatesUnknown      bool              `json:"refuted_dominates_unknown"`
	Cases                        []Case            `json:"cases"`
	IR                           semantic.IR       `json:"-"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type markers map[string][]string

func Load(filename string, source []byte) (Policy, error) {
	if len(source) == 0 {
		return Policy{}, errors.New("semantic resolution repair source is empty")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if file == nil || diagnostics.HasErrors() {
		return Policy{}, errors.New("semantic resolution repair source has syntax diagnostics")
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return Policy{}, fmt.Errorf("lower semantic resolution repair source: %w", err)
	}
	if ir.Package != "partialreuseexample" || ir.Namespace.String() != "partial_reuse_example" {
		return Policy{}, errors.New("semantic resolution repair source is not the canonical partial-reuse example")
	}
	all := markers{}
	partitions := make([]Partition, 0, TestUnitCount)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind != semantic.Activity || node.ValueProgram == "" {
			continue
		}
		current := parseMarkers(node.ValueProgram)
		for key, values := range current {
			all[key] = append(all[key], values...)
		}
		components := current["partial-reuse-component"]
		if len(components) == 1 && len(current["partial-reuse-test"]) == 1 && len(current["partial-reuse-symbols"]) == 1 && len(current["partial-reuse-roots"]) == 1 {
			partitions = append(partitions, Partition{ID: components[0], Activity: node.Name, TestName: current["partial-reuse-test"][0], Symbols: splitCSV(current["partial-reuse-symbols"][0]), Roots: splitCSV(current["partial-reuse-roots"][0])})
		}
	}
	policy := Policy{
		Schema: PolicySchema, EvaluatorSchema: EvaluatorSchema,
		SourceDigest: cache.HashBytes(source).String(), SemanticDigest: ir.StableHash(), EvaluatorDigest: evaluatorDigest(),
		Package: ir.Package, Namespace: ir.Namespace.String(), Activity: "CreateReceipt", Name: first(all, "partial-repair-policy"), Partitions: partitions,
		ResolutionLevels: parseLevels(all["partial-repair-resolution"]), ProofModes: append([]string(nil), all["partial-repair-proof-mode"]...),
		ProofModeObservationCount: parseInt(first(all, "partial-repair-proof-observation-count")), ProofFoundationCount: parseInt(first(all, "partial-repair-proof-foundation-count")),
		ProofCoherenceCount: parseInt(first(all, "partial-repair-proof-coherence-count")), ProofRegressionCount: parseInt(first(all, "partial-repair-proof-regression-count")),
		Proposals: parseProposals(all["partial-repair-proposal"]), Authorizations: parseAuthorizations(all["partial-repair-authorization"]),
		Eligibility: parseEligibility(all["partial-repair-eligibility"]), Transitions: parseTransitions(all["partial-repair-transition"]),
		CanonicalEdges: parseEdges(all["partial-reuse-edge"]), GraphEdgeCountBefore: parseInt(first(all, "partial-repair-graph-edge-count-before")),
		GraphEdgeCountAfter: parseInt(first(all, "partial-repair-graph-edge-count-after")), CanonicalGraphEdgeCount: parseInt(first(all, "partial-repair-canonical-graph-edge-count")),
		TestUnitCount: parseInt(first(all, "partial-repair-test-unit-count")), FallbackTestUnitsExecuted: parseInt(first(all, "partial-repair-fallback-test-units-executed")),
		FallbackTestUnitsReused: parseInt(first(all, "partial-repair-fallback-test-units-reused")), OverlayTestUnitsExecuted: parseInt(first(all, "partial-repair-overlay-test-units-executed")),
		OverlayTestUnitsReused: parseInt(first(all, "partial-repair-overlay-test-units-reused")), SelectivityTestUnitsExecuted: parseInt(first(all, "partial-repair-selectivity-test-units-executed")),
		SelectivityTestUnitsReused: parseInt(first(all, "partial-repair-selectivity-test-units-reused")), ContinuityEdgeCount: parseInt(first(all, "partial-repair-continuity-edge-count")),
		GeneratedArtifactCount: parseInt(first(all, "partial-repair-generated-artifact-count")), EvidenceArtifactCount: parseInt(first(all, "partial-repair-evidence-artifact-count")),
		RuntimeRule: first(all, "partial-repair-runtime-rule"), RefutedDominatesUnknown: first(all, "partial-repair-refuted-dominates-unknown") == "true",
		Cases: parseCases(all["partial-repair-case"]), IR: ir,
	}
	if err := policy.Validate(); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (policy Policy) Validate() error {
	if policy.Schema != PolicySchema || policy.EvaluatorSchema != EvaluatorSchema || policy.Name != PolicyName || policy.Package != "partialreuseexample" ||
		policy.Namespace != "partial_reuse_example" || len(policy.ResolutionLevels) != ResolutionLevelCount || len(policy.ProofModes) != ProofModeCount ||
		policy.ProofModeObservationCount != ProofObservationCount || policy.ProofFoundationCount != ProofFoundationCount || policy.ProofCoherenceCount != ProofCoherenceCount ||
		policy.ProofRegressionCount != ProofRegressionCount || len(policy.Proposals) != RepairProposalCount || len(policy.Authorizations) != AuthorizationDecisionCount ||
		len(policy.Eligibility) != 1 || len(policy.Transitions) != TransitionCount || len(policy.CanonicalEdges) != CanonicalGraphEdgeCount ||
		policy.GraphEdgeCountBefore != GraphEdgeCountBefore || policy.GraphEdgeCountAfter != GraphEdgeCountAfter || policy.CanonicalGraphEdgeCount != CanonicalGraphEdgeCount ||
		policy.TestUnitCount != TestUnitCount || policy.FallbackTestUnitsExecuted != FallbackExecuted || policy.FallbackTestUnitsReused != FallbackReused ||
		policy.OverlayTestUnitsExecuted != OverlayExecuted || policy.OverlayTestUnitsReused != OverlayReused || policy.SelectivityTestUnitsExecuted != SelectivityExecuted ||
		policy.SelectivityTestUnitsReused != SelectivityReused || policy.ContinuityEdgeCount != ContinuityEdgeCount || policy.GeneratedArtifactCount != GeneratedArtifactCount ||
		policy.EvidenceArtifactCount != EvidenceArtifactCount || policy.RuntimeRule == "" || !policy.RefutedDominatesUnknown || len(policy.Cases) != CaseCount {
		return errors.New("semantic resolution repair policy identity or denominator is invalid")
	}
	if policy.SourceDigest == "" || policy.SemanticDigest == "" || policy.EvaluatorDigest == "" {
		return errors.New("semantic resolution repair policy digest identity is missing")
	}
	for index, level := range policy.ResolutionLevels {
		want := []string{ResolutionSelective, ResolutionFallback, ResolutionOverlay}[index]
		if level.Name != want || level.Rank != index+1 {
			return errors.New("semantic resolution ladder is not canonical")
		}
	}
	if !sameStrings(policy.ProofModes, []string{ProofFoundation, ProofCoherence, ProofRegression}) {
		return errors.New("semantic resolution proof modes are not canonical")
	}
	if len(policy.Proposals) != 1 || policy.Proposals[0].From != "shared-contract" || policy.Proposals[0].To != "inventory" || policy.Proposals[0].ProofMode != ProofRegression || policy.Proposals[0].Trigger != OriginalCounterexampleCaseID || policy.Proposals[0].Method != "DETERMINISTIC_MISSING_EDGE" {
		return errors.New("semantic repair proposal is not canonical")
	}
	seenAuth := map[string]bool{}
	for _, authorization := range policy.Authorizations {
		if authorization.Decision == "" || authorization.Method == "" || seenAuth[authorization.Decision] {
			return errors.New("semantic repair authorization table is invalid")
		}
		seenAuth[authorization.Decision] = true
	}
	if !seenAuth[AuthorizationAuthorized] || !seenAuth[AuthorizationRejected] {
		return errors.New("semantic repair authorization table is incomplete")
	}
	if policy.Eligibility[0].ProofMode != ProofRegression || policy.Eligibility[0].Evidence != "PROVEN_HIDDEN_DEPENDENCY" || policy.Eligibility[0].Outcome != ResolutionOverlay {
		return errors.New("semantic repair eligibility is not canonical")
	}
	wantedTransitions := map[string]bool{
		ResolutionSelective + ">" + ResolutionFallback + "|" + ProofRegression + "|HIDDEN_DEPENDENCY": true,
		ResolutionFallback + ">" + ResolutionOverlay + "|" + ProofRegression + "|AUTHORIZED":          true,
		ResolutionFallback + ">UNKNOWN|" + ProofCoherence + "|AMBIGUOUS_OR_UNSUPPORTED":               true,
		ResolutionFallback + ">REFUTED|" + ProofRegression + "|TAMPERED_OR_UNAUTHORIZED":              true,
	}
	for _, transition := range policy.Transitions {
		key := transition.From + ">" + transition.To + "|" + transition.Proof + "|" + transition.Reason
		if !wantedTransitions[key] {
			return errors.New("semantic repair transition is not canonical")
		}
		delete(wantedTransitions, key)
	}
	if len(wantedTransitions) != 0 {
		return errors.New("semantic repair transition table is incomplete")
	}
	if graphDigest(policy.CanonicalEdges) == "" {
		return errors.New("semantic repair canonical graph is empty")
	}
	if len(policy.Partitions) != TestUnitCount {
		return errors.New("semantic repair partition table is incomplete")
	}
	seenPartitions := map[string]bool{}
	for _, partition := range policy.Partitions {
		if partition.ID == "" || partition.Activity == "" || partition.TestName == "" || len(partition.Symbols) == 0 || len(partition.Roots) == 0 || seenPartitions[partition.ID] {
			return errors.New("semantic repair partition relation is invalid")
		}
		seenPartitions[partition.ID] = true
	}
	if !seenPartitions["orders"] || !seenPartitions["inventory"] {
		return errors.New("semantic repair partitions are not canonical")
	}
	for _, edge := range policy.CanonicalEdges {
		if edge.From == "" || edge.To == "" || edge.From != "shared-contract" || !seenPartitions[edge.To] {
			return errors.New("semantic repair canonical graph edge is invalid")
		}
	}
	counts := map[string]int{}
	seenCases := map[string]bool{}
	for _, item := range policy.Cases {
		if item.ID == "" || seenCases[item.ID] || item.Decision == "" || item.Changed == "" || item.Variant == "" || item.ResolutionFrom == "" || item.ResolutionTo == "" || item.ProofMode == "" || item.RepairVariant == "" || item.Authorization == "" {
			return errors.New("semantic resolution repair case table is invalid")
		}
		seenCases[item.ID] = true
		counts[item.Decision]++
	}
	if counts[DecisionClosed] != 2 || counts[DecisionUnknown] != 2 || counts[DecisionRefuted] != 2 {
		return errors.New("semantic resolution repair cases are not 2/2/2")
	}
	return nil
}

func (policy Policy) Case(id string) (Case, bool) {
	for _, item := range policy.Cases {
		if item.ID == id {
			return item, true
		}
	}
	return Case{}, false
}

func (policy Policy) CanonicalGraphDigest() string { return graphDigest(policy.CanonicalEdges) }

func evaluatorDigest() string {
	return cache.HashBytes([]byte(EvaluatorSchema + "\x00resolution-ladder-proof-transition-v1")).String()
}

func parseMarkers(value string) markers {
	result := markers{}
	for part := range strings.SplitSeq(value, ";") {
		fields := strings.SplitN(part, "=", 2)
		if len(fields) != 2 || strings.TrimSpace(fields[0]) == "" {
			continue
		}
		result[strings.TrimSpace(fields[0])] = append(result[strings.TrimSpace(fields[0])], strings.TrimSpace(fields[1]))
	}
	return result
}

func first(values markers, key string) string {
	if len(values[key]) == 0 {
		return ""
	}
	return values[key][0]
}

func parseInt(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func parseLevels(values []string) []ResolutionLevel {
	result := make([]ResolutionLevel, 0, len(values))
	for _, value := range values {
		fields := strings.Split(value, "|")
		if len(fields) == 2 {
			result = append(result, ResolutionLevel{Name: fields[0], Rank: parseInt(fields[1])})
		}
	}
	return result
}

func parseEdges(values []string) []Edge {
	result := make([]Edge, 0, len(values))
	for _, value := range values {
		fields := strings.SplitN(value, ">", 2)
		if len(fields) == 2 {
			result = append(result, Edge{From: strings.TrimSpace(fields[0]), To: strings.TrimSpace(fields[1])})
		}
	}
	return result
}

func parseProposals(values []string) []Proposal {
	result := make([]Proposal, 0, len(values))
	for _, value := range values {
		fields := strings.Split(value, "|")
		if len(fields) != 4 {
			continue
		}
		edge := parseEdges([]string{fields[0]})
		if len(edge) != 1 {
			continue
		}
		result = append(result, Proposal{From: edge[0].From, To: edge[0].To, ProofMode: fields[1], Trigger: fields[2], Method: fields[3], Reason: "OBSERVED_AFFECTED_TEST_AND_COMPONENT"})
	}
	return result
}

func parseAuthorizations(values []string) []Authorization {
	result := make([]Authorization, 0, len(values))
	for _, value := range values {
		fields := strings.SplitN(value, "|", 2)
		if len(fields) == 2 {
			result = append(result, Authorization{Decision: fields[0], Method: fields[1]})
		}
	}
	return result
}

func parseEligibility(values []string) []Eligibility {
	result := make([]Eligibility, 0, len(values))
	for _, value := range values {
		fields := strings.Split(value, "|")
		if len(fields) == 3 {
			result = append(result, Eligibility{ProofMode: fields[0], Evidence: fields[1], Outcome: fields[2]})
		}
	}
	return result
}

func parseTransitions(values []string) []Transition {
	result := make([]Transition, 0, len(values))
	for _, value := range values {
		fields := strings.Split(value, "|")
		if len(fields) != 3 {
			continue
		}
		states := strings.SplitN(fields[0], ">", 2)
		if len(states) == 2 {
			result = append(result, Transition{From: states[0], To: states[1], Proof: fields[1], Reason: fields[2]})
		}
	}
	return result
}

func parseCases(values []string) []Case {
	result := make([]Case, 0, len(values))
	for _, value := range values {
		fields := strings.Split(value, "|")
		if len(fields) == 9 {
			result = append(result, Case{ID: fields[0], Decision: fields[1], Changed: fields[2], Variant: fields[3], ResolutionFrom: fields[4], ResolutionTo: fields[5], ProofMode: fields[6], RepairVariant: fields[7], Authorization: fields[8]})
		}
	}
	return result
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

func graphDigest(edges []Edge) string {
	copyEdges := append([]Edge(nil), edges...)
	sort.Slice(copyEdges, func(i, j int) bool {
		return copyEdges[i].From+">"+copyEdges[i].To < copyEdges[j].From+">"+copyEdges[j].To
	})
	data := make([]string, 0, len(copyEdges))
	for _, edge := range copyEdges {
		data = append(data, edge.From+">"+edge.To)
	}
	return cache.HashBytes([]byte(strings.Join(data, "\x00"))).String()
}
