package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func sourceIndexProgram(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/indexed-record-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestSourceIndexCarriesCollectionAndAuthoredIndex(t *testing.T) {
	r, err := InspectTypedValueOrigins("index.gooo", sourceIndexProgram(t), "Read")
	if err != nil || r.Status != "RESOLVED" || r.Schema != "gooo/typed-value-origins/v3" {
		t.Fatal("source index unresolved", err, r.Reason)
	}
	if len(r.Flow.Loops) != 0 || len(r.Flow.LoopCarries) != 0 {
		t.Fatal("direct index invented a loop", r.Flow.Loops, r.Flow.LoopCarries)
	}
	seen := readerProbeAncestors(r, r.Outputs[0].Root, false, false)
	fields := []string{}
	for _, input := range r.Inputs {
		if seen[input.Node] {
			fields = append(fields, input.FieldID)
		}
	}
	slices.Sort(fields)
	if !reflect.DeepEqual(fields, []string{"urn:index:items", "urn:index:offset"}) {
		t.Fatal("collection, index expression or unused field origin changed", fields)
	}
	indices, selected, ignored := 0, false, false
	for i, node := range r.Flow.Nodes {
		if node.Kind == "element" && node.Operator == "source_index" {
			indices++
			if node.Parents[0] == 0 || node.Parents[1] == 0 || r.Sources[i].Status != "MAPPED" ||
				r.Sources[i].Expression != "items[커서]" {
				t.Fatal("authored index lost original Korean expression or either dependency", node, r.Sources[i])
			}
		}
		selected = selected || seen[node.ID] && node.Kind == "element_field" && node.FieldID == "urn:index:value"
		ignored = ignored || seen[node.ID] && node.Kind == "element_field" && node.FieldID == "urn:index:ignored"
	}
	if indices != 1 || !selected || ignored {
		t.Fatal("index collapsed field identity", indices, selected, ignored)
	}
	if _, err := TypedValueOriginInputs(r, "index:Inputs", "", false); err == nil {
		t.Fatal("old columns silently dropped indexed structure")
	}
}

func TestSourceIndexKeepsExistingExecutableBehavior(t *testing.T) {
	p, err := NewFixedProgram(context.Background(), "index.gooo", sourceIndexProgram(t), "Read")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct{ input, want string }{
		{`[{"items":[{"value":7,"ignored":91},{"value":11,"ignored":92}],"offset":-1,"unused":99}]`, "7"},
		{`[{"items":[{"value":7,"ignored":91},{"value":11,"ignored":92}],"offset":0,"unused":99}]`, "11"},
	} {
		got, err := p.Evaluate(context.Background(), []byte(row.input))
		if err != nil || strings.TrimSpace(string(got.Value)) != row.want {
			t.Fatal(got, err, row.want)
		}
	}
}

func TestSourceIndexVariantsKeepFieldAndIndexIdentity(t *testing.T) {
	source := string(sourceIndexProgram(t))
	start, end := strings.Index(source, "`"), strings.LastIndex(source, "`")
	for _, row := range []struct {
		name, body string
		inputs     []string
		scalarList bool
	}{
		{"literal", "return input.items[0].value", []string{"urn:index:items"}, false},
		{"arithmetic", "return input.items[input.offset+1].value", []string{"urn:index:items", "urn:index:offset"}, false},
		{"scalar", "return input.items[input.offset+1]", []string{"urn:index:items", "urn:index:offset"}, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			candidate := source[:start+1] + row.body + source[end:]
			if row.scalarList {
				candidate = strings.Replace(candidate, "type Item required many", "type integer required many", 1)
			}
			r, err := InspectTypedValueOrigins("index-variant.gooo", []byte(candidate), "Read")
			if err != nil || r.Status != "RESOLVED" {
				t.Fatal(err, r.Reason)
			}
			seen := readerProbeAncestors(r, r.Outputs[0].Root, false, false)
			fields := []string{}
			for _, input := range r.Inputs {
				if seen[input.Node] {
					fields = append(fields, input.FieldID)
				}
			}
			slices.Sort(fields)
			if !reflect.DeepEqual(fields, row.inputs) {
				t.Fatal("literal or computed index changed its source inputs", fields, row.inputs)
			}
		})
	}
}

func TestSourceIndexRuntimeBoundsRemainIndependent(t *testing.T) {
	p, err := NewFixedProgram(context.Background(), "index.gooo", sourceIndexProgram(t), "Read")
	if err != nil {
		t.Fatal(err)
	}
	for _, offset := range []string{"-2", "1"} {
		input := `[{"items":[{"value":7,"ignored":91},{"value":11,"ignored":92}],"offset":` + offset + `,"unused":99}]`
		got, err := p.Evaluate(context.Background(), []byte(input))
		if err == nil || len(got.Value) != 0 {
			t.Fatal("static source-index support accepted a runtime out-of-range value", got, err)
		}
	}
}

func TestSourceIndexOverflowPublishesNoAcceptedPrefix(t *testing.T) {
	source := string(sourceIndexProgram(t))
	start, end := strings.Index(source, "`"), strings.LastIndex(source, "`")
	body := "let total = 0\n" + strings.Repeat("total = total + input.items[input.offset].value\n", 150) + "return total"
	candidate := source[:start+1] + body + source[end:]
	r, err := InspectTypedValueOrigins("index-overflow.gooo", []byte(candidate), "Read")
	if err != nil || r.Status != "UNRESOLVED" || r.Reason != "FLOW_NODE_BOUND" ||
		len(r.Outputs) != 0 || len(r.Flow.Nodes) != 0 || len(r.Flow.RecordVersions) != 0 {
		t.Fatal("index overflow emitted accepted partial source origins", err, r.Reason)
	}
}

func TestSourceIndexQueryGeneratedNativeVisitsOriginalNodes(t *testing.T) {
	r, err := InspectTypedValueOrigins("index.gooo", sourceIndexProgram(t), "Read")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal(err, r.Reason)
	}
	reader, err := os.ReadFile("../../examples/record-origin-reader/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	programs := map[string]Result{}
	for _, name := range []string{"Begin", "Step", "Finish"} {
		programs[name], err = Generate("reader.gooo", reader, name)
		if err != nil {
			t.Fatal(err)
		}
	}
	seen := readerProbeAncestors(r, r.Outputs[0].Root, false, false)
	words := make([]int64, 16)
	for i, selected := range seen[1:] {
		if selected {
			words[i/32] |= int64(1) << (i % 32)
		}
	}
	cases := []readerProbeCase{{Name: "authored-index", Graph: readerProbeGraph(t, r, r.Outputs[0], false, false),
		Paths: []string{"input:urn:index:items", "input:urn:index:offset"}, Words: words}}
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
