package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/token"
)

type CompositionFieldSite struct {
	Binding        int    `json:"binding"`
	Field          int    `json:"field"`
	Kind           string `json:"kind"`
	DependencyMode string `json:"dependency_mode"`
	Operation      string `json:"operation,omitempty"`
	Start          int    `json:"projection_start"`
	End            int    `json:"projection_end"`
	Expression     string `json:"projection_expression"`
}

type compositionFieldNode struct {
	node           ast.Node
	start, end     int
	kind           string
	binding, first int
	field, mode    int
	fault          int
	function       int
	parents        []int
}

type compositionFieldFunction struct {
	name       string
	bindings   []int
	parameters []int
	paramSites []int
	paramSlots []int
	recordOut  bool
	scalarVars []int
	returns    map[*ast.ReturnStmt]int
}

type compositionFieldProjection struct {
	sourceSHA256 string
	faultPrefix  string
	sources      *CompositionFieldSources
	inventory    compositionFieldInventory
	sites        []CompositionFieldSite
	uses         []CompositionFieldSite
	nodes        []compositionFieldNode
	functions    []compositionFieldFunction
}

func buildCompositionFieldProjection(file *ast.File, fset *token.FileSet, source string, prior Composition, ignorePrefix string) (*compositionFieldProjection, error) {
	inventory, err := compositionFieldBindings(prior)
	if err != nil {
		return nil, err
	}
	p := &compositionFieldProjection{inventory: inventory, sourceSHA256: digest([]byte(source))}
	functions := arithmeticFunctionIDs(prior)
	functionIndex := map[string]int{}
	objects := map[*ast.Object]int{}
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || functions[fn.Name.Name][0] == "" {
			continue
		}
		ids := functions[fn.Name.Name]
		current, err := p.bindFunction(fn, ids, objects, fset, source, ignorePrefix)
		if err != nil {
			return nil, err
		}
		if len(current.bindings) == 0 {
			continue
		}
		functionIndex[fn.Name.Name] = len(p.functions)
		p.functions = append(p.functions, current)
	}
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		index, found := functionIndex[fn.Name.Name]
		if !found {
			continue
		}
		p.collectFunction(fn, index, objects, functionIndex, fset, source)
	}
	for i := range p.uses {
		p.uses[i].Expression = source[p.uses[i].Start:p.uses[i].End]
	}
	if len(p.sites) > 4096 || len(p.uses) > 65536 || len(p.nodes) > 131072 {
		return nil, fmt.Errorf("record field projection bound exceeded")
	}
	if err := p.bindSources(file, prior); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *compositionFieldProjection) addSites(binding int, kind, mode string, n ast.Node, fset *token.FileSet, source string) int {
	first := len(p.sites)
	start, end := fset.Position(n.Pos()).Offset, fset.Position(n.End()).Offset
	for field := range p.inventory.bindings[binding].Record.Fields {
		p.sites = append(p.sites, CompositionFieldSite{Binding: binding, Field: field, Kind: kind,
			DependencyMode: mode, Start: start, End: end, Expression: source[start:end]})
	}
	return first
}

func (p *compositionFieldProjection) addNode(n ast.Node, value ast.Node, kind string, binding, first, field int, fset *token.FileSet) int {
	index := len(p.nodes)
	p.nodes = append(p.nodes, compositionFieldNode{node: n, start: fset.Position(value.Pos()).Offset,
		end: fset.Position(value.End()).Offset, kind: kind, binding: binding, first: first, field: field})
	return index
}

func (p *compositionFieldProjection) collectFunction(fn *ast.FuncDecl, function int, objects map[*ast.Object]int, functions map[string]int, fset *token.FileSet, source string) {
	index := p.addNode(fn.Body, fn.Body, "scope", 0, 0, 0, fset)
	p.nodes[index].function = function
	current := p.functions[function]
	if len(current.parameters) > 0 {
		p.nodes = append(p.nodes, compositionFieldNode{node: fn.Type.Params, start: fset.Position(fn.Type.Params.Closing).Offset,
			end: fset.Position(fn.Type.Params.Closing).Offset, kind: "signature"})
	}
	nestedReturns := map[*ast.ReturnStmt]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if literal, ok := n.(*ast.FuncLit); ok {
			ast.Inspect(literal.Body, func(n ast.Node) bool {
				if ret, ok := n.(*ast.ReturnStmt); ok {
					nestedReturns[ret] = true
				}
				return true
			})
		}
		return true
	})
	skipped := p.collectFunctionWrites(fn, current, nestedReturns, objects, functions, fset, source)
	p.collectFunctionReads(fn, skipped, objects, functions, fset)
}
