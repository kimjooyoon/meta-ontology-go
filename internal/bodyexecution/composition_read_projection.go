package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type readProjectionNode struct {
	node        ast.Node
	site        int
	elementType string
	start, end  int
	localKind   string
	localWrite  int
	variables   []int
}
type readProjection struct {
	source, prefix string
	fset           *token.FileSet
	nodes          []readProjectionNode
	localSites     []CompositionLocalSite
	localUseSites  []CompositionLocalUseSite
}

func readObservationPrefix(source string) string {
	prefix := "GoooObservedRead"
	for suffix := 0; strings.Contains(source, prefix); suffix++ {
		prefix = "GoooObservedRead" + strconv.Itoa(suffix)
	}
	return prefix
}

func observedReadProjection(source, prefix string, prior Composition, sites []CompositionReadSite, constructed ...[]CompositionConstructionSite) (string, error) {
	var constructions []CompositionConstructionSite
	if len(constructed) > 0 {
		constructions = constructed[0]
	}
	return observedReadProjectionWithLocals(source, prefix, prior, sites, constructions, nil, nil)
}

func observedReadProjectionWithLocals(source, prefix string, prior Composition, sites []CompositionReadSite, constructionSites []CompositionConstructionSite, locals []CompositionLocalSite, uses []CompositionLocalUseSite) (string, error) {
	p := readProjection{source: source, prefix: prefix, fset: token.NewFileSet()}
	p.localSites, p.localUseSites = locals, uses
	file, err := parser.ParseFile(p.fset, "read.go", source, 0)
	if err != nil {
		return "", err
	}
	lookup := map[conditionOriginKey]int{}
	for i, s := range sites {
		lookup[conditionOriginKey{s.RootActivityID, s.ActivityID, s.NativeOrdinal}] = i
	}
	constructs := map[conditionOriginKey]int{}
	for i, s := range constructionSites {
		constructs[conditionOriginKey{s.RootActivityID, s.ActivityID, s.NativeOrdinal}] = i
	}
	bindings := arithmeticFunctionIDs(prior)
	local, err := bindLocalProjection(file, p.fset, source, prior, locals)
	if err != nil {
		return "", err
	}
	if len(local.sites) != len(uses) {
		return "", fmt.Errorf("local use count differs from retained sites")
	}
	for index, current := range local.sites {
		original := uses[index]
		if current.RootActivityID != original.RootActivityID || current.ActivityID != original.ActivityID || current.VariableOrdinal != original.VariableOrdinal || current.Name != original.Name {
			return "", fmt.Errorf("local use binding changed during native observation")
		}
	}
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
	if len(p.nodes) != len(sites)+len(constructionSites)+len(locals)-localRanges+len(uses)+len(local.scopes) {
		return "", fmt.Errorf("collection read projection differs from retained sites")
	}
	raw, err := format.Source([]byte(p.render(0, len(source))))
	return string(raw), err
}

func (p readProjection) offset(pos token.Pos) int { return p.fset.Position(pos).Offset }

// Recursive replacement evaluates slice/index expressions once and preserves
// short-circuiting. Range temporaries retain the original single evaluation.
func (p readProjection) render(start, end int) string {
	return p.renderExcept(start, end, nil)
}

func (p readProjection) renderExcept(start, end int, exclude ast.Node) string {
	var out strings.Builder
	cursor := start
	for _, item := range p.nodes {
		if item.node == exclude || item.start < cursor || item.end > end {
			continue
		}
		out.WriteString(p.source[cursor:item.start])
		id := strconv.Itoa(item.site)
		if item.localKind != "" {
			out.WriteString(p.renderLocal(item))
			cursor = item.end
			continue
		}
		switch node := item.node.(type) {
		case *ast.CompositeLit:
			out.WriteString(p.prefix + "Make[" + item.elementType + "](" + id)
			for _, element := range node.Elts {
				out.WriteString(",func() " + item.elementType + " {return " + p.render(p.offset(element.Pos()), p.offset(element.End())) + "}")
			}
			out.WriteString(")")
		case *ast.IndexExpr:
			base := p.render(p.offset(node.X.Pos()), p.offset(node.X.End()))
			index := p.render(p.offset(node.Index.Pos()), p.offset(node.Index.End()))
			out.WriteString(p.prefix + "Read(" + id + "," + base + "," + index + ")")
		case *ast.RangeStmt:
			sequence, index := p.prefix+"Sequence"+id, p.prefix+"Index"+id
			base := p.render(p.offset(node.X.Pos()), p.offset(node.X.End()))
			out.WriteString("{ " + sequence + ":=" + base + "; for " + index + "," + p.source[p.offset(node.Value.Pos()):p.offset(node.Value.End())] + " := range " + sequence + " { ")
			read := p.prefix + "Read(" + id + "," + sequence + "," + index + ")"
			if item.localWrite >= 0 {
				out.WriteString(p.localAssignment(item.localWrite, read) + ";")
			} else {
				out.WriteString(read + ";")
			}
			out.WriteString(p.render(p.offset(node.Body.Lbrace)+1, p.offset(node.Body.Rbrace)))
			out.WriteString("} }")
		}
		cursor = item.end
	}
	out.WriteString(p.source[cursor:end])
	return out.String()
}
