package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestScalarSequenceGraphNeedsNoFabricatedRecordDictionary(t *testing.T) {
	source, err := os.ReadFile("../constructed-record-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	start, end := strings.Index(text, "`"), strings.LastIndex(text, "`")
	text = text[:start+1] + "let items = []int64{input.left, input.right}\n let total = 0\n for item in items { total = total + item }\n return total" + text[end:]
	r, err := bodycodegen.InspectTypedValueOrigins("scalar-list.gooo", []byte(text), "Read")
	if err != nil || r.Status != "RESOLVED" || len(r.Flow.RecordLayouts) != 0 || len(r.Flow.RecordValues) != 0 {
		t.Fatal("scalar source invented records", err, r.Reason)
	}
	raw, err := project(r, r.Outputs[0].FieldID, false, false)
	if err != nil {
		t.Fatal(err)
	}
	var graph originGraph
	if err := json.Unmarshal(raw, &graph); err != nil || graph.Layouts == nil || graph.ValuePages == nil {
		t.Fatal("empty tables became unusable nulls", err, string(raw))
	}
	for _, fault := range []string{"none", "orphan-layout", "ghost-value", "ghost-version"} {
		t.Run(fault, func(t *testing.T) {
			var changed originGraph
			if err := json.Unmarshal(raw, &changed); err != nil {
				t.Fatal(err)
			}
			switch fault {
			case "orphan-layout":
				changed.Layouts = []layout{{TypeID: "orphan", Members: []field{{FieldID: "unused", TypeID: "urn:gooo:type:integer"}}}}
			case "ghost-value":
				changed.Pages[0].Nodes[0].Value = 1
			case "ghost-version":
				changed.Pages[0].Nodes[1].Version = 1
			}
			input, _ := json.Marshal(changed)
			result, err := query(context.Background(), readerSource, input)
			if err != nil {
				t.Fatal(err)
			}
			var summary struct{ Status, Paths string }
			if err := json.Unmarshal(result.Result, &summary); err != nil {
				t.Fatal(err)
			}
			if fault == "none" {
				if summary.Status != "RESOLVED" || summary.Paths != "input:urn:construct:left;input:urn:construct:right;" || result.Steps < 1 || result.Steps > result.Budget {
					t.Fatal("source-owned empty-table work plan changed", summary, result.Steps, result.Budget)
				}
			} else if summary.Status != "UNUSABLE_INPUT" || summary.Paths != "" {
				t.Fatal("malformed empty dictionary published accepted paths", fault, summary)
			}
		})
	}
}
