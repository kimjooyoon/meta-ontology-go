package bodycodegen

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func helperFlowSource(body, first, second string) []byte {
	return []byte(textExpressionPrelude + "activity Keep(Text) -> Text computes `" + body + "`\n" +
		"activity Select(Text) -> Result computes `return Result{text: " + first + ", size: 0, matches: false}` assembling {\n" +
		"choice \"text\" field_value at \"0\" alternative " + fmt.Sprintf("%q", second) + " intent \"Keep the returned text.\"\n" +
		"choice \"size\" field_value at \"1\" alternative \"1\" intent \"Set one.\"\n" +
		"choice \"matches\" field_value at \"2\" alternative \"true\" intent \"Set true.\"\n" +
		"value_case \"[\\\"x\\\"]\" -> \"{\\\"text\\\":\\\"x\\\",\\\"size\\\":1,\\\"matches\\\":true}\"\nattempts \"8\"\n}\n")
}

func reachableFlowNodes(flow *RecordValueFlow, root uint16) map[uint16]bool {
	seen := map[uint16]bool{}
	var visit func(uint16)
	visit = func(id uint16) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		node := flow.Nodes[id-1]
		for _, parent := range [4]uint16{node.Parents[0], node.Parents[1], node.Guard, node.Condition} {
			visit(parent)
		}
	}
	visit(root)
	return seen
}

