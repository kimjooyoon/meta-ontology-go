package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func require(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func read(name string) map[string]any {
	f, err := os.Open(name)
	must(err)
	defer f.Close()
	d := json.NewDecoder(f)
	d.UseNumber()
	var value map[string]any
	must(d.Decode(&value))
	return value
}
func at(value any, keys ...string) any {
	for _, key := range keys {
		value = value.(map[string]any)[key]
	}
	return value
}
func number(value any) int {
	n, err := value.(json.Number).Int64()
	must(err)
	return int(n)
}

// Expectations and inputs come from the separately retained fixture files.
// UseNumber keeps adjacent integers above 2^53 distinct.
func checkRuntime(runtime any, cases []any) int {
	r := runtime.(map[string]any)
	traces := r["traces"].([]any)
	require(r["stage"] == "COMPLETE" && r["projection_replayed"] == true && r["runtime_replayed"] == true && number(r["model_calls"]) == 0, "incomplete native observation")
	require(len(traces) == len(cases), "native row count differs")
	passed, seen := 0, make([]bool, len(cases))
	for _, trace := range traces {
		i := number(at(trace, "case_index"))
		require(i >= 0 && i < len(cases) && !seen[i], "native row identity differs")
		seen[i] = true
		deliveries := at(trace, "deliveries").([]any)
		require(len(deliveries) == 1, "expected one caller delivery")
		d := deliveries[0]
		expected := at(cases[i], "expected", "Main")
		require(reflect.DeepEqual(at(d, "input"), at(cases[i], "inputs", "Main")) && reflect.DeepEqual(at(d, "expected"), expected), "native input or original expectation changed")
		match := reflect.DeepEqual(at(d, "actual"), expected)
		require(at(d, "passed") == match, "native passed flag differs from exact actual value")
		if match {
			passed++
		}
	}
	require(number(r["finite_passed"]) == passed && number(r["finite_total"]) == len(cases), "native totals differ")
	return passed
}

func recount(root, fixture, name string) map[string]any {
	r := read(filepath.Join(root, name+".json"))
	prefix := ""
	if strings.HasPrefix(name, "mixed") {
		prefix = "mixed-"
	}
	feedback := read(filepath.Join(fixture, prefix+"construction-cases.json"))["cases"].([]any)
	evaluation := read(filepath.Join(fixture, prefix+"evaluation-cases.json"))["cases"].([]any)
	attempts := at(r, "construction", "attempts").([]any)
	rejected := 0
	for _, value := range attempts {
		a := value.(map[string]any)
		if rejection, ok := a["rejection"]; ok {
			rejected++
			require(at(rejection, "stage") == "LOCAL_SOURCE_SEARCH" && at(a, "runtime", "stage") == "" && number(at(a, "runtime", "finite_total")) == 0, "rejected candidate claims execution")
			searches := a["search_candidates"].([]any)
			last := searches[len(searches)-1]
			require(at(last, "attempt", "error") == at(rejection, "reason") && at(last, "attempt", "scoring_completed") == false && at(last, "attempt", "accuracy_percent") == nil, "rejection falsely scored")
			continue
		}
		checkRuntime(a["runtime"], feedback)
	}
	passed := checkRuntime(at(r, "evaluation", "runtime"), evaluation)
	require(number(at(r, "evaluation", "new_model_calls")) == 0, "evaluation performed inference")
	return map[string]any{"name": name, "attempts": len(attempts), "rejected": rejected,
		"native_program_attempts": len(attempts) - rejected, "decision": at(r, "construction", "decision"),
		"evaluation_passed": passed, "evaluation_total": len(evaluation), "replayed": at(r, "evaluation", "construction_replayed")}
}

func main() {
	if len(os.Args) != 3 {
		panic("provide observation root and independent fixture directory")
	}
	root, fixture := os.Args[1], os.Args[2]
	var rows []any
	for _, name := range []string{"scalar", "partial", "scalar-replay", "mixed-fixed", "mixed-model", "mixed-model-replay"} {
		rows = append(rows, recount(root, fixture, name))
	}
	fixed, model, replay := read(filepath.Join(root, "mixed-fixed.json")), read(filepath.Join(root, "mixed-model.json")), read(filepath.Join(root, "mixed-model-replay.json"))
	require(at(fixed, "construction", "selected_source") == at(model, "construction", "selected_source"), "fixed/model source differs")
	require(reflect.DeepEqual(at(model, "evaluation", "runtime", "traces"), at(replay, "evaluation", "runtime", "traces")), "saved mixed values differ")
	simple, saved := read(filepath.Join(root, "scalar.json")), read(filepath.Join(root, "scalar-replay.json"))
	require(reflect.DeepEqual(at(simple, "evaluation", "runtime", "traces"), at(saved, "evaluation", "runtime", "traces")), "saved scalar values differ")
	result := map[string]any{"schema": "gooo/caller-search-rejection-observation/v1", "compiler": read(filepath.Join(root, "build.json")),
		"rows": rows, "fixed_model_selected_source_equal": true, "saved_values_equal": true,
		"host_cpu_utilization": "NOT_SAMPLED", "scope": "Two fixed programs, four construction runs and two saved replays; unchanged weights; finite original expectations; no training or generalization conclusion."}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	must(e.Encode(result))
	fmt.Fprintln(os.Stderr, "PASS: exact original caller/evaluation values recounted; rejections remain unscored; selected source and saved values agree.")
}
