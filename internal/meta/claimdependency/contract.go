package claimdependency

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type sourceGraph struct {
	IR          semantic.IR
	Graph       Graph
	RootProgram string
}

// GraphFromSource reconstructs executable propositions and typed dependency
// edges from syntax.ParseFile -> bidir.Lower canonical IR. Case names do not
// participate in the graph or in the final state decision.
func GraphFromSource(source []byte, sourcePath string) (Graph, error) {
	parsed, err := graphFromSource(source, sourcePath)
	if err != nil {
		return Graph{}, err
	}
	return parsed.Graph, nil
}

func graphFromSource(source []byte, sourcePath string) (sourceGraph, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return sourceGraph{}, fmt.Errorf("source parse failed: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return sourceGraph{}, fmt.Errorf("source lower failed: %w", err)
	}
	if err := ir.Validate(); err != nil {
		return sourceGraph{}, fmt.Errorf("canonical IR invalid: %w", err)
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
	claims, activityIndex := make([]Claim, 0, ClaimTotal), map[string]int{}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		node, ok := activities[activity.Name]
		if !ok || node.ValueProgram == "" {
			return sourceGraph{}, fmt.Errorf("activity %q is not a semantic value claim", activity.Name)
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
		proposition := normalizedProposition(node.ID.String(), inputs, output, artifact, node.ValueProgram)
		activityIndex[node.ID.String()] = len(claims)
		claims = append(claims, Claim{Ordinal: len(claims) + 1, Axis: strings.ToLower(activity.Name), ClaimID: node.ID.String(), ActivityID: node.ID.String(), ActivityName: activity.Name, Proposition: proposition, PropositionDigest: digestBytes([]byte(proposition)), Target: TargetAddress{Inputs: inputs, Output: output, Artifact: artifact}, ValueProgram: node.ValueProgram, Producer: ProducerID, Consumer: ConsumerID, MetaOperation: MetaOperationID, ProofChoice: ProofChoice, Coordinate: Coordinate{Stage: "CLAIM", Step: activity.Name, Reason: "NORMALIZED_EXECUTION_PROPOSITION"}})
	}
	if len(claims) != ClaimTotal {
		return sourceGraph{}, fmt.Errorf("source must contain exactly %d activity claims, got %d", ClaimTotal, len(claims))
	}
	digests := map[string]bool{}
	for _, claim := range claims {
		digests[claim.PropositionDigest] = true
	}
	if len(digests) != ClaimTotal {
		return sourceGraph{}, fmt.Errorf("claims do not contain %d distinct observed propositions", ClaimTotal)
	}
	type edgeCandidate struct {
		from, to int
		kind     EdgeKind
	}
	candidates := make([]edgeCandidate, 0, EdgeTotal)
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
				return sourceGraph{}, fmt.Errorf("activity %q does not declare a typed dependency edge", claims[to].ActivityName)
			}
			candidates = append(candidates, edgeCandidate{from: from, to: to, kind: kind})
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
	if len(candidates) != EdgeTotal {
		return sourceGraph{}, fmt.Errorf("semantic activity relations must reconstruct exactly %d edges, got %d", EdgeTotal, len(candidates))
	}
	edges := make([]Edge, len(candidates))
	for i, c := range candidates {
		edges[i] = Edge{EdgeID: fmt.Sprintf("E%02d", i+1), FromClaimID: claims[c.from].ClaimID, ToClaimID: claims[c.to].ClaimID, Kind: c.kind, SemanticBasis: "prov:wasGeneratedBy + prov:used + source-derived value-program edge predicate"}
	}
	graph := Graph{Schema: GraphSchema, Authority: "CANONICAL_IR_FROM_SYNTAX_PARSE_AND_BIDIR_LOWER", Completeness: "CLOSED_WORLD_SOURCE_RECONSTRUCTED", CanonicalIRDigest: prefixedDigest(ir.StableHash()), NodeTotal: len(claims), EdgeTotal: len(edges), Nodes: claims, Edges: edges}
	graph.Digest, err = graphDigest(graph)
	if err != nil {
		return sourceGraph{}, err
	}
	return sourceGraph{IR: ir, Graph: graph, RootProgram: claims[0].ValueProgram}, nil
}

func normalizedProposition(activityID string, inputs []string, output, artifact, value string) string {
	return fmt.Sprintf("execute(activity=%s,inputs=[%s],output=%s,artifact=%s,value=%s)", activityID, strings.Join(inputs, ","), output, artifact, value)
}