func TestRecordFlowTraversesFilenameHelpersWithBoundSourceSpans(t *testing.T) {
	source, err := os.ReadFile("../../examples/text-operations/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "names.gooo", source, "Classify", false, jointdecision.RecordOriginSharedFeatureVersion)
	if err != nil || r.ValueFlow.Status != "RESOLVED" || r.Context.Status != "ENCODED" || len(r.ValueFlow.Helpers) != 4 ||
		r.ModelPredictions != 0 || r.CandidateTests != 0 {
		t.Fatal("helper origin context", r, err)
	}
	plan, err := prepareRecordAssembly(context.Background(), "names.gooo", source, "Classify")
	if err != nil {
		t.Fatal(err)
	}
	functions, _, _, err := plan.body.resolvePureCalls(plan.body.body)
	if err != nil {
		t.Fatal(err)
	}
	bodies := map[string]string{}
	for _, fn := range functions {
		bodies[fn.identity.ActivityID] = fn.body
	}
	calls := 0
	for _, node := range r.ValueFlow.Nodes {
		for _, parent := range [4]uint16{node.Parents[0], node.Parents[1], node.Guard, node.Condition} {
			if parent >= node.ID {
				t.Fatal("helper graph is not ordered", node)
			}
		}
		if node.Kind == "call" {
			calls++
			if bodies[node.CalleeID] == "" {
				t.Fatal("call lost its source identity", node)
			}
		}
		span := node.Span
		if span.ActivityID != "" && span.View == "computes" &&
			(span.Start < 0 || span.Start >= span.End || span.End > len(bodies[span.ActivityID])) {
			t.Fatal("helper span does not bind its own body", node)
		}
	}
	if calls != 5 || !flowContains(r.ValueFlow, r.ValueFlow.Choices[1].Second, "join", "") ||
		!flowContains(r.ValueFlow, r.ValueFlow.Choices[1].Second, "expression", "slice") {
		t.Fatal("nested suffix return paths were lost", calls)
	}
}

func TestRecordFlowAlternativeHelperAndInternalRootNameAreIndependent(t *testing.T) {
	source := helperFlowSource("return input", "input", "Keep(input)")
	source = []byte(strings.ReplaceAll(string(source), "Keep", "selected"))
	r := exportedFlow(t, source).ValueFlow
	if len(r.Helpers) != 1 || r.Helpers[0].Name != "selected" ||
		flowContains(r, r.Choices[0].First, "call", "") || !flowContains(r, r.Choices[0].Second, "call", "") {
		t.Fatal("alternative-only closure or private root name changed", r)
	}
	call := r.Nodes[r.Choices[0].Second-1]
	if call.Span.View != "alternative:text" || call.Span.ActivityID != "" || call.CalleeID == "" {
		t.Fatal("alternative call lost its caller and callee coordinates", call)
	}
}

func TestRecordFlowHelperConstantsExcludeUnreachableReturnValues(t *testing.T) {
	for _, body := range []string{
		"if false { return \"dead\" }; return input",
		"if true { return input }; return \"dead\"",
		"if false { if input == \"x\" { return \"dead\" } }; return input",
	} {
		r := exportedFlow(t, helperFlowSource(body, "Keep(input)", "input")).ValueFlow
		for id := range reachableFlowNodes(r, r.Choices[0].First) {
			node := r.Nodes[id-1]
			if node.Span.ActivityID != "" && node.Span.View == "computes" && node.Kind == "literal" &&
				body[node.Span.Start:node.Span.End] == "\"dead\"" {
				t.Fatal("constant-unreachable return affected a helper result", node)
			}
		}
	}
}

func TestRecordFlowHelperPreservesRecordSnapshotsAndParameterIdentity(t *testing.T) {
	body := "let copy = Keep(input0)\ncopy.title = \"draft\"\ncopy.state = \"wait\"\ncopy.reason = copy.state\nreturn copy"
	source := append(flowSource(t, body), []byte("\nactivity Keep(Candidate) -> Candidate computes `let saved = input; let work = input; work.title = \"changed\"; return saved`\n")...)
	r := exportedFlow(t, source).ValueFlow
	for _, node := range r.Nodes {
		if node.Kind == "call" {
			if flowContains(r, node.ID, "write", "") || !flowContains(r, node.ID, "call_parameter", "") || node.FieldID == "" {
				t.Fatal("record snapshot followed the helper's later write", node)
			}
		}
	}
	renamed := []byte(strings.ReplaceAll(string(source), "saved", "photo"))
	other := exportedFlow(t, renamed).ValueFlow
	for _, f := range []*RecordValueFlow{r, other} {
		f.BodySHA256 = ""
		for i := range f.Helpers {
			f.Helpers[i].ProgramSHA256 = ""
		}
	}
	if !reflect.DeepEqual(r, other) {
		t.Fatal("local rename changed source helper relationships")
	}
}

func TestRecordFlowBoundsAreSharedAcrossHelperCalls(t *testing.T) {
	nodes := helperFlowSource("let copy = input; "+strings.Repeat("copy = \"x\"; ", 260)+"return copy", "Keep(input)", "input")
	locals := string(helperFlowSource("return input", "Keep(input)", "input"))
	var declarations, uses strings.Builder
	for i := range 63 {
		fmt.Fprintf(&declarations, "let a%d = \"\"; ", i)
		fmt.Fprintf(&uses, " + a%d", i)
	}
	locals = strings.Replace(locals, "computes `return Result", "computes `"+declarations.String()+"return Result", 1)
	locals = strings.Replace(locals, "text: Keep(input)", "text: Keep(input)"+uses.String(), 1)
	for _, test := range []struct {
		source []byte
		reason string
	}{
		{nodes, "FLOW_NODE_BOUND"}, {[]byte(locals), "FLOW_LOCAL_BOUND"},
	} {
		r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "bounds.gooo", test.source, "Select", false, "")
		if err != nil || r.ValueFlow.Status != "UNRESOLVED" || r.ValueFlow.Reason != test.reason ||
			len(r.ValueFlow.Nodes) != 0 || len(r.ValueFlow.Choices) != 0 || len(r.ValueFlow.Helpers) != 0 {
			t.Fatal("helper limits returned a partial graph", r, err)
		}
	}
}

func TestRecordFlowConcurrentHelperFramesRemainRequestOwned(t *testing.T) {
	source := helperFlowSource("if input == \"\" { return \"empty\" }; return input", "Keep(input)", "input")
	want := exportedFlow(t, source).ValueFlow
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "flow.gooo", source, "Select", false, jointdecision.RecordSharedFeatureVersion)
			if err != nil || !reflect.DeepEqual(want, r.ValueFlow) {
				t.Error("concurrent helper frames differ", err)
			}
		})
	}
	group.Wait()
}
