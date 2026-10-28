package bodycodegen

import (
	"context"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unsafe"
)

func constructedSequenceSource(t *testing.T, body string) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/constructed-record-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		text := string(source)
		first, last := strings.Index(text, "`"), strings.LastIndex(text, "`")
		source = []byte(text[:first+1] + body + text[last:])
	}
	return source
}

func constructedSequenceInputs(t *testing.T, source []byte) TypedValueOrigins {
	t.Helper()
	r, err := InspectTypedValueOrigins("constructed.gooo", source, "Read")
	if err != nil || r.Status != "RESOLVED" {
		t.Fatal("constructed source unresolved", err, r.Reason)
	}
	return r
}

func assertConstructedContributors(t *testing.T, source []byte, control bool, want ...string) {
	t.Helper()
	r := constructedSequenceInputs(t, source)
	if _, err := TypedValueOriginInputs(r, "construct:Inputs", "", false); err == nil {
		t.Fatal("old flat model input silently erased constructed item structure")
	}
	seen := readerProbeAncestors(r, r.Outputs[0].Root, control, false)
	got := []string{}
	for _, input := range r.Inputs {
		if seen[input.Node] {
			got = append(got, strings.TrimPrefix(input.FieldID, "urn:construct:"))
		}
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatal("constructed item field acquired irrelevant data or lost an origin", got, want)
	}
}

func TestConstructedSequenceFieldSpecificSourceOrigins(t *testing.T) {
	source := constructedSequenceSource(t, "")
	assertConstructedContributors(t, source, false, "left", "pick", "right")
	literal := []byte(strings.Replace(string(source), "items[input.pick]", "items[0]", 1))
	assertConstructedContributors(t, literal, false, "left")
	alias := []byte(strings.Replace(string(source), "return items[input.pick]", "let 선택 = items\n return 선택[input.pick]", 1))
	assertConstructedContributors(t, alias, false, "left", "pick", "right")
	aliased := constructedSequenceInputs(t, alias)
	for i, node := range aliased.Flow.Nodes {
		if node.Operator == "source_index" && (aliased.Sources[i].Status != "MAPPED" || aliased.Sources[i].Expression != "선택[input.pick]") {
			t.Fatal("constructed index lost original Korean byte positions", aliased.Sources[i])
		}
	}
	r := constructedSequenceInputs(t, source)
	positions := []string{}
	for _, node := range r.Flow.Nodes {
		if strings.HasPrefix(node.Operator, "sequence_item:") {
			positions = append(positions, node.Operator)
		}
	}
	if !reflect.DeepEqual(positions, []string{"sequence_item:0", "sequence_item:1"}) {
		t.Fatal("authored item positions collapsed", positions)
	}
}

func TestConstructedSequenceCopiesStorageAndConditionalJoins(t *testing.T) {
	source := string(constructedSequenceSource(t, ""))
	box := "entity Box id \"urn:construct:box\" fields { field items id \"urn:construct:stored\" type Item required many }\n"
	stored := strings.Replace(source, "activity Read", box+"activity Read", 1)
	stored = strings.Replace(stored, "return items[input.pick].value", "let box = Box{items: items}\n let copy = box\n return copy.items[input.pick].value", 1)
	assertConstructedContributors(t, []byte(stored), false, "left", "pick", "right")
	joined := strings.Replace(source, "let items = []Item{first, second}",
		"let items = []Item{first}\n if input.pick == 0 { items = []Item{first} } else { items = []Item{second} }", 1)
	joined = strings.Replace(joined, "items[input.pick]", "items[0]", 1)
	assertConstructedContributors(t, []byte(joined), false, "left", "right")
	assertConstructedContributors(t, []byte(joined), true, "left", "right", "pick")
	storedJoin := strings.Replace(stored, "let copy = box", "if input.pick == 0 { box.items = []Item{first} } else { box.items = []Item{second} }\n let copy = box", 1)
	storedJoin = strings.Replace(storedJoin, "copy.items[input.pick]", "copy.items[0]", 1)
	assertConstructedContributors(t, []byte(storedJoin), false, "left", "right")
	assertConstructedContributors(t, []byte(storedJoin), true, "left", "right", "pick")
}

func TestConstructedSequenceNestedFieldsKeepIndexOrigin(t *testing.T) {
	source := string(constructedSequenceSource(t, ""))
	source = strings.Replace(source, "entity Item", "entity Detail id \"urn:construct:detail\" fields { field amount id \"urn:construct:amount\" type integer required one }\nentity Item", 1)
	source = strings.Replace(source, "urn:construct:value\" type integer", "urn:construct:value\" type Detail", 1)
	source = strings.ReplaceAll(source, "value: input.left", "value: Detail{amount: input.left}")
	source = strings.ReplaceAll(source, "value: input.right", "value: Detail{amount: input.right}")
	source = strings.Replace(source, "items[input.pick].value", "items[input.pick].value.amount", 1)
	assertConstructedContributors(t, []byte(source), false, "left", "pick", "right")
}

