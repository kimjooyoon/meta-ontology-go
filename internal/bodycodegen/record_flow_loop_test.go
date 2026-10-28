package bodycodegen

import (
	"bytes"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unsafe"
)

func TestSymbolicCrossCarrySourceClosure(t *testing.T) {
	source, err := os.ReadFile("../../examples/loop-carried-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	r, err := InspectTypedValueOrigins("carry.gooo", source, "Shift")
	if err != nil || r.Status != "RESOLVED" || r.Schema != "gooo/typed-value-origins/v3" {
		t.Fatal("symbolic collection range unresolved", err, r.Reason)
	}
	if len(r.Flow.Loops) != 1 || len(r.Flow.LoopCarries) != 2 || r.Flow.Loops[0].IndexKind != "symbolic_any" {
		t.Fatal("loop and cross-local carry identities missing", r.Flow.Loops, r.Flow.LoopCarries)
	}
	seen := symbolicLoopAncestors(r, r.Outputs[0].Root)
	var inputs, fields []string
	for _, input := range r.Inputs {
		if seen[input.Node] {
			inputs = append(inputs, input.FieldID)
		}
	}
	for _, n := range r.Flow.Nodes {
		if n.Parents[0] >= n.ID || n.Parents[1] >= n.ID {
			t.Fatal("recurrence escaped into ordinary topological parents", n)
		}
		if seen[n.ID] && n.Kind == "element_field" {
			fields = append(fields, n.FieldID)
		}
	}
	slices.Sort(inputs)
	if !reflect.DeepEqual(inputs, []string{"urn:carryprobe:items", "urn:carryprobe:left", "urn:carryprobe:right"}) ||
		!reflect.DeepEqual(fields, []string{"urn:carryprobe:value"}) {
		t.Fatal("mutual recurrence lost seed or selected field, or included unused field", inputs, fields)
	}
	if _, err := TypedValueOriginInputs(r, "loop:Inputs", "", false); err == nil {
		t.Fatal("legacy flat consumer silently dropped recurrence")
	}
	for i, mapped := range r.Sources {
		if r.Flow.Nodes[i].Span.View != "input-port" && mapped.Status != "MAPPED" {
			t.Fatal("loop source annotation escaped original source", r.Flow.Nodes[i], mapped)
		}
	}
	var text bytes.Buffer
	if err := WriteTypedOriginText(&text, r, source); err != nil || !strings.Contains(text.String(), "index=symbolic_any") ||
		strings.Count(text.String(), "source recurrence #") != 2 {
		t.Fatal("human projection lost explicit recurrence or source identity", err, text.String())
	}
	t.Logf("symbolic loop %d nodes; 2 carry relations; both original seeds and used item field retained; actual indices absent; no model/native execution by source inspection", len(r.Flow.Nodes))
}

func TestSymbolicLoopConcurrentIndependentGraphs(t *testing.T) {
	source, err := os.ReadFile("../../examples/loop-carried-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := InspectTypedValueOrigins("parallel.gooo", source, "Shift")
	if err != nil || baseline.Status != "RESOLVED" {
		t.Fatal(err, baseline.Reason)
	}
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			for range 4 {
				got, err := InspectTypedValueOrigins("parallel.gooo", source, "Shift")
				if err != nil || !reflect.DeepEqual(got, baseline) {
					t.Error("concurrent source inspection shared a mutable loop frame", err)
					return
				}
				got.Flow.LoopCarries[0].Yield = 0
			}
		})
	}
	workers.Wait()
	if baseline.Flow.LoopCarries[0].Yield == 0 {
		t.Fatal("returned recurrence tables alias another inspection")
	}
}

