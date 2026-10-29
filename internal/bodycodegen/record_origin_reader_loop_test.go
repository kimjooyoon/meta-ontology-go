package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRecordOriginReaderLoopClosureInterpretedAndNative(t *testing.T) {
	reader, err := os.ReadFile("../../examples/record-origin-reader/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("../../examples/loop-carried-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	r, err := InspectTypedValueOrigins("carry.gooo", source, "Shift")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	programs := map[string]Result{}
	prepared := map[string]*FixedProgram{}
	for _, name := range []string{"Begin", "Step", "Finish"} {
		programs[name], err = Generate("reader.gooo", reader, name)
		if err != nil {
			t.Fatal(name, err)
		}
		prepared[name], err = NewFixedProgram(context.Background(), "reader.gooo", reader, name)
		if err != nil {
			t.Fatal(err)
		}
	}
	cases := []readerProbeCase{}
	for _, mode := range [][2]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
		graph := readerProbeGraph(t, r, r.Outputs[0], mode[0], mode[1])
		first, err := prepared["Begin"].Evaluate(context.Background(), []byte(readerProbeInputs(graph)))
		if err != nil {
			t.Fatal(err)
		}
		cursor, state := first.Value, readerProbeObserve(t, first.Value)
		for state.Status == "MORE" {
			if state.Steps >= state.Budget {
				t.Fatal("source recurrence did not terminate", mode, state)
			}
			next, err := prepared["Step"].Evaluate(context.Background(), []byte(readerProbeInputs(graph, cursor)))
			if err != nil {
				t.Fatal(err)
			}
			observed := readerProbeObserve(t, next.Value)
			if observed.Steps != state.Steps+1 || observed.Budget != state.Budget {
				t.Fatal("recurrence changed the source-owned work plan", observed, state)
			}
			cursor, state = next.Value, observed
		}
		final, err := prepared["Finish"].Evaluate(context.Background(), []byte(readerProbeInputs(cursor)))
		if err != nil {
			t.Fatal(err)
		}
		summary := readerProbeObserve(t, final.Value)
		if summary.Status != "RESOLVED" || summary.Paths != "input:urn:carryprobe:left;input:urn:carryprobe:right;input:urn:carryprobe:items;" {
			t.Fatal("mutual recurrence lost an authored input", mode, summary)
		}
		seen := readerProbeAncestors(r, r.Outputs[0].Root, mode[0], mode[1])
		words := make([]int64, 16)
		for i, selected := range seen[1:] {
			if selected {
				words[i/32] |= int64(1) << (i % 32)
			}
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal(final.Value, &object); err != nil {
			t.Fatal(err)
		}
		if got := readerMaskWords(t, object["nodes"]); !reflect.DeepEqual(got, words) {
			t.Fatal("source cycle closure disagrees with independent original-node traversal", mode, got, words)
		}
		cases = append(cases, readerProbeCase{Name: fmt.Sprintf("control=%v-history=%v", mode[0], mode[1]),
			Graph: graph, Paths: readerProbeOracle(t, r, r.Outputs[0].Root, mode[0], mode[1]), Words: words})
		t.Logf("control=%v history=%v: 3 inputs; %d nodes; %d/%d source steps; static call closure %d/4096",
			mode[0], mode[1], len(r.Flow.Nodes), state.Steps, state.Budget, programs["Step"].Report.CallClosure.MaxCallsPerInvocation)
	}
	directory := t.TempDir()
	raw, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "native-cases.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	readerProbeNative(t, programs, directory)
}
