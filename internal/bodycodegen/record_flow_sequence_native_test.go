package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConstructedSequenceIndependentNativeValues(t *testing.T) {
	source := constructedSequenceSource(t, "")
	result, err := Generate("constructed.gooo", source, "Read")
	if err != nil {
		t.Fatal(err)
	}
	cases := [][2]string{
		{`{"left":7,"right":11,"noise":99,"pick":0}`, `7`},
		{`{"left":7,"right":11,"noise":99,"pick":1}`, `11`},
		{`{"left":7,"right":7,"noise":99,"pick":1}`, `7`},
		{`{"left":9007199254740993,"right":11,"noise":99,"pick":0}`, `9007199254740993`},
	}
	p, err := NewFixedProgram(context.Background(), "constructed.gooo", source, "Read")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range cases {
		got, err := p.Evaluate(context.Background(), []byte("["+row[0]+"]"))
		if err != nil || strings.TrimSpace(string(got.Value)) != row[1] {
			t.Fatal("independent authored value expectation failed", got, err, row[1])
		}
	}
	assertNominalNativeCases(t, result, "State", cases)
}

func TestConstructedSequenceNestedIndependentNativeValues(t *testing.T) {
	for _, row := range []struct {
		name                  string
		emptyLeft, emptyRight bool
		normal, large         string
	}{
		{"nonempty", false, false, "18", "9007199254741004"},
		{"empty-first", true, false, "11", "11"},
		{"empty-second", false, true, "7", "9007199254740993"},
		{"both-empty", true, true, "0", "0"},
	} {
		t.Run(row.name, func(t *testing.T) {
			source := []byte(nestedConstructedLoopSource(t, row.emptyLeft, row.emptyRight))
			result, err := Generate("nested.gooo", source, "Read")
			if err != nil {
				t.Fatal(err)
			}
			assertNominalNativeCases(t, result, "State", [][2]string{
				{`{"left":7,"right":11,"noise":99,"pick":0,"sub":0}`, row.normal},
				{`{"left":9007199254740993,"right":11,"noise":99,"pick":1,"sub":0}`, row.large},
			})
		})
	}
}

func TestConstructedSequenceGeneratedGoooReader(t *testing.T) {
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
	source := string(constructedSequenceSource(t, ""))
	branch := strings.Replace(nestedConstructedLoopSource(t, true, false), "let items = []Item{first, second}",
		"if input.pick != 0 { first.details = []Detail{Detail{amount: input.left}} }\n let items = []Item{first, second}", 1)
	rows := []struct {
		name, source string
		paths        []string
		control      bool
	}{
		{"computed", source, []string{"input:urn:construct:left", "input:urn:construct:pick", "input:urn:construct:right"}, false},
		{"literal", strings.Replace(source, "items[input.pick]", "items[0]", 1), []string{"input:urn:construct:left"}, false},
		{"loop", strings.Replace(source, "return items[input.pick].value", "let total = 0\n for item in items { total = total + item.value }\n return total", 1), []string{"input:urn:construct:left", "input:urn:construct:right"}, false},
		{"scalar-loop", string(constructedSequenceSource(t, "let items = []int64{input.left, input.right}\n let total = 0\n for item in items { total = total + item }\n return total")), []string{"input:urn:construct:left", "input:urn:construct:right"}, false},
		{"nested-loop", nestedConstructedLoopSource(t, false, false), []string{"input:urn:construct:left", "input:urn:construct:right"}, false},
		{"nested-empty-first", nestedConstructedLoopSource(t, true, false), []string{"input:urn:construct:right"}, false},
		{"nested-empty-second", nestedConstructedLoopSource(t, false, true), []string{"input:urn:construct:left"}, false},
		{"nested-both-empty", nestedConstructedLoopSource(t, true, true), []string{}, false},
		{"nested-empty-control", branch,
			[]string{"input:urn:construct:left", "input:urn:construct:pick", "input:urn:construct:right"}, true},
	}
	cases := []readerProbeCase{}
	for _, row := range rows {
		r := constructedSequenceInputs(t, []byte(row.source))
		seen := readerProbeAncestors(r, r.Outputs[0].Root, row.control, false)
		words := make([]int64, 16)
		for i, selected := range seen[1:] {
			if selected {
				words[i/32] |= int64(1) << (i % 32)
			}
		}
		cases = append(cases, readerProbeCase{Name: row.name, Graph: readerProbeGraph(t, r, r.Outputs[0], row.control, false),
			Paths: row.paths, Words: words})
	}
	dir := t.TempDir()
	raw, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "native-cases.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	readerProbeNative(t, programs, dir)
}
