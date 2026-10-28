package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func readerBoundarySource(result string) []byte {
	declaration := `package boundary
namespace boundary
entity Integer id "urn:gooo:type:integer"
entity Leaf id "boundary://leaf" fields {
 field value id "boundary://leaf/value" type integer required one
}
entity Box id "boundary://box" fields {
 field leaf id "boundary://box/leaf" type Leaf required one
}
activity Read(input Box) -> Integer id "boundary://read" computes `
	body := "let result=input.leaf.value\n" + strings.Repeat("result=result+1\n", 126) + "return " + result
	return []byte(declaration + fmt.Sprintf("%q\n", body))
}

func readerMaskWords(t *testing.T, raw json.RawMessage) []int64 {
	t.Helper()
	var scalar int64
	if json.Unmarshal(raw, &scalar) == nil {
		return []int64{scalar}
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	if raw, ok := object["bits"]; ok {
		var word int64
		if err := json.Unmarshal(raw, &word); err != nil {
			t.Fatal(err)
		}
		return []int64{word}
	}
	return append(readerMaskWords(t, object["left"]), readerMaskWords(t, object["right"])...)
}

func TestRecordOriginReaderUpperBoundaries(t *testing.T) {
	reader, err := os.ReadFile("../../examples/record-origin-reader/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	programs := map[string]Result{}
	prepared := map[string]*FixedProgram{}
	for _, name := range []string{"Begin", "Step", "Finish"} {
		program, err := Generate("reader.gooo", reader, name)
		if err != nil {
			t.Fatal(err)
		}
		programs[name] = program
		fixed, err := NewFixedProgram(context.Background(), "reader.gooo", reader, name)
		if err != nil {
			t.Fatal(err)
		}
		prepared[name] = fixed
	}
	cases := []readerProbeCase{}
	for _, row := range []struct {
		nodes      int
		expression string
		history    bool
	}{{511, "result", false}, {512, "-result", false}, {512, "-result", true}} {
		report, err := InspectTypedValueOrigins("boundary.gooo", readerBoundarySource(row.expression), "Read")
		if err != nil || report.Status != "RESOLVED" {
			t.Fatal(err, report.Reason)
		}
		if len(report.Flow.Nodes) != row.nodes || int(report.Outputs[0].Root) != row.nodes {
			t.Fatal("lost exact boundary", len(report.Flow.Nodes), report.Outputs)
		}
		graph := readerProbeGraph(t, report, report.Outputs[0], false, row.history)
		first, err := prepared["Begin"].Evaluate(context.Background(), []byte(readerProbeInputs(graph)))
		if err != nil {
			t.Fatal(err)
		}
		cursor := first.Value
		state := readerProbeObserve(t, cursor)
		for state.Status == "MORE" {
			if state.Steps >= state.Budget {
				t.Fatal("source work plan did not terminate")
			}
			next, err := prepared["Step"].Evaluate(context.Background(), []byte(readerProbeInputs(graph, cursor)))
			if err != nil {
				t.Fatal(err)
			}
			cursor = next.Value
			state = readerProbeObserve(t, cursor)
		}
		final, err := prepared["Finish"].Evaluate(context.Background(), []byte(readerProbeInputs(cursor)))
		if err != nil {
			t.Fatal(err)
		}
		summary := readerProbeObserve(t, final.Value)
		if summary.Status != "RESOLVED" || summary.Paths != "input:boundary://box/leaf > boundary://leaf/value;" || state.Steps != 50 || state.Budget != 50 {
			t.Fatal(row, summary, state)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(cursor, &raw); err != nil {
			t.Fatal(err)
		}
		words := readerMaskWords(t, raw["mask"])
		want := make([]int64, 16)
		for i := range want {
			want[i] = 4294967295
		}
		// Original source read4 uses scalar input2 as data and record read3 as
		// history. Data-only reachability therefore omits bit2; history includes it.
		if report.Flow.Nodes[2].Kind != "read" || report.Flow.Nodes[3].Parents[0] != 2 || report.Flow.Nodes[3].RecordVersion != 3 {
			t.Fatal("source history/data separation changed")
		}
		if !row.history {
			want[0] = 4294967291
		}
		if row.nodes == 511 {
			want[15] = 2147483647
		}
		if !reflect.DeepEqual(words, want) {
			t.Fatal(row, "32-bit boundary changed visited origins", words, want)
		}
		cases = append(cases, readerProbeCase{Name: fmt.Sprintf("nodes%d-history%v", row.nodes, row.history), Graph: graph, Paths: []string{"input:boundary://box/leaf > boundary://leaf/value"}, Words: want})
	}
	directory := t.TempDir()
	raw, _ := json.Marshal(cases)
	if err := os.WriteFile(filepath.Join(directory, "native-cases.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	readerProbeNative(t, programs, directory)
}

func TestRecordOriginReaderRejectsOneNodeBeyondBoundary(t *testing.T) {
	report, err := InspectTypedValueOrigins("overflow.gooo", readerBoundarySource("result+0"), "Read")
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "UNRESOLVED" || report.Reason != "FLOW_NODE_BOUND" || report.Flow == nil || len(report.Flow.Nodes) != 0 || len(report.Outputs) != 0 || len(report.Flow.RecordLayouts) != 0 || len(report.Flow.RecordValues) != 0 || len(report.Flow.RecordVersions) != 0 {
		t.Fatal("513-node source exposed an accepted prefix", report.Status, report.Reason)
	}
}
