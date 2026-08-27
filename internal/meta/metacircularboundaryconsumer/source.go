package metacircularboundaryconsumer

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	contract "github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundarycontract"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// observeSource is a consumer-owned source replay. It shares language
// parsing/lowering infrastructure only; it does not call the producer.
func observeSource(path string, source []byte) (contract.SourceObservation, error) {
	file, diagnostics := syntax.ParseFile(path, string(source))
	if diagnostics.HasErrors() || file == nil || file.Package == nil || file.Namespace == nil {
		return contract.SourceObservation{}, fmt.Errorf("consumer could not parse %s", path)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return contract.SourceObservation{}, fmt.Errorf("consumer could not lower %s: %w", path, err)
	}
	normalized, err := ir.Normalized()
	if err != nil {
		return contract.SourceObservation{}, fmt.Errorf("consumer could not normalize %s: %w", path, err)
	}
	if err := normalized.Validate(); err != nil {
		return contract.SourceObservation{}, fmt.Errorf("consumer could not validate %s: %w", path, err)
	}
	entities := []string{}
	activities := []string{}
	computations := []contract.Computation{}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			entities = append(entities, value.Name)
		case *syntax.ActivityDecl:
			activities = append(activities, value.Name)
			if value.ValueProgramPresent || value.ValueProgram != "" {
				computations = append(computations, contract.Computation{Activity: value.Name, Program: value.ValueProgram})
			}
		}
	}
	semanticDigest := digestBytes([]byte(normalized.SemanticCanonical()))
	return contract.SourceObservation{
		Path: path, SourceDigest: digestBytes(source), SemanticDigest: semanticDigest,
		Package: file.Package.Name, Namespace: file.Namespace.Name, Entities: entities, Activities: activities,
		Computations:     computations,
		Graph:            loweredGraph(normalized),
		DescriptionBound: containsAll(entities, requiredEntities),
	}, nil
}

func loweredGraph(ir semantic.IR) contract.GraphObservation {
	graph := contract.GraphObservation{Schema: graphSchema}
	nodes := make(map[semantic.ID]semantic.Node)
	for _, node := range ir.Graph.Nodes() {
		nodes[node.ID] = node
	}
	inputs := make(map[string]string)
	outputs := make(map[string]string)
	for _, fact := range ir.Graph.Facts() {
		subject, subjectOK := nodes[fact.Subject]
		object, objectOK := nodes[fact.Object]
		if !subjectOK || !objectOK {
			continue
		}
		if fact.Predicate == semantic.Used && subject.Kind == semantic.Activity && object.Kind == semantic.Entity {
			inputs[subject.Name] = object.Name
		}
		if fact.Predicate == semantic.WasGeneratedBy && subject.Kind == semantic.Entity && object.Kind == semantic.Activity {
			outputs[object.Name] = subject.Name
		}
	}
	steps := []struct{ from, through, to, relation string }{
		{from: "DescribeMetaOperation", through: "SelfDescription", to: "GrantReadOnlyMetaCapability", relation: "DESCRIBES_TO_GRANTS"},
		{from: "GrantReadOnlyMetaCapability", through: "ReadOnlyCapability", to: "ExecuteMetaOperation", relation: "GRANTS_TO_EXECUTES"},
	}
	expectedFrom := []string{"MetaOperation", "SelfDescription"}
	expectedThrough := []string{"SelfDescription", "ReadOnlyCapability"}
	graph.Valid = true
	for ordinal, step := range steps {
		fromType, throughType, toType := inputs[step.from], outputs[step.from], inputs[step.to]
		graph.Relations = append(graph.Relations, contract.TypedRelation{Ordinal: ordinal + 1, FromActivity: step.from, FromType: fromType, Relation: step.relation, ThroughType: throughType, ToType: toType, ToActivity: step.to})
		if fromType != expectedFrom[ordinal] || throughType != expectedThrough[ordinal] || toType != expectedThrough[ordinal] {
			graph.Valid = false
		}
	}
	graph.Path = []string{"DescribeMetaOperation", "GrantReadOnlyMetaCapability", "ExecuteMetaOperation"}
	if len(graph.Relations) != len(steps) {
		graph.Valid = false
	}
	if !graph.Valid {
		graph.Reason = reasonGraphUnknown
	}
	graph.Digest = digestValue(struct {
		Relations []contract.TypedRelation `json:"relations"`
		Path      []string                 `json:"path"`
		Valid     bool                     `json:"valid"`
		Reason    string                   `json:"reason"`
	}{Relations: graph.Relations, Path: graph.Path, Valid: graph.Valid, Reason: graph.Reason})
	return graph
}

func containsAll(observed, required []string) bool {
	set := make(map[string]struct{}, len(observed))
	for _, value := range observed {
		set[value] = struct{}{}
	}
	for _, value := range required {
		if _, ok := set[value]; !ok {
			return false
		}
	}
	return true
}
