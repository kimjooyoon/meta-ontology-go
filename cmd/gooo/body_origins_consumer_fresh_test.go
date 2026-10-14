package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const source13057 = `package originconsumer
namespace originconsumer
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
entity Pair id "urn:typed13057:pair" fields {
field base id "urn:typed13057:base" type integer required one
field next id "urn:typed13057:next" type integer required one
}
activity Carry(Pair) -> Pair id "urn:typed13057:carry" computes ` + "`let copied=input;copied.base=input.base+1543;return copied`" + `
activity Read(Pair) -> Integer id "urn:typed13057:read" computes ` + "`return input.next`" + `
activity Build(Pair,Integer) -> Integer id "urn:typed13057:build" computes ` + "`let adjustment=1549;let staged=Carry(input0);staged.base=input1+adjustment;return Read(staged)`" + `
activity Switch(Pair,Integer,Boolean) -> Integer id "urn:typed13057:switch" computes ` + "`let staged=input0;if input2 {staged.next=input1+1553};return Read(staged)`" + `
`

func TestTypedAncestryConsumer13057(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	sourceRoot := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if sourceRoot == "" {
		sourceRoot = "../.."
	}
	root := t.TempDir()
	write := func(name string, raw []byte) string {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	call := func(name string, args ...string) []byte {
		t.Helper()
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK {
			t.Fatal(name, code, diagnostics.String(), out.String())
		}
		if dir := os.Getenv("GOOO_TYPED_OUTPUT"); dir != "" {
			var compressed bytes.Buffer
			z := gzip.NewWriter(&compressed)
			if _, err := z.Write(out.Bytes()); err != nil {
				t.Fatal(err)
			}
			if err := z.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name+".json.gz"), compressed.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
		}
		return append([]byte(nil), out.Bytes()...)
	}
	source := write("main13057.gooo", []byte(source13057))
	manifest := write("gooo.workspace.json", []byte(`{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"fresh/origins","activity":"Build"},"packages":[{"path":"fresh/origins","name":"originconsumer","imports":[],"sources":["main13057.gooo"]}]}`))
	cases := write("cases13057.json", []byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"fresh/origins:Build.input0":{"base":1559,"next":1567},"fresh/origins:Build.input1":1571},"expected":{"fresh/origins:Build":1567}},{"inputs":{"fresh/origins:Build.input0":{"base":-1579,"next":-1583},"fresh/origins:Build.input1":-1597},"expected":{"fresh/origins:Build":-1583}}]}`))
	producer := call("producer13057", "package", "execute", "--json", "--cases", cases, manifest)
	var produced packageExecutionReceipt
	if err := json.Unmarshal(producer, &produced); err != nil || produced.Result == nil || produced.Result.Runtime.FinitePassed != 2 || produced.Result.Runtime.ModelCalls != 0 {
		t.Fatal("fresh helper native behavior", err)
	}
	var rows []map[string]json.RawMessage
	for i, activity := range []string{"Build", "Switch", "Switch"} {
		args := []string{"body-origins", "--activity", activity, "--inputs-for", "typed/ancestry:Inputs"}
		if i == 2 {
			args = append(args, "--include-control")
		}
		raw := call([]string{"build13057", "switch-data13057", "switch-control13057"}[i], append(args, source)...)
		var input struct {
			Inputs []map[string]json.RawMessage `json:"inputs"`
		}
		if err := json.Unmarshal(raw, &input); err != nil || len(input.Inputs) != 1 {
			t.Fatal(err)
		}
		rows = append(rows, input.Inputs[0])
	}
	raw, _ := json.Marshal(struct {
		Schema string                       `json:"schema"`
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}{"gooo/body-composition-inputs/v1", rows})
	inputs := write("ancestry13057.json", raw)
	consumer := call("consumer13057", "package", "execute", "--json", "--inputs", inputs, filepath.Join(sourceRoot, "examples/typed-value-origins/ancestry.workspace.json"))
	var observed packageExecutionReceipt
	if err := json.Unmarshal(consumer, &observed); err != nil || observed.Result == nil || observed.Decision != "OBSERVED" || observed.Result.Runtime.FiniteTotal != 0 || observed.Result.Runtime.ModelCalls != 0 || len(observed.Result.Runtime.Traces) != 3 {
		t.Fatal("Gooo input-only observer scope", err)
	}
	want := []string{"input0:urn:typed13057:next;", "input0:urn:typed13057:next;input1:;", "input0:urn:typed13057:next;input1:;input2:;"}
	for i, trace := range observed.Result.Runtime.Traces {
		var actual string
		if len(trace.Deliveries) != 1 || json.Unmarshal(trace.Deliveries[0].Actual, &actual) != nil || actual != want[i] {
			t.Fatal("Gooo field ancestry result", i, actual, want[i])
		}
	}
	t.Log("fresh record-returning-helper producer2/2; real Gooo bitset traversal3/3 external comparisons, input-only finite0, models0")
}
