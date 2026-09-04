package publicpartialreuse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	goparser "go/parser"
	"go/token"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func hashJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return cache.HashBytes(data).String()
}

func partitionSourceDigest(policy Policy, partition Partition) (string, error) {
	projection, err := partitionProjection(policy, partition)
	if err != nil {
		return "", err
	}
	return hashJSON(projection), nil
}

func partitionSemanticDigest(policy Policy, partition Partition) (string, error) {
	projection, err := partitionProjection(policy, partition)
	if err != nil {
		return "", err
	}
	value, ok := projection.(struct {
		Partition string           `json:"partition"`
		Activity  string           `json:"activity"`
		Test      string           `json:"test"`
		Symbols   []string         `json:"symbols"`
		Roots     []string         `json:"roots"`
		Nodes     []nodeProjection `json:"nodes"`
		Facts     []string         `json:"facts"`
	})
	if !ok {
		return "", errorsNew("partial reuse semantic projection has an unexpected shape")
	}
	for index := range value.Nodes {
		value.Nodes[index].Name = ""
		value.Nodes[index].ValueProgram = ""
	}
	return hashJSON(value), nil
}

func partitionProjection(policy Policy, partition Partition) (any, error) {
	rootSet := map[string]bool{}
	for _, root := range partition.Roots {
		rootSet[root] = true
	}
	nodes := make([]nodeProjection, 0, len(partition.Roots))
	ids := map[semantic.ID]bool{}
	for _, node := range policy.IR.Graph.Nodes() {
		if !rootSet[node.Name] {
			continue
		}
		nodes = append(nodes, nodeProjection{ID: node.ID.String(), Kind: node.Kind.String(), Name: node.Name, ValueProgram: node.ValueProgram, Semantic: node.SemanticCanonical()})
		ids[node.ID] = true
	}
	if len(nodes) != len(rootSet) {
		return nil, fmt.Errorf("partition %q owns %d semantic roots, found %d", partition.ID, len(rootSet), len(nodes))
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	facts := make([]string, 0)
	for _, fact := range policy.IR.Graph.DeterministicFacts() {
		if ids[fact.Subject] || ids[fact.Object] {
			facts = append(facts, fact.SemanticCanonical())
		}
	}
	sort.Strings(facts)
	return struct {
		Partition string           `json:"partition"`
		Activity  string           `json:"activity"`
		Test      string           `json:"test"`
		Symbols   []string         `json:"symbols"`
		Roots     []string         `json:"roots"`
		Nodes     []nodeProjection `json:"nodes"`
		Facts     []string         `json:"facts"`
	}{partition.ID, partition.Activity, partition.TestName, append([]string(nil), partition.Symbols...), append([]string(nil), partition.Roots...), nodes, facts}, nil
}

type nodeProjection struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	ValueProgram string `json:"value_program,omitempty"`
	Semantic     string `json:"semantic"`
}

func actualEdges(policy Policy) []Edge {
	result := make([]Edge, 0, len(policy.Edges))
	componentRoots := map[string]string{}
	for _, component := range policy.Components {
		componentRoots[component.ID] = component.Root
	}
	for _, edge := range policy.Edges {
		root := componentRoots[edge.From]
		partition, ok := partitionByID(policy.Partitions, edge.To)
		if !ok {
			continue
		}
		activity, ok := policy.IR.Graph.NodeByName(policy.IR.Namespace, partition.Activity)
		shared, sharedOK := policy.IR.Graph.NodeByName(policy.IR.Namespace, root)
		if !ok || !sharedOK {
			continue
		}
		for _, fact := range policy.IR.Graph.DeterministicFacts() {
			if fact.Subject == activity.ID && fact.Predicate == semantic.RelationUsed && fact.Object == shared.ID {
				result = append(result, edge)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].From+">"+result[i].To < result[j].From+">"+result[j].To })
	return result
}

func partitionByID(partitions []Partition, id string) (Partition, bool) {
	for _, partition := range partitions {
		if partition.ID == id {
			return partition, true
		}
	}
	return Partition{}, false
}

func generatedPartitionDigest(program []byte, partition Partition) (string, error) {
	fileSet := token.NewFileSet()
	file, err := goparser.ParseFile(fileSet, "generated.go", program, goparser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse generated partition: %w", err)
	}
	wanted := map[string]bool{}
	for _, symbol := range partition.Symbols {
		wanted[symbol] = true
	}
	decls := make([]namedDecl, 0, len(wanted))
	for _, declaration := range file.Decls {
		name := declarationName(declaration)
		if name == "" || !wanted[name] {
			continue
		}
		var output bytes.Buffer
		if err := format.Node(&output, fileSet, declaration); err != nil {
			return "", fmt.Errorf("format generated symbol %q: %w", name, err)
		}
		decls = append(decls, namedDecl{Name: name, Source: output.String()})
	}
	if len(decls) != len(wanted) {
		return "", fmt.Errorf("generated output owns %d/%d declared symbols for partition %q", len(decls), len(wanted), partition.ID)
	}
	sort.Slice(decls, func(i, j int) bool { return decls[i].Name < decls[j].Name })
	return hashJSON(struct {
		Partition string      `json:"partition"`
		Decls     []namedDecl `json:"declarations"`
	}{partition.ID, decls}), nil
}

type namedDecl struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

func declarationName(declaration ast.Decl) string {
	switch item := declaration.(type) {
	case *ast.FuncDecl:
		return item.Name.Name
	case *ast.GenDecl:
		if len(item.Specs) != 1 {
			return ""
		}
		switch spec := item.Specs[0].(type) {
		case *ast.TypeSpec:
			return spec.Name.Name
		case *ast.ValueSpec:
			if len(spec.Names) == 1 {
				return spec.Names[0].Name
			}
		}
	}
	return ""
}

func manifestDigest(data []byte) (string, error) {
	var semanticDigest string
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var value struct {
			SemanticDigest string `json:"semantic_digest"`
		}
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			return "", fmt.Errorf("decode generated manifest: %w", err)
		}
		if value.SemanticDigest == "" {
			return "", errorsNew("generated manifest semantic digest is missing")
		}
		if semanticDigest != "" && semanticDigest != value.SemanticDigest {
			return "", errorsNew("generated manifest has contradictory semantic digests")
		}
		semanticDigest = value.SemanticDigest
	}
	if !cache.Digest(semanticDigest).Known() {
		return "", errorsNew("generated manifest semantic digest is unknown")
	}
	return semanticDigest, nil
}

func errorsNew(message string) error { return fmt.Errorf("%s", message) }
