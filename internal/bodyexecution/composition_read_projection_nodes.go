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
		bodycodegen.InspectBodyProjection(f.Body, func(node ast.Node) bool {
			if constructionKind, readKind := collectionReadKinds(node); constructionKind != "" {
				call := node.(*ast.CallExpr)
				c, cf := constructs[conditionOriginKey{ids[0], ids[1], constructionOrdinal}]
				r, rf := lookup[conditionOriginKey{ids[0], ids[1], ordinal}]
				constructionOrdinal++
				ordinal++
				if cf && rf && constructionSites[c].Kind == constructionKind && sites[r].Kind == readKind {
					p.nodes = append(p.nodes, readProjectionNode{node: call, site: c, mapReadSite: r, start: p.offset(call.Pos()), end: p.offset(call.End())})
				}
				return true
			}
			if index, found := local.writes[node]; found {
				localRanges += p.collectLocalWriteNode(node, index)
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

func (p *readProjection) collectLocalWriteNode(node ast.Node, index int) int {
	if _, iteration := node.(*ast.RangeStmt); iteration {
		return 1
	}
	_, value, _, _ := bodycodegen.LocalWriteExpression(node)
	item := readProjectionNode{node: node, start: p.offset(value.Pos()), end: p.offset(value.End()), localKind: "write", localWrite: index}
	if assignment, ok := node.(*ast.AssignStmt); ok {
		if _, compound := bodycodegen.CompoundAssignmentOperator(assignment.Tok); compound {
			item.start, item.end, item.localKind = p.offset(node.Pos()), p.offset(node.End()), "compound_write"
		}
	}
	p.nodes = append(p.nodes, item)
	return 0
}

func collectionReadKinds(node ast.Node) (constructionKind, readKind string) {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return "", ""
	}
	if _, _, mapped := bodycodegen.CollectionMapProjection(call); mapped {
		return "map_values", "map_value"
	}
	if _, _, filtered := bodycodegen.CollectionFilterProjection(call); filtered {
		return "filter_values", "filter_value"
	}
	return "", ""
}
