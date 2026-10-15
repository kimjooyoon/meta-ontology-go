package bodyexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"go/ast"
)

func (p *readProjection) collectLocalReadNodes(file *ast.File, bindings map[string][2]string, local localProjectionBindings, lookup, constructs map[conditionOriginKey]int, sites []CompositionReadSite, constructionSites []CompositionConstructionSite) int {
	localRanges := 0
	for _, declaration := range file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ids, ordinal := bindings[f.Name.Name], 0
		constructionOrdinal := 0
		if variables := local.scopes[f.Body]; len(variables) > 0 {
			p.nodes = append(p.nodes, readProjectionNode{node: f.Body, start: p.offset(f.Body.Pos()), end: p.offset(f.Body.End()), localKind: "scope", variables: variables})
		}
		ast.Inspect(f.Body, func(node ast.Node) bool {
			if index, found := local.writes[node]; found {
				if _, iteration := node.(*ast.RangeStmt); !iteration {
					_, value, _, _ := bodycodegen.LocalWriteExpression(node)
					p.nodes = append(p.nodes, readProjectionNode{node: node, start: p.offset(value.Pos()), end: p.offset(value.End()), localKind: "write", localWrite: index})
					if assignment, ok := node.(*ast.AssignStmt); ok {
						if _, compound := bodycodegen.CompoundAssignmentOperator(assignment.Tok); compound {
							p.nodes[len(p.nodes)-1].start = p.offset(node.Pos())
							p.nodes[len(p.nodes)-1].end = p.offset(node.End())
							p.nodes[len(p.nodes)-1].localKind = "compound_write"
						}
					}
				} else {
					localRanges++
				}
			}
			if name, ok := node.(*ast.Ident); ok {
				if index, found := local.uses[name]; found {
					p.nodes = append(p.nodes, readProjectionNode{node: node, site: index, start: p.offset(node.Pos()), end: p.offset(node.End()), localKind: "use"})
				}
			}
			if literal := constructionLiteral(node); literal != nil {
				i, found := constructs[conditionOriginKey{ids[0], ids[1], constructionOrdinal}]
				constructionOrdinal++
				if found {
					p.nodes = append(p.nodes, readProjectionNode{node: node, site: i, elementType: constructionSites[i].ElementType, start: p.offset(node.Pos()), end: p.offset(node.End())})
				}
			}
			_, _, kind := compositionReadNode(node)
			if kind == "" {
				return true
			}
			i, found := lookup[conditionOriginKey{ids[0], ids[1], ordinal}]
			ordinal++
			if found && sites[i].Kind == kind {
				item := readProjectionNode{node: node, site: i, start: p.offset(node.Pos()), end: p.offset(node.End()), localWrite: -1}
				if index, found := local.writes[node]; found {
					item.localWrite = index
				}
				p.nodes = append(p.nodes, item)
			}
			return true
		})
	}
	return localRanges
}
