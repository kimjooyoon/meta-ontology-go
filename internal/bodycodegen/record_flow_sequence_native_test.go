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
	rows := []struct {
		name, source string
		paths        []string
	}{
		{"computed", source, []string{"input:urn:construct:left", "input:urn:construct:pick", "input:urn:construct:right"}},
		{"literal", strings.Replace(source, "items[input.pick]", "items[0]", 1), []string{"input:urn:construct:left"}},
		{"loop", strings.Replace(source, "return items[input.pick].value", "let total = 0\n for item in items { total = total + item.value }\n return total", 1), []string{"input:urn:construct:left", "input:urn:construct:right"}},
		{"scalar-loop", string(constructedSequenceSource(t, "let items = []int64{input.left, input.right}\n let total = 0\n for item in items { total = total + item }\n return total")), []string{"input:urn:construct:left", "input:urn:construct:right"}},
	}
	cases := []readerProbeCase{}
	for _, row := range rows {
		r := constructedSequenceInputs(t, []byte(row.source))
		seen := readerProbeAncestors(r, r.Outputs[0].Root, false, false)
		words := make([]int64, 16)
		for i, selected := range seen[1:] {
			if selected {
				words[i/32] |= int64(1) << (i % 32)
			}
		}
		cases = append(cases, readerProbeCase{Name: row.name, Graph: readerProbeGraph(t, r, r.Outputs[0], false, false),
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