func TestConstructedSequenceWholeValuesAndMixedInput(t *testing.T) {
	source := string(constructedSequenceSource(t, ""))
	box := "entity Box id \"urn:construct:box\" fields { field items id \"urn:construct:stored\" type Item required many }\n"
	whole := strings.Replace(source, "activity Read", box+"activity Read", 1)
	whole = strings.Replace(whole, "-> Integer", "-> Box", 1)
	whole = strings.Replace(whole, "return items[input.pick].value", "return Box{items: items}", 1)
	assertConstructedContributors(t, []byte(whole), false, "left", "right", "noise")
	mixed := strings.Replace(source, "entity State id \"urn:construct:state\" fields {", "entity State id \"urn:construct:state\" fields { field items id \"urn:construct:items\" type Item required many", 1)
	mixed = strings.Replace(mixed, "let items = []Item{first, second}", "let items = []Item{first}\n if input.pick == 0 { items = input.items }", 1)
	mixed = strings.Replace(mixed, "items[input.pick]", "items[0]", 1)
	assertConstructedContributors(t, []byte(mixed), false, "left", "items")
	assertConstructedContributors(t, []byte(mixed), true, "left", "items", "pick")
}

func TestConstructedSequenceScalarLiterals(t *testing.T) {
	source := constructedSequenceSource(t, "let items = []int64{input.left, input.right}\n return items[input.pick]")
	assertConstructedContributors(t, source, false, "left", "right", "pick")
	constant := []byte(strings.Replace(string(source), "items[input.pick]", "items[int64(0)]", 1))
	assertConstructedContributors(t, constant, false, "left")
	loop := constructedSequenceSource(t, "let items = []int64{input.left, input.right}\n let total = 0\n for item in items { total = total + item }\n return total")
	assertConstructedContributors(t, loop, true, "left", "right")
	r := constructedSequenceInputs(t, source)
	indices := 0
	for _, node := range r.Flow.Nodes {
		if node.Operator == "source_index" {
			indices++
			if node.Kind != "element" {
				t.Fatal("source index lost its collection/index node roles", node)
			}
		}
	}
	if indices != 1 {
		t.Fatal("scalar list invented additional authored index expressions", indices)
	}
}

func TestConstructedSequenceLoopAndEmptyLiteral(t *testing.T) {
	source := string(constructedSequenceSource(t, ""))
	loop := strings.Replace(source, "return items[input.pick].value", "let total = 0\n for item in items { total = total + item.value }\n return total", 1)
	assertConstructedContributors(t, []byte(loop), false, "left", "right")
	assertConstructedContributors(t, []byte(loop), true, "left", "right")
	empty := strings.Replace(loop, "[]Item{first, second}", "[]Item{}", 1)
	assertConstructedContributors(t, []byte(empty), false)
	for _, text := range []string{source, loop, empty} {
		p, err := NewFixedProgram(context.Background(), "constructed.gooo", []byte(text), "Read")
		if err != nil {
			t.Fatal(err)
		}
		want := "7"
		if text == loop {
			want = "18"
		} else if text == empty {
			want = "0"
		}
		got, err := p.Evaluate(context.Background(), []byte(`[{"left":7,"right":11,"noise":99,"pick":0}]`))
		if err != nil || strings.TrimSpace(string(got.Value)) != want {
			t.Fatal("existing execution changed", got, err, want)
		}
	}
}

func TestConstructedSequenceBoundariesPublishNoPrefix(t *testing.T) {
	source := string(constructedSequenceSource(t, ""))
	late := strings.Replace(source, "entity State id \"urn:construct:state\" fields {", "entity State id \"urn:construct:state\" fields { field items id \"urn:construct:items\" type Item required many", 1)
	late = strings.Replace(late, "let items = []Item{first, second}", "let items = input.items", 1)
	late = strings.Replace(late, "return items[input.pick].value", "for item in []Item{first} { items = []Item{item} }\n return items[0].value", 1)
	for _, row := range []struct{ source, reason string }{
		{strings.Replace(source, "[]Item{first, second}", "[]Item{}", 1), "FLOW_SEQUENCE_LITERAL_INDEX_BOUND"},
		{strings.Replace(source, "[]Item{first, second}", "[]Item{"+strings.Repeat("first,", 180)+"}", 1), "FLOW_NODE_BOUND"},
		{strings.Replace(source, "return items[input.pick].value", "for item in []Item{second} { items = []Item{item} }\n return items[0].value", 1), "FLOW_CONSTRUCTED_SEQUENCE_RECURRENCE_UNRESOLVED"},
		{late, "FLOW_CONSTRUCTED_SEQUENCE_RECURRENCE_UNRESOLVED"},
	} {
		r, err := InspectTypedValueOrigins("constructed-bound.gooo", []byte(row.source), "Read")
		if err != nil || r.Status != "UNRESOLVED" || r.Reason != row.reason || len(r.Flow.Nodes) != 0 || len(r.Outputs) != 0 {
			t.Fatal("constructed limit emitted accepted prefix or lost its reason", err, r.Reason, row.reason)
		}
		if row.reason == "FLOW_NODE_BOUND" {
			if _, err := Generate("constructed-bound.gooo", []byte(row.source), "Read"); err != nil {
				t.Fatal("optional origin limit blocked existing code generation", err)
			}
		}
	}
	if unsafe.Sizeof(flowSequenceOrigin{}) > 16 {
		t.Fatal("sequence-origin index exceeded its fixed compact layout")
	}
}
