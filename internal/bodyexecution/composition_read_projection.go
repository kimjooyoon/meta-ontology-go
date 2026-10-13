package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

type readProjectionNode struct {
	node       ast.Node
	site       int
	start, end int
}
type readProjection struct {
	source, prefix string
	fset           *token.FileSet
	nodes          []readProjectionNode
}

func readObservationPrefix(source string) string {
	prefix := "GoooObservedRead"
	for suffix := 0; strings.Contains(source, prefix); suffix++ {
		prefix = "GoooObservedRead" + strconv.Itoa(suffix)
	}
	return prefix
}

func observedReadProjection(source, prefix string, prior Composition, sites []CompositionReadSite) (string, error) {
	p := readProjection{source: source, prefix: prefix, fset: token.NewFileSet()}
	file, err := parser.ParseFile(p.fset, "read.go", source, 0)
	if err != nil {
		return "", err
	}
	lookup := map[conditionOriginKey]int{}
	for i, s := range sites {
		lookup[conditionOriginKey{s.RootActivityID, s.ActivityID, s.NativeOrdinal}] = i
	}
	bindings := arithmeticFunctionIDs(prior)
	for _, declaration := range file.Decls {
		f, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ids, ordinal := bindings[f.Name.Name], 0
		ast.Inspect(f.Body, func(node ast.Node) bool {
			_, _, kind := compositionReadNode(node)
			if kind == "" {
				return true
			}
			i, found := lookup[conditionOriginKey{ids[0], ids[1], ordinal}]
			ordinal++
			if found && sites[i].Kind == kind {
				p.nodes = append(p.nodes, readProjectionNode{node, i, p.offset(node.Pos()), p.offset(node.End())})
			}
			return true
		})
	}
	if len(p.nodes) != len(sites) {
		return "", fmt.Errorf("collection read projection differs from retained sites")
	}
	raw, err := format.Source([]byte(p.render(0, len(source))))
	return string(raw), err
}

func (p readProjection) offset(pos token.Pos) int { return p.fset.Position(pos).Offset }

// Recursive replacement evaluates slice/index expressions once and preserves
// short-circuiting. Range temporaries retain the original single evaluation.
func (p readProjection) render(start, end int) string {
	var out strings.Builder
	cursor := start
	for _, item := range p.nodes {
		if item.start < cursor || item.end > end {
			continue
		}
		out.WriteString(p.source[cursor:item.start])
		id := strconv.Itoa(item.site)
		switch node := item.node.(type) {
		case *ast.IndexExpr:
			base := p.render(p.offset(node.X.Pos()), p.offset(node.X.End()))
			index := p.render(p.offset(node.Index.Pos()), p.offset(node.Index.End()))
			out.WriteString(p.prefix + "Read(" + id + "," + base + "," + index + ")")
		case *ast.RangeStmt:
			sequence, index := p.prefix+"Sequence"+id, p.prefix+"Index"+id
			base := p.render(p.offset(node.X.Pos()), p.offset(node.X.End()))
			out.WriteString("{ " + sequence + ":=" + base + "; for " + index + "," + p.source[p.offset(node.Value.Pos()):p.offset(node.Value.End())] + " := range " + sequence + " { ")
			out.WriteString(p.prefix + "Read(" + id + "," + sequence + "," + index + ");")
			out.WriteString(p.render(p.offset(node.Body.Lbrace)+1, p.offset(node.Body.Rbrace)))
			out.WriteString("} }")
		}
		cursor = item.end
	}
	out.WriteString(p.source[cursor:end])
	return out.String()
}
