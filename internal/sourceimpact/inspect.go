package sourceimpact

import (
	"context"
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func inspect(ctx context.Context, filename string, source []byte, revision string) (snapshot, error) {
	file, ir, err := parseImpactSource(ctx, filename, source, revision)
	if err != nil {
		return snapshot{}, err
	}
	s, names := declarationSnapshot(ir)
	arities := sourceInputArities(file)
	bindSourceScope(&s, file, ir, revision)
	addStructuralEdges(&s, ir)
	for _, node := range ir.Graph.SortedNodes() {
		if err := ctx.Err(); err != nil {
			return s, err
		}
		if node.Kind == semantic.Activity {
			inspectActivityReferences(&s, node, names, arities[node.Name], revision)
		}
	}
	err = finishDependencies(&s, revision)
	return s, err
}

func parseImpactSource(ctx context.Context, filename string, source []byte, revision string) (*syntax.File, semantic.IR, error) {
	if err := ctx.Err(); err != nil {
		return nil, semantic.IR{}, err
	}
	if len(source) > maxSourceBytes {
		return nil, semantic.IR{}, fmt.Errorf("%s source exceeds 2 MiB", revision)
	}
	file, diagnostics := bodycodegen.ParseBodyFile(filename, source)
	if diagnostics.HasErrors() || file == nil {
		return nil, semantic.IR{}, fmt.Errorf("%s source: %s", revision, diagnostics.Error())
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(ctx, file, syntax.EntityFieldsSupportForFile(file, bidir.EntityFieldsV4Support()))
	if err != nil {
		return nil, semantic.IR{}, fmt.Errorf("%s source: %w", revision, err)
	}
	if len(ir.Graph.Nodes()) > maxDeclarations {
		return nil, semantic.IR{}, fmt.Errorf("%s source exceeds %d declarations", revision, maxDeclarations)
	}
	return file, ir, nil
}

func declarationSnapshot(ir semantic.IR) (snapshot, map[string]string) {
	s := snapshot{scope: map[string]string{}, nodes: map[string]semantic.Node{}, facets: map[string]map[string]string{}, edges: []Edge{}, unknowns: []Unknown{}}
	names := map[string]string{}
	for _, node := range ir.Graph.SortedNodes() {
		id := string(node.ID)
		s.nodes[id] = node
		if node.Kind == semantic.Activity {
			names[node.Name] = id
		}
		s.facets[id] = nodeFacets(node)
	}
	return s, names
}

func sourceInputArities(file *syntax.File) map[string]int {
	arities := map[string]int{}
	for _, declaration := range file.Decls {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok {
			arities[activity.Name] = len(activity.Inputs)
		}
	}
	return arities
}

func bindSourceScope(s *snapshot, file *syntax.File, ir semantic.IR, revision string) {
	s.scope["package"], s.scope["namespace"] = ir.Package, string(ir.Namespace)
	imports := []string{}
	for _, imported := range file.Imports {
		imports = append(imports, encoded([]string{imported.Alias, imported.Path}))
		s.addUnknown(revision, "", "import", imported.Path, "imported package dependencies require workspace resolution")
	}
	slices.Sort(imports)
	s.scope["imports"] = encoded(imports)
	for id := range s.nodes {
		s.facets[id]["source_scope"] = encoded([]string{ir.Package, string(ir.Namespace)})
	}
	for _, policy := range ir.Policies {
		key := "policy:" + string(policy.ID)
		s.scope[key] = encoded([]string{policy.Name, policy.SemanticCanonical()})
		s.addUnknown(revision, "", key, "", "policy consumer dependencies require the policy evaluator")
	}
}

func finishDependencies(s *snapshot, revision string) error {
	slices.SortFunc(s.edges, func(a, b Edge) int { return compareStrings(encoded(a), encoded(b)) })
	s.edges = slices.Compact(s.edges)
	if len(s.edges) > maxEdges {
		return fmt.Errorf("%s source exceeds %d dependency edges", revision, maxEdges)
	}
	byOwner := map[string][]Edge{}
	for _, e := range s.edges {
		byOwner[e.Dependent] = append(byOwner[e.Dependent], e)
	}
	for id := range s.nodes {
		s.facets[id]["dependencies"] = encoded(byOwner[id])
	}
	return nil
}

func nodeFacets(n semantic.Node) map[string]string {
	fields := make([]any, 0, len(n.Fields))
	for _, f := range n.Fields {
		fields = append(fields, []any{f.ID, f.Parent, f.Name, f.Aliases, f.TypeRef, f.Presence, f.Cardinality})
	}
	return map[string]string{"identity_role": encoded([]any{n.Kind, n.Namespace}), "presentation": encoded([]any{n.Name, n.Aliases}),
		"fields": encoded(fields), "body": n.ValueProgram, "input_sequence": encoded(n.InputSequence), "assembly": encoded(n.Assembly)}
}

func compareStrings(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
