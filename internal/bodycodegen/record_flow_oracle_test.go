package bodycodegen

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// This test-only evaluator reads the exported graph, not the body interpreter
// or generated code. Its deliberately small vocabulary covers these fixtures.
type flowOracle struct {
	t     *testing.T
	flow  *RecordValueFlow
	views map[string]string
	input string
	cache map[uint16]any
}

func newFlowOracle(t *testing.T, source []byte, activity string) *flowOracle {
	t.Helper()
	r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "oracle.gooo", source, activity, false, "")
	if err != nil || r.ValueFlow.Status != "RESOLVED" {
		t.Fatal("oracle source graph", r.ValueFlow, err)
	}
	plan, err := prepareRecordAssembly(context.Background(), "oracle.gooo", source, activity)
	if err != nil {
		t.Fatal(err)
	}
	views := map[string]string{"/computes": plan.body.body}
	for _, choice := range plan.choices {
		views["/alternative:"+choice.ID] = choice.Second
	}
	functions, _, _, err := plan.body.resolvePureCalls(plan.body.body)
	if err != nil {
		t.Fatal(err)
	}
	for _, fn := range functions {
		views[fn.identity.ActivityID+"/computes"] = fn.body
	}
	return &flowOracle{t: t, flow: r.ValueFlow, views: views}
}

func (o *flowOracle) evaluate(root uint16) any {
	if root == 0 {
		return true // no enclosing execution guard
	}
	if value, ok := o.cache[root]; ok {
		return value
	}
	node := o.flow.Nodes[root-1]
	value := o.node(node)
	o.cache[root] = value
	return value
}

func (o *flowOracle) node(n RecordFlowNode) any {
	left := func() any { return o.evaluate(n.Parents[0]) }
	right := func() any { return o.evaluate(n.Parents[1]) }
	switch n.Kind {
	case "input":
		return o.input
	case "literal":
		if n.Operator == "true" || n.Operator == "false" {
			return n.Operator == "true"
		}
		if n.Operator == "0" {
			return int64(0) // omitted low slice bound
		}
		body := o.views[n.Span.ActivityID+"/"+n.Span.View]
		text := body[n.Span.Start:n.Span.End]
		if n.Operator == "STRING" {
			value, err := strconv.Unquote(text)
			if err != nil {
				o.t.Fatal(err)
			}
			return value
		}
		value, err := strconv.ParseInt(text, 0, 64)
		if err != nil {
			o.t.Fatal(err)
		}
		return value
	case "read", "copy", "write", "return", "call", "call_parameter":
		return left()
	case "join":
		if o.evaluate(n.Condition).(bool) {
			return left()
		}
		return right()
	case "guard":
		return right().(bool) && (left().(bool) == (n.Operator == "true"))
	case "guard_join":
		return left().(bool) || right().(bool)
	case "expression":
		switch n.Operator {
		case "len":
			return int64(len(left().(string)))
		case "int64":
			return left().(int64)
		case "&&":
			return left().(bool) && right().(bool)
		case "||":
			return left().(bool) || right().(bool)
		case "!":
			return !left().(bool)
		case "==":
			return left() == right()
		case ">=":
			return left().(int64) >= right().(int64)
		case ">":
			return left().(int64) > right().(int64)
		case "-":
			return left().(int64) - right().(int64)
		case "slice_bounds":
			return [2]int64{left().(int64), right().(int64)}
		case "slice":
			bounds := right().([2]int64)
			return left().(string)[bounds[0]:bounds[1]]
		}
	}
	o.t.Fatalf("oracle does not cover node %+v", n)
	return nil
}

func TestRecordFlowFilenameRootsAgreeWithIndependentGoPolicy(t *testing.T) {
	source, err := os.ReadFile("../../examples/text-operations/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	o := newFlowOracle(t, source, "Classify")
	for _, name := range []string{"", "a", "gooo", "file.gooo", ".file.gooo", "언어.gooo", "🙂.gooo", "x.gooo.gooo", "notes.txt", "file.gooo ", "é.gooo", "README.GOOO"} {
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			o.input, o.cache = name, map[uint16]any{}
			suffix, visible := strings.HasSuffix(name, ".gooo"), !strings.HasPrefix(name, ".")
			want := [3][2]any{{suffix || visible, suffix && visible}, {name, strings.TrimSuffix(name, ".gooo")}, {int64(0), int64(len(name))}}
			for i, choice := range o.flow.Choices {
				for j, root := range [2]uint16{choice.First, choice.Second} {
					if got := o.evaluate(root); got != want[i][j] {
						t.Fatalf("field %d alternative %d: got %v, want %v", i, j, got, want[i][j])
					}
				}
			}
		})
	}
}

func TestRecordFlowHelperThreeReturnPathsHaveCorrectGuards(t *testing.T) {
	source := helperFlowSource("if input == \"\" { return \"empty\" }; if len(input) > 2 { return input[:2] }; return input", "Keep(input)", "input")
	o := newFlowOracle(t, source, "Select")
	for _, input := range []string{"", "a", "ab", "abc", "abcdef"} {
		want := input
		if input == "" {
			want = "empty"
		} else if len(input) > 2 {
			want = input[:2]
		}
		o.input, o.cache = input, map[uint16]any{}
		if got := o.evaluate(o.flow.Choices[0].First); got != want {
			t.Fatalf("input %q: got %v, want %v", input, got, want)
		}
	}
}

func TestRecordFlowHelperCanPassAnInternalRecordBetweenNestedCalls(t *testing.T) {
	source := helperFlowSource("return Read(Make(input))", "Keep(input)", "input")
	source = append(source, []byte("\nentity Inner id \"gooo://inner\" fields { field text id \"gooo://inner/text\" type string required one }\n"+
		"activity Make(Text) -> Inner computes `return Inner{text: input}`\n"+
		"activity Read(Inner) -> Text computes `return input.text`\n")...)
	o := newFlowOracle(t, source, "Select")
	for _, input := range []string{"", "한글", "saved value"} {
		o.input, o.cache = input, map[uint16]any{}
		if got := o.evaluate(o.flow.Choices[0].First); got != input {
			t.Fatalf("nested record transport: got %v, want %q", got, input)
		}
	}
}
