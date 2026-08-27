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

	activities := make(map[string]semantic.Node)
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity {
			activities[node.Name] = node
		}
	}
	claims := make([]Claim, 0, ClaimTotal)
	activityIndex := make(map[string]int)
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		node, ok := activities[activity.Name]
		if !ok || node.ValueProgram == "" {
			return sourceGraph{}, fmt.Errorf("activity %q is not a semantic value claim", activity.Name)
		}
		ordinal := len(claims) + 1
		activityIndex[node.ID.String()] = ordinal - 1
		claims = append(claims, Claim{
			Ordinal: ordinal, Axis: strings.ToLower(activity.Name), ClaimID: node.ID.String(),
			ActivityID: node.ID.String(), ActivityName: activity.Name,
			Statement:    fmt.Sprintf("activity %s declares value claim %q", activity.Name, node.ValueProgram),
			ValueProgram: node.ValueProgram, Producer: ProducerID, Consumer: ConsumerID,
			MetaOperation: MetaOperationID, ProofChoice: ProofChoice,
			Coordinate: Coordinate{Stage: "CLAIM", Step: activity.Name, Reason: "SEMANTIC_ACTIVITY_VALUE"},
		})
	}
	if len(claims) != ClaimTotal {
		return sourceGraph{}, fmt.Errorf("source must contain exactly %d activity claims, got %d", ClaimTotal, len(claims))
	}

	generatedBy := make(map[string]string)
	usedBy := make(map[string][]string)
	for _, fact := range ir.Graph.AllFacts() {
		switch fact.Predicate {
		case semantic.WasGeneratedBy:
			generatedBy[fact.Subject.String()] = fact.Object.String()
		case semantic.Used:
			usedBy[fact.Subject.String()] = append(usedBy[fact.Subject.String()], fact.Object.String())
		}
	}
	type edgeCandidate struct {
		from int
		to   int
		kind EdgeKind
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
	edges := make([]Edge, 0, len(candidates))
	for index, candidate := range candidates {
		edges = append(edges, Edge{
			EdgeID: fmt.Sprintf("E%02d", index+1), FromClaimID: claims[candidate.from].ClaimID,
			ToClaimID: claims[candidate.to].ClaimID, Kind: candidate.kind,
			SemanticBasis: "prov:wasGeneratedBy + prov:used + activity.value-program",
		})
	}
	graph := Graph{
		Schema: GraphSchema, Authority: "CANONICAL_IR_FROM_SYNTAX_PARSE_AND_BIDIR_LOWER",
		Completeness: "CLOSED_WORLD_SOURCE_RECONSTRUCTED", CanonicalIRDigest: prefixedDigest(ir.StableHash()),
		NodeTotal: ClaimTotal, EdgeTotal: EdgeTotal, Nodes: claims, Edges: edges,
	}
	graph.Digest, err = graphDigest(graph)
	if err != nil {
		return sourceGraph{}, err
	}
	return sourceGraph{IR: ir, Graph: graph, RootProgram: claims[0].ValueProgram}, nil
}

func edgeKind(program string) (EdgeKind, bool) {
	const prefix = "claim.edge:"
	if !strings.HasPrefix(program, prefix) {
		return "", false
	}
	switch strings.TrimPrefix(program, prefix) {
	case "supports":
		return Supports, true
	case "requires":
		return Requires, true
	case "contradicts":
		return Contradicts, true
	case "failure-entailment":
		return FailureEntailment, true
	default:
		return "", false
	}
}

func graphDigest(graph Graph) (string, error) {
	graph.Digest = ""
	return digestJSON(graph)
}
func receiptDigest(receipt Receipt) (string, error) {
	receipt.Digest = ""
	return digestJSON(receipt)
}
func observationDigest(observation Observation) (string, error) {
	observation.Digest = ""
	return digestJSON(observation)
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
