package bodycodegen

import "testing"

func TestLocalIterationSources9883(t *testing.T) {
	const source = `package locals9883
namespace locals9883
entity Integer id "urn:gooo:type:integer"
entity Values id "urn:locals9883:values" fields {field items id "urn:locals9883:items" type integer required many}
activity Sum(Values) -> Integer id "urn:locals9883:sum" computes "let total=53;for part in input.items {total=total+part};let cursor=0;for cursor=0; cursor < 3; cursor=cursor+1 {total=total+59};return total"
`
	generated, err := Generate("반복지역9883.gooo", []byte(source), "Sum")
	if err != nil {
		t.Fatal(err)
	}
	bound, err := BindBodyLocalSources([]byte(source), generated.Report.BodyLocals)
	if err != nil {
		t.Fatal(err)
	}
	saveLocalOrigin9811(t, "local9883.json", map[string]any{"source": source, "generation": generated, "bound": bound, "native_executions": 0})
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
