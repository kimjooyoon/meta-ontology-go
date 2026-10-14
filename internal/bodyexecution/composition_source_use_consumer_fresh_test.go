package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoooSourceUseConsumer10841(t *testing.T) {
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, "examples/native-origin-reader", name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	reader := read("locals/locals.gooo")
	for _, sample := range []struct {
		name, classification, mapping string
		offset                        int
	}{{"local10773-runtime.json", "assignment", "MAPPED", 0}, {"local10433-runtime.json", "declaration-only", "UNMAPPED", 6}} {
		var producer CompositionRuntime
		if err := json.Unmarshal(read(sample.name), &producer); err != nil {
			t.Fatal(err)
		}
		rows, page, err := SelectRecordedNativeOrigins(producer, NativeOriginQuery{Kind: "local-uses", Offset: sample.offset, Limit: 2})
		if err != nil || len(rows) != 2 || page.ProducerExecutions != 0 || page.ProducerModelCalls != 0 {
			t.Fatal("saved input selection", err, page)
		}
		suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
		for _, row := range rows {
			var fact CompositionLocalUseFact
			if err := json.Unmarshal(row, &fact); err != nil || fact.SourceMapping != sample.mapping {
				t.Fatal("source mapping", err, fact)
			}
			if sample.mapping == "MAPPED" && (fact.SourceSHA256 != producer.SelectedSourceSHA256 || fact.SourceStart == fact.DeclarationStart || fact.SourceEnd <= fact.SourceStart) {
				t.Fatal("use and declaration coordinates", fact)
			}
			if sample.mapping == "UNMAPPED" && (fact.SourceSHA256 != "" || fact.SourceStart != 0 || fact.SourceEnd != 0) {
				t.Fatal("legacy record gained an invented source location", fact)
			}
			expected, _ := json.Marshal(map[string]any{"case_index": fact.CaseIndex, "use_index": fact.UseIndex, "assignment_index": fact.AssignmentIndex, "variable_name": fact.VariableName, "classification": sample.classification})
			suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Explain": row}, Expected: map[string]json.RawMessage{"Explain": expected}})
		}
		composition, err := GenerateCompositionWithOptions(context.Background(), "reader10841.gooo", reader, suite, CompositionOptions{EntryActivity: "Explain"})
		if err != nil {
			t.Fatal(err)
		}
		observed, err := ExecuteComposition(context.Background(), "reader10841.gooo", reader, composition, suite, "")
		if err != nil || observed.FinitePassed != 2 || !observed.RuntimeReplayed {
			t.Fatal("Gooo classification", sample.name, err, observed.FinitePassed)
		}
		saveReadObservation(t, "consumer10841-"+sample.name, map[string]any{"record": sample.name, "source": string(reader), "cases": suite, "selected_rows": rows, "origin_input": page, "composition": composition, "runtime": observed})
	}
	for name, mutate := range map[string]func(*CompositionLocalUseSite){
		"carrier":            func(s *CompositionLocalUseSite) { s.Source.SourceSHA256 = "sha256:other" },
		"expression":         func(s *CompositionLocalUseSite) { s.Source.Expression = "unrelated" },
		"ordinal":            func(s *CompositionLocalUseSite) { s.VariableUseOrdinal = -1 },
		"declaration_as_use": func(s *CompositionLocalUseSite) { s.Source.Start = s.Declaration.Start },
	} {
		var producer CompositionRuntime
		if err := json.Unmarshal(read("local10773-runtime.json"), &producer); err != nil {
			t.Fatal(err)
		}
		mutate(&producer.LocalUseSites[0])
		if _, _, err := SelectRecordedNativeOrigins(producer, NativeOriginQuery{Kind: "local-uses", Limit: 1}); err == nil {
			t.Fatal("contradictory source-use claim accepted", name)
		}
	}
	t.Log("Gooo reader2/2 new-source uses,2/2 legacy declaration-only uses;four source-claim contradictions rejected;original producer executions0;models0")
}