func TestSymbolicLoopHelperKeepsCallerInputsAndCalleeSource(t *testing.T) {
	source, err := os.ReadFile("../../examples/loop-carried-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	source = append(source, []byte("\nactivity Forward(State) -> Integer id \"urn:carryprobe:forward\" computes `return Shift(input)`\n")...)
	r, err := InspectTypedValueOrigins("helper-loop.gooo", source, "Forward")
	if err != nil || r.Status != "RESOLVED" || len(r.Flow.Helpers) != 1 || len(r.Flow.Loops) != 1 ||
		r.Flow.Loops[0].Span.ActivityID != "urn:carryprobe:shift" {
		t.Fatal("helper frame lost loop ownership", err, r.Reason, r.Flow.Loops)
	}
	seen := symbolicLoopAncestors(r, r.Outputs[0].Root)
	inputs := 0
	for _, input := range r.Inputs {
		if seen[input.Node] {
			if input.Declaration == nil || input.Declaration.ActivityID != "urn:carryprobe:forward" {
				t.Fatal("helper parameter became a fabricated external input", input)
			}
			inputs++
		}
	}
	if inputs != 3 {
		t.Fatal("caller input roots did not survive helper recurrence", inputs)
	}
	for i, node := range r.Flow.Nodes {
		if node.Span.ActivityID == "urn:carryprobe:shift" && node.Span.View == "computes" &&
			r.Sources[i].Status != "MAPPED" {
			t.Fatal("callee loop offsets escaped its original carrier", node, r.Sources[i])
		}
	}
}

func TestSymbolicNestedCarries(t *testing.T) {
	source := []byte(`package nestedloop
namespace nestedloop
entity Integer id "urn:gooo:type:integer"
entity Item id "urn:nested:item" fields {
 field value id "urn:nested:value" type integer required one
 field ignored id "urn:nested:ignored" type integer required one
}
entity State id "urn:nested:state" fields {
 field seed id "urn:nested:seed" type integer required one
 field first id "urn:nested:first" type Item required many
 field second id "urn:nested:second" type Item required many
}
activity Sum(State) -> Integer id "urn:nested:sum" computes ` + "`" + `
 let total = input.seed
 for a in input.first {
  for b in input.second { total = total + a.value*b.value }
 }
 return total
` + "`" + `
`)
	r, err := InspectTypedValueOrigins("nested.gooo", source, "Sum")
	if err != nil || r.Status != "RESOLVED" || len(r.Flow.Loops) != 2 || r.Flow.Loops[1].Parent != 1 ||
		len(r.Flow.LoopCarries) != 2 {
		t.Fatal("nested symbolic loop ownership or carries", err, r.Reason, r.Flow.Loops, r.Flow.LoopCarries)
	}
	seen := symbolicLoopAncestors(r, r.Outputs[0].Root)
	fields := map[string]bool{}
	for _, input := range r.Inputs {
		if seen[input.Node] {
			fields[input.FieldID] = true
		}
	}
	if len(fields) != 3 || !fields["urn:nested:seed"] || !fields["urn:nested:first"] || !fields["urn:nested:second"] {
		t.Fatal("nested recurrence lost an original seed or collection", fields)
	}
}

func TestSymbolicNominalCarryPreservesUntouchedField(t *testing.T) {
	source := []byte(`package nominalcarry
namespace nominalcarry
entity Integer id "urn:gooo:type:integer"
entity Item id "urn:nominal:item" fields {
 field value id "urn:nominal:item-value" type integer required one
}
entity Pair id "urn:nominal:pair" fields {
 field value id "urn:nominal:value" type integer required one
 field label id "urn:nominal:label" type string required one
}
entity State id "urn:nominal:state" fields {
 field saved id "urn:nominal:saved" type Pair required one
 field items id "urn:nominal:items" type Item required many
}
activity Update(State) -> Pair id "urn:nominal:update" computes ` + "`" + `
 let current = input.saved
 for item in input.items { current.value = item.value }
 return current
` + "`" + `
`)
	r, err := InspectTypedValueOrigins("nominal.gooo", source, "Update")
	if err != nil || r.Status != "RESOLVED" || len(r.Flow.RecordVersions) == 0 {
		t.Fatal("nominal loop carry unresolved", err, r.Reason)
	}
	for _, output := range r.Outputs {
		seen := symbolicLoopAncestors(r, output.Root)
		items, label := false, false
		for _, input := range r.Inputs {
			items = items || seen[input.Node] && input.FieldID == "urn:nominal:items"
			label = label || seen[input.Node] && input.FieldID == "urn:nominal:label"
		}
		if output.FieldID == "urn:nominal:label" && (items || !label) ||
			output.FieldID == "urn:nominal:value" && (!items || label) {
			t.Fatal("field-sensitive carry mixed untouched and replaced fields", output, items, label)
		}
	}
}

