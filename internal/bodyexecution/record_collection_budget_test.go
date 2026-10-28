package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestNestedRecordValueBudgetChargesRepeatedScalarLeaves(t *testing.T) {
	records := []bodycodegen.RecordType{
		{Name: "Root", ID: "urn:budget:root", Fields: []bodycodegen.RecordField{{Name: "rows", ID: "urn:budget:rows", TypeID: "urn:budget:row", Cardinality: "many"}}},
		{Name: "Row", ID: "urn:budget:row", Fields: []bodycodegen.RecordField{{Name: "values", ID: "urn:budget:values", TypeID: "urn:gooo:type:integer", Cardinality: "many"}}},
	}
	for _, extra := range []bool{false, true} {
		// Two outer nodes +31*(one record+one list+254 leaves) +
		// one record+one list+252 leaves =8192 JSON value nodes.
		rows := make([]map[string][]int, 32)
		for i := range rows {
			count := 254
			if i == 31 {
				count = 252
				if extra {
					count++
				}
			}
			rows[i] = map[string][]int{"values": make([]int, count)}
		}
		raw, _ := json.Marshal(map[string]any{"rows": rows})
		canonical, err := canonicalRecordWithLayouts(raw, records[0], records)
		if extra {
			if err == nil || !strings.Contains(err.Error(), "8192 slots") || canonical != nil {
				t.Fatal("aggregate nested value budget widened", err)
			}
		} else if err != nil {
			t.Fatal("exact nested value budget rejected", err)
		}
	}
}

func TestNestedRecordBudgetPreservesExistingFlatCollectionLimits(t *testing.T) {
	record := bodycodegen.RecordType{Name: "Flat", ID: "urn:flat"}
	values := map[string][]int{}
	for i := range 32 {
		name := fmt.Sprintf("f%d", i)
		record.Fields = append(record.Fields, bodycodegen.RecordField{Name: name, ID: "urn:flat:" + name, TypeID: "urn:gooo:type:integer", Cardinality: "many"})
		values[name] = make([]int, 256)
	}
	raw, _ := json.Marshal(values)
	if _, err := canonicalRecord(raw, record); err != nil {
		t.Fatal("new nested quota narrowed the previous flat profile", err)
	}
}

func TestRecordCollectionTransitiveScopeKeepsSixteenTypeLimit(t *testing.T) {
	for _, count := range []int{16, 17} {
		var source strings.Builder
		source.WriteString("package scope\nnamespace scope\nentity Integer id \"urn:gooo:type:integer\"\nentity Root id \"urn:scope:root\" fields {\n")
		for i := 1; i < count; i++ {
			fmt.Fprintf(&source, "field r%d id \"urn:scope:field:%d\" type R%d required one\n", i, i, i)
		}
		source.WriteString("}\n")
		for i := 1; i < count; i++ {
			fmt.Fprintf(&source, "entity R%d id \"urn:scope:record:%d\" fields {\nfield value id \"urn:scope:value:%d\" type integer required one\n}\n", i, i, i)
		}
		source.WriteString("activity Read(Root) -> Integer computes `return 0`\n")
		result, err := GenerateComputedProjection(context.Background(), "scope.gooo", []byte(source.String()), "Read")
		if count == 17 {
			if err == nil || result.Source != "" || !strings.Contains(err.Error(), "supports at most 16") {
				t.Fatal("transitive field types escaped the selected-record limit", err)
			}
		} else if err != nil {
			t.Fatal("exact transitive record limit rejected", err)
		}
	}
}
