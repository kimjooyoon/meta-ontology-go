package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

const nativeFieldSource13421 = `package fields13421
namespace fields13421
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:fields13421:input" fields {
field left id "urn:fields13421:input-left" type integer required one
field right id "urn:fields13421:input-right" type integer required one
field flip id "urn:fields13421:input-flip" type boolean required one
}
entity Pair id "urn:fields13421:pair" fields {
field left id "urn:fields13421:left" type integer required one
field right id "urn:fields13421:right" type integer required one
}
activity Copy(Pair) -> Pair id "urn:fields13421:copy" computes ` + "`let saved=input;saved.left=saved.left+1931;return saved`" + `
activity Read(Pair) -> Integer id "urn:fields13421:read" computes ` + "`return input.right`" + `
activity Build(Input) -> Integer id "urn:fields13421:build" computes ` + "`let original=Pair{left:input.left,right:input.right};let copied=Copy(original);if input.flip {copied=Pair{left:copied.right,right:copied.left}};return Read(copied)`" + `
`

func TestNativeFieldHistory13421(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, tc := range [][2]string{
		{`{"left":1949,"right":1951,"flip":false}`, `1951`},
		{`{"left":1949,"right":1951,"flip":true}`, `3880`},
		{`{"left":-1931,"right":0,"flip":false}`, `0`},
		{`{"left":-1931,"right":0,"flip":true}`, `0`},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(tc[0])}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(tc[1])}})
	}
	observation := map[string]any{"source": nativeFieldSource13421, "cases": suite}
	defer saveReadObservation(t, "field-native13421.json", observation)
	p, err := GenerateCompositionWithOptions(context.Background(), "field13421.gooo", []byte(nativeFieldSource13421), suite, CompositionOptions{EntryActivity: "Build"})
	if err != nil {
		t.Fatal(err)
	}
	observation["composition"] = p
	prefix := readObservationPrefix(p.Source)
	projection, plan, err := observedCompositionFieldProjection(p.Source, prefix, p)
	if err != nil {
		t.Fatal(err)
	}
	observation["projection"] = projection
	observation["bindings"], observation["sites"], observation["uses"] = plan.inventory.bindings, plan.sites, plan.uses
	if len(plan.sites) < 12 || len(plan.uses) < 8 {
		t.Fatal("field source inventory incomplete", len(plan.sites), len(plan.uses))
	}
	observer, err := fieldReadObserver()
	if err != nil {
		t.Fatal(err)
	}
	driver, err := observedReadDriverWithObserver(p.Driver, prefix, p, observer)
	if err != nil {
		t.Fatal(err)
	}
	observation["driver"] = driver
	root := t.TempDir()
	for name, raw := range map[string][]byte{"go.mod": []byte("module example/field13421\n\ngo 1.27.2\n"), "projection.go": []byte(projection), "driver.go": []byte(driver)} {
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(root, "native13421")
	build := exec.CommandContext(context.Background(), "go", "build", "-p=1", "-o", binary, ".")
	build.Dir = root
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatal("native field build", err, string(raw))
	}
	graph, err := prepareCompositionGraphForEntry(context.Background(), "field13421.gooo", []byte(nativeFieldSource13421), "Build")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := graph.inputRows(suite)
	if err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(rows)
	var first []byte
	var native nativeFieldFile
	for attempt := range 2 {
		cmd := exec.CommandContext(context.Background(), binary)
		cmd.Dir, cmd.Stdin = root, bytes.NewReader(input)
		output, err := cmd.Output()
		if err != nil {
			t.Fatal("native execution", err)
		}
		_, passed, err := graph.nativeTraces(output, suite)
		if err != nil || passed != 4 {
			t.Fatal("actual finite outcomes", err, passed, string(output))
		}
		raw, err := os.ReadFile(filepath.Join(root, "gooo-native-field-observation.json"))
		if err != nil {
			t.Fatal(err)
		}
		if attempt == 0 {
			first = raw
			if err := json.Unmarshal(raw, &native); err != nil {
				t.Fatal(err)
			}
		} else if !bytes.Equal(first, raw) {
			t.Fatal("field identity/sequence replay differs")
		}
	}
	observation["native"] = native
	if err := validateNativeFieldFile(native, plan, 4); err != nil {
		t.Fatal(err)
	}
	for c, want := range []string{"urn:fields13421:input-right", "urn:fields13421:input-left", "urn:fields13421:input-right", "urn:fields13421:input-left"} {
		var writes, uses [][]int
		for _, row := range native.Writes {
			if row[0] == c {
				writes = append(writes, row)
			}
		}
		for _, row := range native.Uses {
			if row[0] == c {
				uses = append(uses, row)
			}
		}
		rootWrite := -1
		for _, use := range uses {
			site := plan.uses[use[1]]
			if plan.inventory.bindings[site.Binding].ActivityID == "urn:fields13421:read" {
				rootWrite = use[3]
			}
		}
		if rootWrite < 0 {
			t.Fatal("final helper field read missing")
		}
		got := fieldRoots13421(plan, writes, uses, rootWrite)
		if !reflect.DeepEqual(got, []string{want}) {
			t.Fatal("actual per-field lineage", c, got, want)
		}
	}
	changed := native
	changed.Uses = append([][]int(nil), native.Uses...)
	last := len(changed.Uses) - 1
	changed.Uses[last] = append([]int(nil), changed.Uses[last]...)
	changed.Uses[last][3]--
	if err := validateNativeFieldFile(changed, plan, 4); err == nil {
		t.Fatal("stale/wrong field assignment accepted")
	}
	t.Log("new Gooo producer4/4;exact2-run field replay;copy+field mutation+helper arguments/record returns+atomic swap;equal zero payloads keep different branch-selected input-field roots;wrong field read rejected;models0")
}

func fieldRoots13421(p *compositionFieldProjection, writes, uses [][]int, root int) []string {
	seen := map[int]bool{}
	roots := map[string]bool{}
	var visit func(int)
	visit = func(index int) {
		if index < 0 || index >= len(writes) || seen[index] {
			return
		}
		seen[index] = true
		write := writes[index]
		site := p.sites[write[1]]
		binding := p.inventory.bindings[site.Binding]
		if write[7] >= 0 {
			visit(write[7])
			return
		}
		if site.Kind == "parameter" && binding.ActivityID == binding.RootActivityID {
			roots[binding.Record.Fields[site.Field].ID] = true
			return
		}
		for _, use := range uses[write[3]:write[4]] {
			visit(use[3])
		}
	}
	visit(root)
	result := make([]string, 0, len(roots))
	for field := range roots {
		result = append(result, field)
	}
	slices.Sort(result)
	return result
}
