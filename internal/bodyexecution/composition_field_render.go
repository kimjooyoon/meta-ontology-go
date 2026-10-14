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
			case "write", "part", "record_write", "return":
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

func (r compositionFieldRenderer) node(n compositionFieldNode) string {
	var out strings.Builder
	frame := r.prefix + "FieldScope"
	args := r.prefix + "FieldParents"
	call := func(name string, values ...string) string {
		return r.prefix + name + "(" + strings.Join(values, ",") + ")"
	}
	integer := strconv.Itoa
	count := func(binding int) string { return integer(len(r.plan.inventory.bindings[binding].Record.Fields)) }
	switch n.kind {
	case "signature":
		return "," + args + " ...[32]int32"
	case "scope":
		fn := r.plan.functions[n.function]
		out.WriteString("{" + frame + ":=" + r.prefix + "FieldFrame();")
		for _, binding := range fn.bindings {
			out.WriteString("var " + r.slot(binding) + "=" + r.prefix + "FieldNewSlots();")
		}
		for i, binding := range fn.parameters {
			out.WriteString(call("FieldParameter", integer(fn.paramSites[i]), count(binding), frame, "&"+r.slot(binding), args, integer(i)) + ";")
		}
		out.WriteString(r.render(n.start+1, n.end-1, n.node))
		out.WriteString("}")
	case "record_write":
		parent := -1
		if len(n.parents) > 0 {
			parent = n.parents[0]
		}
		typeName := r.plan.inventory.bindings[n.binding].Record.GoName
		return call("FieldAssignRecord", integer(n.first), count(n.binding), frame, "&"+r.slot(n.binding), r.slot(parent), integer(n.mode),
			"func() "+typeName+" {return "+r.render(n.start, n.end, n.node)+"}")
	case "write", "part":
		typeName := "int64"
		switch r.plan.inventory.bindings[n.binding].Record.Fields[n.field].TypeID {
		case "urn:gooo:type:boolean":
			typeName = "bool"
		case "urn:gooo:type:string":
			typeName = "string"
		}
		if n.kind == "part" {
			return call("FieldPart", integer(n.first), "func() "+typeName+" {return "+r.render(n.start, n.end, n.node)+"}")
		}
		return call("FieldAssign", integer(n.first), frame, "&"+r.slot(n.binding)+"["+integer(n.field)+"]",
			"func() "+typeName+" {return "+r.render(n.start, n.end, n.node)+"}")
	case "use":
		return call("FieldUse", integer(n.first), frame, r.slot(n.binding)+"["+integer(n.field)+"]", r.source[n.start:n.end])
	case "record_use":
		return call("FieldRecordUse", integer(n.first), count(n.binding), frame, r.slot(n.binding), r.source[n.start:n.end])
	case "return":
		return call("FieldReturn", r.slot(n.binding), r.render(n.start, n.end, n.node))
	case "call":
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
	default:
		return r.source[n.start:n.end]
	}
	return out.String()
}

func (r compositionFieldRenderer) sourceOffsets(n ast.Node) (int, int) {
	// The one-file parser starts token positions at one.
	return int(n.Pos()) - 1, int(n.End()) - 1
}