func TestSymbolicFirstElementKeepsZeroIterationExit(t *testing.T) {
	source := []byte(`package first
namespace first
entity Integer id "urn:gooo:type:integer"
entity State id "urn:first:state" fields {
 field seed id "urn:first:seed" type integer required one
 field items id "urn:first:items" type integer required many
}
activity First(State) -> Integer id "urn:first:first" computes ` + "`" + `
 for item in input.items { return item }
 return input.seed
` + "`" + `
`)
	r, err := InspectTypedValueOrigins("first.gooo", source, "First")
	if err != nil || r.Status != "RESOLVED" || len(r.Flow.Loops) != 1 || r.Flow.Loops[0].Continues {
		t.Fatal("returning range lost its zero-iteration continuation", err, r.Reason, r.Flow.Loops)
	}
	seen := symbolicLoopAncestors(r, r.Outputs[0].Root)
	count := 0
	for _, input := range r.Inputs {
		if seen[input.Node] {
			count++
		}
	}
	if count != 2 {
		t.Fatal("first element or zero-iteration seed disappeared", count)
	}
	if _, err := TypedSourceControlInputs(r, "loop:Inputs"); err == nil {
		t.Fatal("legacy control columns silently dropped scalar recurrence metadata")
	}
	for _, model := range []*RecordOrdinalContext{recordOriginContext(nil, r.Flow), recordGraphContext(recordAssemblyPlan{}, r.Flow)} {
		if model.Status != "DECLINED_TO_DETERMINISTIC" || model.Text != "" || len(model.Parts) != 0 {
			t.Fatal("legacy model accepted an incomplete loop context", model)
		}
	}
}

func TestSymbolicLoopMapsEscapedOriginalCarrier(t *testing.T) {
	source, err := os.ReadFile("../../examples/loop-carried-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	s := strings.ReplaceAll(string(source), "previous", "이전")
	start, end := strings.Index(s, "`"), strings.LastIndex(s, "`")
	s = s[:start] + strconv.QuoteToASCII(s[start+1:end]) + s[end+1:]
	r, err := InspectTypedValueOrigins("escaped-loop.gooo", []byte(s), "Shift")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal("escaped symbolic loop unresolved", err, r.Reason)
	}
	matched := 0
	for _, source := range r.Sources {
		if source.Expression == "이전" {
			if source.Status != "MAPPED" || source.SourceSHA256 != digest([]byte(s)) ||
				s[source.Start:source.End] != `\uc774\uc804` {
				t.Fatal("normalized range did not map to original Unicode escape bytes", source)
			}
			matched++
		}
	}
	if matched != 1 {
		t.Fatal("original escaped local read identity missing", matched)
	}
}

func TestSymbolicLoopNodeOverflowRetainsNoAcceptedPrefix(t *testing.T) {
	source, err := os.ReadFile("../../examples/loop-carried-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.Replace(string(source), "right = previous + item.value",
		strings.Repeat("right = previous + item.value\n", 150), 1))
	r, err := InspectTypedValueOrigins("overflow.gooo", source, "Shift")
	if err != nil || r.Status != "UNRESOLVED" || r.Reason != "FLOW_NODE_BOUND" || len(r.Outputs) != 0 ||
		len(r.Flow.Nodes) != 0 || len(r.Flow.LoopCarries) != 0 || len(r.Flow.RecordVersions) != 0 {
		t.Fatal("node limit emitted an accepted symbolic prefix", err, r.Reason)
	}
}

func TestSymbolicSequenceTagFitsExistingFlowValuePadding(t *testing.T) {
	var before struct {
		record int
		scalar uint16
		fields [RecordFieldLimit]uint16
	}
	if unsafe.Sizeof(flowValue{}) != unsafe.Sizeof(before) {
		t.Fatal("sequence discriminator grew every existing flow value", unsafe.Sizeof(flowValue{}), unsafe.Sizeof(before))
	}
}

func symbolicLoopAncestors(r TypedValueOrigins, root uint16) []bool {
	seen := make([]bool, len(r.Flow.Nodes)+1)
	pending := []uint16{root}
	for len(pending) != 0 {
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		n := r.Flow.Nodes[id-1]
		pending = append(pending, n.Parents[0], n.Parents[1])
		for _, carry := range r.Flow.LoopCarries {
			if carry.Header == id {
				pending = append(pending, carry.Yield)
			}
		}
	}
	return seen
}
