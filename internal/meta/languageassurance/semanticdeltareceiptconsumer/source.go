package semanticdeltareceiptconsumer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
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
		return projectedSource{}, fmt.Errorf("consumer syntax rejection: %v", diagnostics.Error())
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return projectedSource{}, fmt.Errorf("consumer syntax adaptation: %w", err)
	}
	ir, err := bidir.LowerDocument(document)
	if err != nil {
		return projectedSource{}, fmt.Errorf("consumer lowering rejection: %w", err)
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

func factLess(left, right Fact) bool {
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	return left.Object < right.Object
}
