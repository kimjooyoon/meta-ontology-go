package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type readerProbeCase struct {
	Name  string          `json:"name"`
	Graph json.RawMessage `json:"graph"`
	Paths []string        `json:"paths"`
}

type readerProbeState struct {
	Status string `json:"status"`
	Phase  string `json:"phase"`
	Offset int    `json:"offset"`
	Steps  int    `json:"steps"`
	Budget int    `json:"step_budget"`
	Paths  string `json:"paths"`
}

func readerProbeGraph(t *testing.T, r TypedValueOrigins, root TypedValueOutput, control, history bool) json.RawMessage {
	t.Helper()
	values := make(map[uint16]uint16)
	ports := make(map[uint16]string)
	for _, v := range r.Flow.RecordVersions {
		values[v.Node] = v.Value
	}
	for _, in := range r.Inputs {
		ports[in.Node] = in.Port
	}
	nodes := make([]any, 0, len(r.Flow.Nodes))
	for _, n := range r.Flow.Nodes {
		nodes = append(nodes, map[string]any{"node_id": n.ID, "a": n.Parents[0], "b": n.Parents[1], "guard": n.Guard, "condition": n.Condition, "version": n.RecordVersion, "kind": n.Kind, "field_id": n.FieldID, "port": ports[n.ID], "bit": int64(1) << ((n.ID - 1) % 32), "value": values[n.ID], "callee_id": n.CalleeID, "activity_id": n.Span.ActivityID})
	}
	pages := []any{}
	for first := 0; first < len(nodes); first += 256 {
		pages = append(pages, map[string]any{"first": first + 1, "nodes": nodes[first:min(first+256, len(nodes))]})
	}
	layouts := []any{}
	for _, l := range r.Flow.RecordLayouts {
		fields := []any{}
		for _, f := range l.Fields {
			fields = append(fields, map[string]any{"field_id": f.ID, "type_id": f.TypeID})
		}
		layouts = append(layouts, map[string]any{"type_id": l.ID, "members": fields})
	}
	payloads := []any{}
	for _, v := range r.Flow.RecordValues {
		payloads = append(payloads, map[string]any{"layout": v.Layout, "roots": v.Fields})
	}
	valuePages := []any{}
	for first := 0; first < len(payloads); first += 256 {
		valuePages = append(valuePages, map[string]any{"first": first + 1, "values": payloads[first:min(first+256, len(payloads))]})
	}
	flow, _ := json.Marshal(r.Flow)
	report, _ := json.Marshal(r)
	graph := map[string]any{"schema": "gooo/record-origin-graph/v1", "source_sha256": r.SourceSHA256, "flow_sha256": digest(flow), "report_sha256": digest(report), "activity_id": r.ActivityID, "field_id": root.FieldID, "root": root.Root, "count": len(nodes), "value_count": len(payloads), "control": control, "history": history, "pages": pages, "layouts": layouts, "value_pages": valuePages}
	raw, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// This oracle reads the original report, independently from the paged payload
// and independently from the Gooo implementation or its generated source.
func readerProbeOracle(t *testing.T, r TypedValueOrigins, root uint16, control, history bool) []string {
	t.Helper()
	seen := make([]bool, len(r.Flow.Nodes)+1)
	seen[root] = true
	for _, n := range slices.Backward(r.Flow.Nodes) {
		if !seen[n.ID] {
			continue
		}
		seen[n.Parents[0]], seen[n.Parents[1]] = true, true
		if control {
			seen[n.Guard], seen[n.Condition] = true, true
		}
		if history {
			seen[n.RecordVersion] = true
		}
	}
	containers := map[uint16]bool{}
	for _, v := range r.Flow.RecordVersions {
		containers[v.Node] = true
	}
	paths := []string{}
	for _, in := range r.Inputs {
		if !seen[in.Node] || containers[in.Node] {
			continue
		}
		parts := []string{}
		for id := in.Node; id != 0; id = r.Flow.Nodes[id-1].Parents[0] {
			n := r.Flow.Nodes[id-1]
			if n.Kind != "input" {
				t.Fatal("source path lost input ancestry")
			}
			if n.FieldID != "" {
				parts = append(parts, n.FieldID)
			}
		}
		slices.Reverse(parts)
		paths = append(paths, in.Port+":"+strings.Join(parts, " > "))
	}
	slices.Sort(paths)
	return paths
}

func readerProbeInputs(values ...json.RawMessage) string {
	raw, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func readerProbeObserve(t *testing.T, raw json.RawMessage) readerProbeState {
	t.Helper()
	var s readerProbeState
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRecordOriginReaderInterpretedAndNativeQueries(t *testing.T) {
	readerProbeDir := t.TempDir()
	source, err := os.ReadFile("../../examples/record-origin-reader/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	programs := map[string]Result{}
	evaluators := map[string]*criteriaTupleEvaluator{}
	for _, name := range []string{"Begin", "Step", "Finish"} {
		p, err := Generate("reader.gooo", source, name)
		if err != nil {
			t.Fatal(name, err)
		}
		if !p.Report.TypecheckPassed || !p.Report.DeterministicReplay {
			t.Fatal(name, "unverified generation")
		}
		e, err := newCriteriaTupleEvaluator(context.Background(), []byte(p.Source), p.Report.Activity, p.Report.RecordTypes, nil)
		if err != nil {
			t.Fatal(name, err)
		}
		programs[name], evaluators[name] = p, e
		raw, _ := json.Marshal(p)
		if err := os.WriteFile(filepath.Join(readerProbeDir, strings.ToLower(name)+"-functional-program.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	actual, err := os.ReadFile("../../examples/openfga-model-change/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	report, err := InspectTypedValueOrigins("comparison.gooo", actual, "Compare")
	if err != nil || report.Status != "RESOLVED" {
		t.Fatal(err, report.Reason)
	}
	cases := []readerProbeCase{}
	for _, out := range report.Outputs {
		cases = append(cases, readerProbeCase{out.FieldID, readerProbeGraph(t, report, out, false, false), readerProbeOracle(t, report, out.Root, false, false)})
	}
	for _, mode := range [][2]bool{{true, false}, {false, true}, {true, true}} {
		out := report.Outputs[len(report.Outputs)-1]
		cases = append(cases, readerProbeCase{fmt.Sprintf("edge-modes-%v", mode), readerProbeGraph(t, report, out, mode[0], mode[1]), readerProbeOracle(t, report, out.Root, mode[0], mode[1])})
	}
	before := "input:gooo://authorization-delta/pair/before > gooo://authorization-delta/observation/input-origin"
	after := "input:gooo://authorization-delta/pair/after > gooo://authorization-delta/observation/input-origin"
	authored := map[string][]string{"gooo://authorization-delta/comparison/before-input": {before}, "gooo://authorization-delta/comparison/after-input": {after}, "gooo://authorization-delta/comparison/input-match": {after, before}, "gooo://authorization-delta/comparison/case": {"input:gooo://authorization-delta/pair/case"}}
	for _, c := range cases {
		if want, ok := authored[c.Name]; ok {
			slices.Sort(want)
			if !reflect.DeepEqual(c.Paths, want) {
				t.Fatal("independent oracle disagrees with authored source origin", c.Name, c.Paths, want)
			}
		}
	}
	for _, activity := range []string{"After", "Before", "Both", "Preserve"} {
		r, err := InspectTypedValueOrigins("paths.gooo", []byte(nominalOriginSource), activity)
		if err != nil || r.Status != "RESOLVED" {
			t.Fatal(err, r.Reason)
		}
		out := r.Outputs[0]
		cases = append(cases, readerProbeCase{"required-" + activity, readerProbeGraph(t, r, out, false, false), readerProbeOracle(t, r, out.Root, false, false)})
	}
	stepsTotal := 0
	for _, c := range cases {
		cursor, err := evaluators["Begin"].evaluate(readerProbeInputs(c.Graph), 0)
		if err != nil {
			t.Fatal(c.Name, "begin", err)
		}
		state := readerProbeObserve(t, cursor)
		calls := 0
		if state.Status != "MORE" || state.Budget < 1 || state.Budget > 258 {
			t.Fatal(c.Name, "invalid source work plan", state)
		}
		for state.Status == "MORE" && calls < 260 {
			next, err := evaluators["Step"].evaluate(readerProbeInputs(c.Graph, cursor), calls)
			if err != nil {
				t.Fatal(c.Name, state.Phase, state.Offset, err)
			}
			if bytes.Equal(cursor, next) {
				t.Fatal(c.Name, "stalled cursor")
			}
			cursor = next
			calls++
			state = readerProbeObserve(t, cursor)
		}
		if state.Status != "RESOLVED" || state.Steps != state.Budget || calls != state.Budget {
			t.Fatal(c.Name, "did not finish declared work", state, calls)
		}
		summary, err := evaluators["Finish"].evaluate(readerProbeInputs(cursor), 0)
		if err != nil {
			t.Fatal(c.Name, "finish", err)
		}
		final := readerProbeObserve(t, summary)
		paths := []string{}
		if final.Paths != "" {
			paths = strings.Split(strings.TrimSuffix(final.Paths, ";"), ";")
		}
		slices.Sort(paths)
		if final.Status != "RESOLVED" || !reflect.DeepEqual(paths, c.Paths) {
			t.Fatal(c.Name, "wrong original source paths", paths, c.Paths)
		}
		stepsTotal += calls
		t.Logf("%s: source steps=%d/%d, paths=%d", c.Name, state.Steps, state.Budget, len(paths))
	}
	raw, _ := json.Marshal(cases)
	if err := os.WriteFile(filepath.Join(readerProbeDir, "native-cases.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	readerProbeNative(t, programs, readerProbeDir)
	t.Logf("interpreted + native queries=%d each; source nodes=%d; step closure=%d/4096; interpreted steps=%d; limits unchanged", len(cases), len(report.Flow.Nodes), programs["Step"].Report.CallClosure.MaxCallsPerInvocation, stepsTotal)
}

func readerProbeNative(t *testing.T, programs map[string]Result, readerProbeDir string) {
	t.Helper()
	var generated strings.Builder
	generated.WriteString("package recordreader\n")
	seen := map[string]string{}
	for _, name := range []string{"Step", "Begin", "Finish"} {
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, "generated.go", programs[name].Source, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			key := ""
			switch n := decl.(type) {
			case *ast.FuncDecl:
				key = "func " + n.Name.Name
			case *ast.GenDecl:
				key = "type " + n.Specs[0].(*ast.TypeSpec).Name.Name
			default:
				t.Fatal("unexpected generated declaration")
			}
			var body bytes.Buffer
			if err := format.Node(&body, set, decl); err != nil {
				t.Fatal(err)
			}
			next := body.String()
			if previous, ok := seen[key]; ok {
				if previous != next {
					t.Fatal("shared generated declaration differs", key)
				}
				continue
			}
			seen[key] = next
			generated.WriteString(next + "\n")
		}
	}
	graph := recordTypeByName(programs["Step"].Report.RecordTypes, "Graph")
	if graph == nil {
		t.Fatal("missing generated graph")
	}
	native := fmt.Sprintf(`package recordreader
import("testing";"os";"encoding/json";"reflect";"strings";"slices";"bytes")
func TestOriginalOriginPaths(t *testing.T) {
 var cases []struct{Name string;Graph json.RawMessage;Paths []string}
 raw,err:=os.ReadFile(%s);if err!=nil {t.Fatal(err)};if err:=json.Unmarshal(raw,&cases);err!=nil {t.Fatal(err)}
 for _,c:=range cases {
  var graph %s;if err:=json.Unmarshal(c.Graph,&graph);err!=nil {t.Fatal(err)}
  before,_:=json.Marshal(graph);cursor:=Begin(graph);calls:=0
  for {
   raw,err:=json.Marshal(cursor);if err!=nil {t.Fatal(err)}
   var state struct{Status string;Steps int;StepBudget int `+"`json:\"step_budget\"`"+`};if err:=json.Unmarshal(raw,&state);err!=nil {t.Fatal(err)}
   if state.Status!="MORE" {if state.Status!="RESOLVED"||state.Steps!=state.StepBudget||calls!=state.StepBudget {t.Fatal(c.Name,state,calls)};break}
   if calls>=260 {t.Fatal("native source work did not terminate")};cursor=Step(graph,cursor);calls++
  }
  raw,err:=json.Marshal(Finish(cursor));if err!=nil {t.Fatal(err)};var result struct{Status string;Paths string};if err:=json.Unmarshal(raw,&result);err!=nil {t.Fatal(err)}
  paths:=[]string{};if result.Paths!="" {paths=strings.Split(strings.TrimSuffix(result.Paths,";"),";")};slices.Sort(paths)
  if result.Status!="RESOLVED"||!reflect.DeepEqual(paths,c.Paths) {t.Fatal(c.Name,paths,c.Paths)}
  after,_:=json.Marshal(graph);if !bytes.Equal(before,after) {t.Fatal(c.Name,"native query mutated original graph")}
 }
}
`, strconv.Quote(filepath.Join(readerProbeDir, "native-cases.json")), graph.GoName)
	dir := filepath.Join(readerProbeDir, "native")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"go.mod": "module reader-prototype\n\ngo 1.27.3\n", "generated.go": generated.String(), "generated_test.go": native} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "test", "-p", "1", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	output, err := cmd.CombinedOutput()
	os.WriteFile(filepath.Join(readerProbeDir, "native.log"), output, 0600)
	if err != nil {
		t.Fatalf("native source query: %v\n%s", err, output)
	}
}
