package bodycodegen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func saveLocalOrigin9811(t *testing.T, name string, value any) {
	t.Helper()
	if dir := os.Getenv("GOOO_LOCAL_ORIGIN_OBSERVATION"); dir != "" {
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLocalWriteSources9811(t *testing.T) {
	const source = `package locals9811
namespace locals9811
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:locals9811:input" fields {
 field values id "urn:locals9811:values" type integer required many
 field reverse id "urn:locals9811:reverse" type boolean required one
}
entity Batch id "urn:locals9811:batch" fields { field values id "urn:locals9811:out-values" type integer required many }
activity Select(Input) -> Integer id "urn:locals9811:select" computes "if input.reverse {let chosen=input.values[1]+17;return chosen}else{let chosen=input.values[0]-19;return chosen}"
activity Build(Input) -> Batch id "urn:locals9811:build" computes "let 이전=Select(input);if input.reverse {이전=input.values[0]};let adjusted=이전+23;let unused=31;let values=[]int64{adjusted,이전};return Batch{values:values}"
`
	generated, err := Generate("지역변수9811.gooo", []byte(source), "Build")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := BindBodyLocalSources([]byte(source), generated.Report.BodyLocals)
	if err != nil {
		t.Fatal(err)
	}
	saveLocalOrigin9811(t, "local9811.json", map[string]any{"source": source, "generation": generated, "bound": bound, "native_executions": 0})
	origins := generated.Report.BodyLocals
	if len(origins) != 7 || len(bound) != 7 {
		t.Fatal("write population", len(origins), len(bound))
	}
	variables := map[string][]BodyLocalOrigin{}
	for i, o := range origins {
		b := bound[i]
		if b.Value.Status != "MAPPED" || b.Declaration.Status != "MAPPED" || b.Value.SourceSHA256 != digest([]byte(source)) || b.Declaration.Expression != o.Name {
			t.Fatal("source binding", o, b)
		}
		if string([]byte(source)[b.Declaration.Start:b.Declaration.End]) != o.Name {
			t.Fatal("original declaration byte range", o, b)
		}
		variables[o.ActivityID+":"+o.Name] = append(variables[o.ActivityID+":"+o.Name], o)
	}
	writes := variables["urn:locals9811:build:이전"]
	if len(writes) != 2 || writes[0].VariableOrdinal != writes[1].VariableOrdinal || writes[0].Declaration != writes[1].Declaration || writes[0].Expression == writes[1].Expression {
		t.Fatal("reassignment lost variable identity", writes)
	}
	siblings := variables["urn:locals9811:select:chosen"]
	if len(siblings) != 2 || siblings[0].VariableOrdinal == siblings[1].VariableOrdinal || siblings[0].Declaration == siblings[1].Declaration {
		t.Fatal("same-name sibling bindings merged", siblings)
	}
	if !strings.Contains(generated.Source, "_ = unused") {
		t.Fatal("unused initializer marker missing")
	}
	for _, o := range origins {
		if o.Name == "_" {
			t.Fatal("synthetic discarded read invented a write")
		}
	}
	changed := append([]BodyLocalOrigin(nil), origins...)
	changed[0].VariableOrdinal++
	if EqualRetainedLocalOrigins(changed, origins) || !EqualRetainedLocalOrigins(nil, origins) {
		t.Fatal("retained origin comparison")
	}
	t.Log("7 source writes bound; Korean reassignment shares declaration; sibling chosen variables remain distinct; unused initializer retained; no native execution")
}

func TestLocalIterationSources9833(t *testing.T) {
	const source = `package locals9833
namespace locals9833
entity Integer id "urn:gooo:type:integer"
entity Values id "urn:locals9833:values" fields {field items id "urn:locals9833:items" type integer required many}
activity Sum(Values) -> Integer id "urn:locals9833:sum" computes "let total=37;for item in input.items {total=total+item};let index=0;for index=0; index < 2; index=index+1 {total=total+41};return total"
`
	generated, err := Generate("반복지역9833.gooo", []byte(source), "Sum")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := BindBodyLocalSources([]byte(source), generated.Report.BodyLocals)
	if err != nil {
		t.Fatal(err)
	}
	saveLocalOrigin9811(t, "local9833.json", map[string]any{"source": source, "generation": generated, "bound": bound, "native_executions": 0})
	origins := generated.Report.BodyLocals
	if len(origins) != 7 {
		t.Fatal("loop write population", len(origins), origins)
	}
	counts := map[string]int{}
	for i, o := range origins {
		counts[o.Kind]++
		if bound[i].Value.Status != "MAPPED" || bound[i].Declaration.Status != "MAPPED" {
			t.Fatal("loop source mapping", o, bound[i])
		}
	}
	if counts["iteration"] != 1 || counts["update"] != 0 || counts["assignment"] != 4 {
		t.Fatal(counts)
	}
	byName := map[string]int{}
	for _, o := range origins {
		if prior, ok := byName[o.Name]; ok && prior != o.VariableOrdinal {
			t.Fatal("loop write identity differs", o)
		}
		byName[o.Name] = o.VariableOrdinal
	}
	t.Log("7 write sites retain collection iteration, counted-loop initialization/update and total reassignments; native executions0")
}

func TestSyntheticJoinLocalOrigins9851(t *testing.T) {
	for i, route := range []string{preserveRoute, guardReturnRoute, mergeResultRoute} {
		body := fmt.Sprintf("if input < %d {return input + %d}else{return input - %d}", 71+i, 43+i, 47+i)
		generated, err := generateRoute("locals9851", "Choose", "urn:locals9851:choose", "int64", "int64", body, route)
		if err != nil {
			t.Fatal(route, err)
		}
		if len(generated.report.BodyLocals) != 0 {
			t.Fatal("synthetic route join claimed an authored local", route, generated.report.BodyLocals)
		}
		source := []byte("package locals9851\nnamespace locals9851\nentity Integer id \"urn:gooo:type:integer\"\nactivity Choose(Integer) -> Integer id \"urn:locals9851:choose\" computes " + strconv.Quote(body) + "\n")
		saveLocalOrigin9811(t, "local9851-"+strconv.Itoa(i)+".json", map[string]any{"source": string(source), "report": generated.report, "projection": string(generated.source), "native_executions": 0})
	}
	t.Log("three distinct route projections do not assign authored origin to synthetic result joins")
}