func edgeKind(program string) (EdgeKind, bool) {
	const prefix = "claim.edge:"
	if !strings.HasPrefix(program, prefix) {
		return "", false
	}
	value := strings.TrimPrefix(program, prefix)
	if pipe := strings.IndexByte(value, '|'); pipe >= 0 {
		value = value[:pipe]
	}
	switch value {
	case "supports":
		return Supports, true
	case "requires":
		return Requires, true
	case "contradicts":
		return Contradicts, true
	case "failure-entailment":
		return FailureEntailment, true
	}
	return "", false
}

func TruthTableCases() []TruthTableCase {
	return []TruthTableCase{
		{TruthTableSchema, "SUPPORTS-POSITIVE", Supports, "established supports target", "DISCHARGED", "EVIDENCE_ACCEPTED", "OPEN", true, "support never discharges by itself"},
		{TruthTableSchema, "SUPPORTS-NEGATIVE", Supports, "refuted supports target", "REFUTED", "UNKNOWN", "OPEN", false, "support does not refute"},
		{TruthTableSchema, "REQUIRES-POSITIVE", Requires, "required proposition established", "DISCHARGED", "EVIDENCE_ACCEPTED", "DISCHARGED", true, "upstream and local requirement hold"},
		{TruthTableSchema, "REQUIRES-NEGATIVE", Requires, "required proposition established", "DISCHARGED", "UNKNOWN", "OPEN", false, "local requirement missing"},
		{TruthTableSchema, "CONTRADICTS-POSITIVE", Contradicts, "established contradiction of target", "REFUTED", "EXPLICIT_CONTRADICTION", "REFUTED", true, "direction is from established contradiction to target"},
		{TruthTableSchema, "CONTRADICTS-NEGATIVE", Contradicts, "ordinary support direction", "REFUTED", "UNKNOWN", "OPEN", false, "name alone cannot refute"},
		{TruthTableSchema, "FAILURE_ENTAILMENT-POSITIVE", FailureEntailment, "failure entails target failure", "REFUTED", "EXPLICIT_CONTRADICTION", "REFUTED", true, "failure entailment is directional"},
		{TruthTableSchema, "FAILURE_ENTAILMENT-NEGATIVE", FailureEntailment, "success or ordinary dependency", "REFUTED", "UNKNOWN", "OPEN", false, "failure evidence is absent"},
	}
}

// validateTruthTable executes the closed edge algebra over each positive and
// negative case before the table is emitted. The table is therefore an
// executable counterexample set, not a list of unchecked labels.
func validateTruthTable(cases []TruthTableCase) error {
	if len(cases) != 2*len(EdgeKinds()) {
		return fmt.Errorf("truth table has %d cases, want %d", len(cases), 2*len(EdgeKinds()))
	}
	seen := map[EdgeKind]int{}
	for _, test := range cases {
		actual := "OPEN"
		switch test.Kind {
		case Requires:
			if test.UpstreamState == "DISCHARGED" && test.LocalPredicate == string(ObservationEvidence) {
				actual = "DISCHARGED"
			}
		case Contradicts, FailureEntailment:
			if test.UpstreamState == "REFUTED" && test.LocalPredicate == string(ObservationContradiction) {
				actual = "REFUTED"
			}
		case Supports:
			// SUPPORTS is never a discharge or refutation entailment.
		default:
			return fmt.Errorf("truth table contains unknown edge kind %q", test.Kind)
		}
		if actual != test.ExpectedState {
			return fmt.Errorf("truth table case %q computed %s, expected %s", test.CaseID, actual, test.ExpectedState)
		}
		seen[test.Kind]++
	}
	for _, kind := range EdgeKinds() {
		if seen[kind] != 2 {
			return fmt.Errorf("truth table edge kind %s has %d cases", kind, seen[kind])
		}
	}
	return nil
}

func graphDigest(graph Graph) (string, error)       { graph.Digest = ""; return digestJSON(graph) }
func receiptDigest(receipt Receipt) (string, error) { receipt.Digest = ""; return digestJSON(receipt) }
func evidenceReceiptDigest(receipt EvidenceReceipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
}
func claimEvidenceDigest(value EvidenceClaim) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
func capabilityDigest(value CapabilityEvidence) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
func transitionDigest(transition Transition) (string, error) {
	transition.TransitionDigest = ""
	return digestJSON(transition)
}
func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func prefixedDigest(raw string) string {
	if strings.HasPrefix(raw, "sha256:") {
		return raw
	}
	return "sha256:" + raw
}
func EdgeKinds() []EdgeKind { return []EdgeKind{Supports, Requires, Contradicts, FailureEntailment} }
