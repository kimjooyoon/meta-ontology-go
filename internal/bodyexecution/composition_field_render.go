package bodyexecution

import (
	"cmp"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"strings"
)

type compositionFieldRenderer struct {
	source, prefix string
	plan           *compositionFieldProjection
}

func observedCompositionFieldProjection(source, prefix string, prior Composition) (string, *compositionFieldProjection, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "observed-fields.go", source, 0)
	if err != nil {
		return "", nil, err
	}
	plan, err := buildCompositionFieldProjection(file, fset, source, prior, prefix)
	if err != nil {
		return "", nil, err
	}
	slices.SortStableFunc(plan.nodes, func(a, b compositionFieldNode) int {
		if a.start != b.start {
			return cmp.Compare(a.start, b.start)
		}
		if a.end != b.end {
			return cmp.Compare(b.end, a.end)
		}
		priority := func(s string) int {
			switch s {
			case "scope":
				return 0
			case "write", "part", "record_write", "return", "return_write":
				return 1
			case "call":
				return 2
			}
			return 3
		}
		return cmp.Compare(priority(a.kind), priority(b.kind))
	})
	r := compositionFieldRenderer{source: source, prefix: prefix, plan: plan}
	raw, err := format.Source([]byte(r.render(0, len(source), nil)))
	return string(raw), plan, err
}

func (r compositionFieldRenderer) slot(binding int) string {
	if binding < 0 {
		return r.prefix + "FieldNewSlots()"
	}
	return r.prefix + "FieldSlot" + strconv.Itoa(binding)
}

func (r compositionFieldRenderer) render(start, end int, exclude ast.Node) string {
	var out strings.Builder
	cursor := start
	for _, item := range r.plan.nodes {
		if item.node == exclude || item.start < cursor || item.end > end {
			continue
		}
		out.WriteString(r.source[cursor:item.start])
		out.WriteString(r.node(item))
		cursor = item.end
	}
	out.WriteString(r.source[cursor:end])
	return out.String()
}

func (r compositionFieldRenderer) call(name string, values ...string) string {
	return r.prefix + name + "(" + strings.Join(values, ",") + ")"
}

func (r compositionFieldRenderer) fieldCount(binding int) string {
	return strconv.Itoa(len(r.plan.inventory.bindings[binding].Record.Fields))
}

func (r compositionFieldRenderer) node(n compositionFieldNode) string {
	frame := r.prefix + "FieldScope"
	switch n.kind {
	case "signature":
		return "," + r.prefix + "FieldParents ...[32]int32"
	case "scope":
		return r.scope(n)
	case "return_write":
		return r.returnWrite(n)
	case "record_write":
		return r.recordWrite(n)
	case "write", "part":
		return r.scalarWrite(n)
	case "use":
		return r.call("FieldUse", strconv.Itoa(n.first), frame, r.slot(n.binding)+"["+strconv.Itoa(n.field)+"]", r.source[n.start:n.end])
	case "record_use":
		return r.call("FieldRecordUse", strconv.Itoa(n.first), r.fieldCount(n.binding), frame, r.slot(n.binding), r.source[n.start:n.end])
	case "return":
		return r.call("FieldReturn", r.slot(n.binding), r.render(n.start, n.end, n.node))
	case "call":
		return r.helperCall(n)
	default:
		return r.source[n.start:n.end]
	}
}

func (r compositionFieldRenderer) scope(n compositionFieldNode) string {
	var out strings.Builder
	fn := r.plan.functions[n.function]
	frame := r.prefix + "FieldScope"
	out.WriteString("{" + frame + ":=" + r.prefix + "FieldFrame();")
	for _, binding := range fn.bindings {
		out.WriteString("var " + r.slot(binding) + "=" + r.prefix + "FieldNewSlots();")
	}
	for i, binding := range fn.parameters {
		out.WriteString(r.call("FieldParameter", strconv.Itoa(fn.paramSites[i]), r.fieldCount(binding), frame,
			"&"+r.slot(binding), r.prefix+"FieldParents", strconv.Itoa(i)) + ";")
	}
	out.WriteString(r.render(n.start+1, n.end-1, n.node))
	out.WriteString("}")
	return out.String()
}

func (r compositionFieldRenderer) recordWrite(n compositionFieldNode) string {
	parent := -1
	if len(n.parents) > 0 {
		parent = n.parents[0]
	}
	typeName := r.plan.inventory.bindings[n.binding].Record.GoName
	return r.call("FieldAssignRecord", strconv.Itoa(n.first), r.fieldCount(n.binding), r.prefix+"FieldScope",
		"&"+r.slot(n.binding), r.slot(parent), strconv.Itoa(n.mode),
		"func() "+typeName+" {return "+r.render(n.start, n.end, n.node)+"}")
}

func (r compositionFieldRenderer) scalarWrite(n compositionFieldNode) string {
	typeName := "int64"
	switch r.plan.inventory.bindings[n.binding].Record.Fields[n.field].TypeID {
	case "urn:gooo:type:boolean":
		typeName = "bool"
	case "urn:gooo:type:string":
		typeName = "string"
	}
	value := "func() " + typeName + " {return " + r.render(n.start, n.end, n.node) + "}"
	if n.kind == "part" {
		return r.call("FieldPart", strconv.Itoa(n.first), value)
	}
	return r.call("FieldAssign", strconv.Itoa(n.first), r.prefix+"FieldScope",
		"&"+r.slot(n.binding)+"["+strconv.Itoa(n.field)+"]", value)
}

func (r compositionFieldRenderer) helperCall(n compositionFieldNode) string {
	var out strings.Builder
	original := n.node.(*ast.CallExpr)
	out.WriteString(r.source[n.start:n.start+len(original.Fun.(*ast.Ident).Name)] + "(")
	for i, argument := range original.Args {
		if i > 0 {
			out.WriteString(",")
		}
		// Argument node positions retain source byte ownership in this parse.
		start, end := r.sourceOffsets(argument)
		out.WriteString(r.render(start, end, nil))
	}
	for _, binding := range n.parents {
		out.WriteString("," + r.slot(binding))
	}
	out.WriteString(")")
	return out.String()
}

func (r compositionFieldRenderer) sourceOffsets(n ast.Node) (int, int) {
	// The one-file parser starts token positions at one.
	return int(n.Pos()) - 1, int(n.End()) - 1
}
