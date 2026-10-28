package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func loopGraph(t *testing.T) originGraph {
	t.Helper()
	source, err := os.ReadFile("../loop-carried-values/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	r, err := bodycodegen.InspectTypedValueOrigins("carry.gooo", source, "Shift")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := project(r, "", false, false)
	if err != nil {
		t.Fatal(err)
	}
	var graph originGraph
	if err := json.Unmarshal(raw, &graph); err != nil {
		t.Fatal(err)
	}
	return graph
}

func TestLoopQueryRejectsIncompleteRecurrences(t *testing.T) {
	for _, kind := range []string{"page", "loop", "seed", "yield", "duplicate", "order", "missing", "actual-index", "old-wire"} {
		t.Run(kind, func(t *testing.T) {
			graph := loopGraph(t)
			carries := graph.CarryPages[0].Carries
			switch kind {
			case "old-wire":
				graph.Schema = "gooo/record-origin-graph/v1"
			case "page":
				graph.CarryPages[0].First = 2
			case "loop":
				carries[0].Loop = 2
			case "seed":
				carries[0].Seed++
			case "yield":
				carries[0].Yield = uint16(graph.Count + 1)
			case "duplicate":
				carries[1] = carries[0]
			case "order":
				carries[0], carries[1] = carries[1], carries[0]
			case "missing":
				graph.CarryCount--
				graph.CarryPages[0].Carries = carries[:1]
			case "actual-index":
				graph.Loops[0].IndexKind = "native_position_0"
			}
			raw, err := json.Marshal(graph)
			if err != nil {
				t.Fatal(err)
			}
			result, err := query(context.Background(), readerSource, raw)
			if err != nil {
				t.Fatal(err)
			}
			var summary struct {
				Status, Paths string
				Contributors  int
			}
			if err := json.Unmarshal(result.Result, &summary); err != nil {
				t.Fatal(err)
			}
			if summary.Status != "UNUSABLE_INPUT" || summary.Paths != "" || summary.Contributors != 0 {
				t.Fatal("incomplete recurrence produced accepted source paths", string(result.Result))
			}
		})
	}
}
