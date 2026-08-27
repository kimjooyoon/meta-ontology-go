package semanticdeltareceipt

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type projectedSource struct {
	nodes          []Node
	facts          []Fact
	claims         []Claim
	semanticDigest string
}

func projectSource(filename string, raw []byte) (projectedSource, error) {
	file, diagnostics := syntax.ParseFile(filename, string(raw))
	if diagnostics.Error() != nil || file == nil {
		return projectedSource{}, fmt.Errorf("canonical syntax rejected source: %v", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return projectedSource{}, fmt.Errorf("canonical lowering rejected source: %w", err)
	}
	result := projectedSource{semanticDigest: "sha256:" + ir.StableHash()}
	for _, node := range ir.Graph.Nodes() {
		result.nodes = append(result.nodes, Node{ID: node.ID.String(), Kind: strings.ToUpper(node.Kind.String())})
	}
	for _, fact := range ir.Graph.DeterministicFacts() {
		result.facts = append(result.facts, Fact{Subject: fact.Subject.String(), Predicate: fact.Predicate.String(), Object: fact.Object.String()})
	}
	result.claims = claimsFromFacts(result.facts)
	sort.Slice(result.nodes, func(i, j int) bool { return result.nodes[i].ID < result.nodes[j].ID })
	sort.Slice(result.facts, func(i, j int) bool { return factLess(result.facts[i], result.facts[j]) })
	sort.Slice(result.claims, func(i, j int) bool { return result.claims[i].ID < result.claims[j].ID })
	return result, nil
}

func claimsFromFacts(facts []Fact) []Claim {
	claims := make([]Claim, 0, len(facts))
	for _, fact := range facts {
		subject, predicate, object := fact.Subject, "uses", fact.Object
		if fact.Predicate == semantic.WasGeneratedBy.String() {
			subject, predicate, object = fact.Object, "generates", fact.Subject
		}
		digest := propositionDigest(ClaimKindObject, subject, predicate, object)
		claims = append(claims, Claim{ID: objectClaimID(digest), Kind: ClaimKindObject, Subject: subject, Predicate: predicate, Object: object, Status: StatusOpen, Stage: "semantic-extraction", Step: "bind-canonical-fact", Reason: "CANONICAL_LOWERING_BOUND", PropositionDigest: digest})
	}
	return claims
}

func factLess(left, right Fact) bool {
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	return left.Object < right.Object
}
