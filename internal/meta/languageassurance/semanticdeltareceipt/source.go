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
	path               string
	rawDigest          string
	nodes              []Node
	facts              []Fact
	claims             []Claim
	semanticDigest     string
	semanticComponents []SemanticComponent
}

func sourceEnvelope(filename string, raw []byte) projectedSource {
	return projectedSource{path: filename, rawDigest: digestBytes(raw)}
}

func projectSource(filename string, raw []byte) (projectedSource, error) {
	return projectSourceSide(filename, raw, false)
}

func projectSourceSide(filename string, raw []byte, before bool) (projectedSource, error) {
	file, diagnostics := syntax.ParseFile(filename, string(raw))
	if diagnostics.Error() != nil || file == nil {
		return projectedSource{}, fmt.Errorf("canonical syntax rejected source: %v", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return projectedSource{}, fmt.Errorf("canonical lowering rejected source: %w", err)
	}
	result := projectedSource{path: filename, rawDigest: digestBytes(raw), semanticDigest: "sha256:" + ir.StableHash()}
	for _, node := range ir.Graph.Nodes() {
		wireNode := Node{ID: node.ID.String(), Kind: strings.ToUpper(node.Kind.String()), Namespace: node.Namespace.String(), ValueProgram: node.ValueProgram}
		for _, field := range node.Fields {
			wireNode.Fields = append(wireNode.Fields, Field{ID: field.ID.String(), Parent: field.Parent.String(), TypeID: field.TypeRef.ID.String(), Presence: string(field.Presence), Cardinality: string(field.Cardinality)})
		}
		sort.Slice(wireNode.Fields, func(i, j int) bool { return wireNode.Fields[i].ID < wireNode.Fields[j].ID })
		result.nodes = append(result.nodes, wireNode)
		result.semanticComponents = append(result.semanticComponents, component(ComponentNode, wireNode.ID, "semantic-canonical", node.SemanticCanonical()))
		for _, field := range node.Fields {
			result.semanticComponents = append(result.semanticComponents, component(ComponentField, field.ID.String(), "semantic-canonical", field.SemanticCanonical()))
		}
		if node.ValueProgram != "" {
			result.semanticComponents = append(result.semanticComponents, component(ComponentValue, node.ID.String(), "value-program", node.ValueProgram))
		}
	}
	for _, fact := range ir.Graph.DeterministicFacts() {
		result.facts = append(result.facts, Fact{Subject: fact.Subject.String(), Predicate: fact.Predicate.String(), Object: fact.Object.String()})
		result.semanticComponents = append(result.semanticComponents, component(ComponentRelation, fact.Subject.String(), fact.Predicate.String(), fact.SemanticCanonical()))
	}
	result.semanticComponents = append(result.semanticComponents, component(ComponentFingerprint, "source-ir", "stable-hash", ir.SemanticCanonical()))
	result.claims = claimsFromFacts(result.facts, filename, result.rawDigest, result.semanticDigest, before, claimIdentityVersionForRaw(raw))
	sort.Slice(result.nodes, func(i, j int) bool { return result.nodes[i].ID < result.nodes[j].ID })
	sort.Slice(result.facts, func(i, j int) bool { return factLess(result.facts[i], result.facts[j]) })
	sort.Slice(result.claims, func(i, j int) bool { return result.claims[i].ID < result.claims[j].ID })
	sort.Slice(result.semanticComponents, func(i, j int) bool {
		return semanticComponentLess(result.semanticComponents[i], result.semanticComponents[j])
	})
	return result, nil
}

func component(kind, subject, predicate, object string) SemanticComponent {
	normalized := normalizedProposition(kind, subject, predicate, object)
	digest := propositionDigest(kind, subject, predicate, object)
	return SemanticComponent{ID: "gooo://semantic-delta/component/" + digest[len("sha256:"):], Kind: kind, Subject: subject, Predicate: predicate, Object: object, PropositionDigest: digestValue(normalized)}
}

func semanticComponentLess(left, right SemanticComponent) bool {
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	return left.Predicate < right.Predicate
}

func claimsFromFacts(facts []Fact, filename, rawDigest, semanticDigest string, before bool, identityVersion string) []Claim {
	claims := make([]Claim, 0, len(facts))
	for _, fact := range facts {
		subject, predicate, object := fact.Subject, "uses", fact.Object
		if fact.Predicate == semantic.WasGeneratedBy.String() {
			subject, predicate, object = fact.Object, "generates", fact.Subject
		}
		normalized := normalizedProposition(ClaimKindObject, subject, predicate, object)
		proposition := propositionDigest(ClaimKindObject, subject, predicate, object)
		target := canonicalTargetAddress(subject, predicate, object)
		relationRole := predicate + "|observation"
		if before {
			relationRole += "|before"
		} else {
			relationRole += "|after"
		}
		claim := Claim{ID: objectClaimIDWithVersion(target, relationRole, identityVersion), ClaimTypeID: claimTypeID(ClaimKindObject, subject, predicate, object), Kind: ClaimKindObject, Subject: subject, Predicate: predicate, Object: object, Status: StatusOpen, Stage: "semantic-extraction", Step: "bind-canonical-fact", Reason: "CANONICAL_LOWERING_BOUND", NormalizedProposition: normalized, PropositionDigest: proposition, TargetAddress: target, TargetAddressDigest: targetAddressDigest(target), RelationRole: relationRole}
		if before {
			claim.BeforeSourcePath = filename
			claim.BeforeSourceDigest, claim.BeforeSemanticDigest = rawDigest, semanticDigest
		} else {
			claim.AfterSourcePath = filename
			claim.AfterSourceDigest, claim.AfterSemanticDigest = rawDigest, semanticDigest
		}
		claims = append(claims, claim)
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
