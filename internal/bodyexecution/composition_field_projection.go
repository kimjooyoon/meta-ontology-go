package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type CompositionFieldSite struct {
	Binding        int    `json:"binding"`
	Field          int    `json:"field"`
	Kind           string `json:"kind"`
	DependencyMode string `json:"dependency_mode"`
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
}

type compositionFieldProjection struct {
	sourceSHA256 string
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
	var mismatch error
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || functions[fn.Name.Name][0] == "" {
			continue
		}
		ids := functions[fn.Name.Name]
		current := compositionFieldFunction{name: fn.Name.Name}
		port := 0
		for _, group := range fn.Type.Params.List {
			for _, name := range group.Names {
				if index, found := inventory.inputs[conditionOriginKey{ids[0], ids[1], -port - 1}]; found {
					objects[name.Obj] = index
					current.bindings = append(current.bindings, index)
					current.parameters = append(current.parameters, index)
					current.paramSlots = append(current.paramSlots, port)
					current.paramSites = append(current.paramSites, p.addSites(index, "parameter", "call_parameter_or_input", name, fset, source))
				}
				port++
			}
		}
		ordinal := 0
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			name, _, _, _ := bodycodegen.LocalWriteExpression(n)
			if name == nil || ignorePrefix != "" && strings.HasPrefix(name.Name, ignorePrefix) {
				return true
			}
			index, found := inventory.locals[conditionOriginKey{ids[0], ids[1], ordinal}]
			ordinal++
			if found && inventory.bindings[index].Name != name.Name {
				mismatch = fmt.Errorf("field lexical projection identity differs")
				return false
			}
			if found && name.Obj != nil {
				objects[name.Obj] = index
				current.bindings = append(current.bindings, index)
			}
			return true
		})
		slices.Sort(current.bindings)
		current.bindings = slices.Compact(current.bindings)
		if len(current.bindings) == 0 {
			continue
		}
		if fn.Type.Results != nil && len(fn.Type.Results.List) == 1 {
			if name, ok := fn.Type.Results.List[0].Type.(*ast.Ident); ok {
				for _, binding := range inventory.bindings {
					current.recordOut = current.recordOut || binding.Record.GoName == name.Name
				}
			}
		}
		functionIndex[fn.Name.Name] = len(p.functions)
		p.functions = append(p.functions, current)
	}
	if mismatch != nil {
		return nil, mismatch
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
	skipped := map[*ast.Ident]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		name, value, _, _ := bodycodegen.LocalWriteExpression(n)
		if name != nil {
			skipped[name] = true
			if binding, found := objects[name.Obj]; found {
				first := p.addSites(binding, "record_assignment", "shared_rhs", value, fset, source)
				i := p.addNode(n, value, "record_write", binding, first, 0, fset)
				p.nodes[i].parents = []int{-1}
				if literal, ok := value.(*ast.CompositeLit); ok {
					matched := 0
					for _, element := range literal.Elts {
						if pair, ok := element.(*ast.KeyValueExpr); ok {
							if name, ok := pair.Key.(*ast.Ident); ok {
								for field, descriptor := range p.inventory.bindings[binding].Record.Fields {
									if descriptor.GoName == name.Name {
										p.addNode(pair, pair.Value, "part", binding, first+field, field, fset)
										p.sites[first+field].Kind, p.sites[first+field].DependencyMode = "record_constructor", "scalar_rhs"
										p.sites[first+field].Start, p.sites[first+field].End = fset.Position(pair.Value.Pos()).Offset, fset.Position(pair.Value.End()).Offset
										p.sites[first+field].Expression = source[p.sites[first+field].Start:p.sites[first+field].End]
										matched++
									}
								}
							}
						}
					}
					if matched == len(p.inventory.bindings[binding].Record.Fields) {
						p.nodes[i].mode = 3
					}
				}
				if id, ok := value.(*ast.Ident); ok {
					if parent, found := objects[id.Obj]; found {
						p.nodes[i].parents[0], p.nodes[i].mode = parent, 1
						for j := range p.inventory.bindings[binding].Record.Fields {
							p.sites[first+j].DependencyMode = "field_copy"
						}
					}
				}
				if call, ok := value.(*ast.CallExpr); ok {
					if target, ok := call.Fun.(*ast.Ident); ok {
						if called, found := functions[target.Name]; found && p.functions[called].recordOut {
							p.nodes[i].mode = 2
							for j := range p.inventory.bindings[binding].Record.Fields {
								p.sites[first+j].DependencyMode = "record_return"
							}
						}
					}
				}
			}
		}
		if assign, ok := n.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
			if id, ok := assign.Lhs[0].(*ast.Ident); ok && id.Name == "_" {
				return false
			}
			if selector, ok := assign.Lhs[0].(*ast.SelectorExpr); ok {
				if id, ok := selector.X.(*ast.Ident); ok {
					skipped[id] = true
					if binding, found := objects[id.Obj]; found {
						for field, descriptor := range p.inventory.bindings[binding].Record.Fields {
							if descriptor.GoName == selector.Sel.Name {
								first := len(p.sites)
								start, end := fset.Position(assign.Rhs[0].Pos()).Offset, fset.Position(assign.Rhs[0].End()).Offset
								p.sites = append(p.sites, CompositionFieldSite{Binding: binding, Field: field, Kind: "field_assignment", DependencyMode: "scalar_rhs", Start: start, End: end, Expression: source[start:end]})
								p.addNode(assign, assign.Rhs[0], "write", binding, first, field, fset)
							}
						}
					}
				}
			}
		}
		if ret, ok := n.(*ast.ReturnStmt); ok && current.recordOut && !nestedReturns[ret] && len(ret.Results) == 1 {
			binding := -1
			if id, ok := ret.Results[0].(*ast.Ident); ok {
				if found, ok := objects[id.Obj]; ok {
					binding = found
				}
			}
			p.addNode(ret, ret.Results[0], "return", binding, 0, 0, fset)
		}
		return true
	})
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 {
			if id, ok := assign.Lhs[0].(*ast.Ident); ok && id.Name == "_" {
				return false
			}
		}
		if selector, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := selector.X.(*ast.Ident); ok {
				if binding, found := objects[id.Obj]; found {
					if !skipped[id] {
						for field, descriptor := range p.inventory.bindings[binding].Record.Fields {
							if descriptor.GoName == selector.Sel.Name {
								first := len(p.uses)
								p.uses = append(p.uses, CompositionFieldSite{Binding: binding, Field: field, Kind: "field_read", Start: fset.Position(n.Pos()).Offset, End: fset.Position(n.End()).Offset})
								p.addNode(n, n, "use", binding, first, field, fset)
							}
						}
					}
					return false
				}
			}
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if id, ok := call.Fun.(*ast.Ident); ok {
				if target, found := functions[id.Name]; found && len(p.functions[target].parameters) > 0 {
					i := p.addNode(n, n, "call", 0, 0, 0, fset)
					for _, port := range p.functions[target].paramSlots {
						binding := -1
						if port < len(call.Args) {
							if arg, ok := call.Args[port].(*ast.Ident); ok {
								if found, ok := objects[arg.Obj]; ok {
									binding = found
								}
							}
						}
						p.nodes[i].parents = append(p.nodes[i].parents, binding)
					}
				}
			}
		}
		if id, ok := n.(*ast.Ident); ok && !skipped[id] {
			if binding, found := objects[id.Obj]; found {
				first := len(p.uses)
				for field := range p.inventory.bindings[binding].Record.Fields {
					p.uses = append(p.uses, CompositionFieldSite{Binding: binding, Field: field, Kind: "record_read", Start: fset.Position(id.Pos()).Offset, End: fset.Position(id.End()).Offset})
				}
				p.addNode(n, n, "record_use", binding, first, 0, fset)
			}
		}
		return true
	})
}
