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

func localCases(rows []any, passed, total any) int {
	n := 0
	for _, row := range rows {
		match := reflect.DeepEqual(at(row, "actual"), at(row, "expected"))
		require(at(row, "passed") == match, "local case flag differs from exact values")
		if match {
			n++
		}
	}
	require(number(passed) == n && number(total) == len(rows), "local case totals differ")
	return n
}

func recount(root, fixture, name string) map[string]any {
	r := read(filepath.Join(root, name+".json"))
	prefix := ""
	if strings.HasPrefix(name, "mixed") {
		prefix = "mixed-"
	}
	feedback := read(filepath.Join(fixture, prefix+"construction-cases.json"))["cases"].([]any)
	evaluation := read(filepath.Join(fixture, prefix+"evaluation-cases.json"))["cases"].([]any)
	c := r["construction"]
	require(at(c, "schema") == "gooo/joint-construction/v4" && at(c, "stage") == "COMPLETE", "source-fill schema or stage differs")
	require(reflect.DeepEqual(at(c, "construction_cases", "cases"), feedback), "original caller cases changed")
	attempts := at(c, "attempts").([]any)
	for _, a := range attempts {
		require(at(a, "rejection") == nil, "unexpected rejection in valid source-fill fixture")
		local, total := 0, 0
		for _, f := range at(a, "fill_candidates").([]any) {
			rows := at(f, "value_case_results").([]any)
			local += localCases(rows, at(f, "test_cases_passed"), at(f, "test_cases_total"))
			total += len(rows)
			localCases(at(f, "value_holdout_results").([]any), at(f, "holdout_cases_passed"), at(f, "holdout_cases_total"))
		}
		if records, ok := at(a, "candidates").([]any); ok {
			for _, candidate := range records {
				rows := at(candidate, "cases").([]any)
				local += localCases(rows, at(candidate, "attempt", "passed"), at(candidate, "attempt", "total"))
				total += len(rows)
			}
		}
		require(number(at(a, "local_passed")) == local && number(at(a, "local_total")) == total, "holdouts mixed into training totals")
		checkRuntime(at(a, "runtime"), feedback)
	}
	passed := checkRuntime(at(r, "evaluation", "runtime"), evaluation)
	require(number(at(r, "evaluation", "new_model_calls")) == 0, "evaluation predicted again")
	historical, fresh := 0, 0
	for _, step := range at(c, "initial", "preparations").([]any) {
		report := at(step, "generation", "report").(map[string]any)
		if fill, ok := report["body_fill"].(map[string]any); ok {
			require(number(fill["local_model_predictions"]) == 0, "unexpected fill-model prediction")
		}
		if record, ok := report["record_assembly"].(map[string]any); ok {
			historical += number(record["model_calls"])
		}
	}
	if r["generated_now"] == true {
		fresh = historical
	}
	model, ok := at(c, "initial", "model").(map[string]any)
	if strings.Contains(name, "model") || name == "mixed-replay" {
		require(ok && model["loaded"] == true && historical == 1, "own model was not used")
		require(model["metadata_sha256"] == "3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202" &&
			model["weights_sha256"] == "76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f", "own model changed")
	} else {
		require(historical == 0, "deterministic run predicted")
	}
	last := attempts[number(at(c, "selected_attempt"))]
	fill := at(last, "fill_candidates").([]any)[0]
	return map[string]any{"name": name, "attempts": len(attempts), "native_program_attempts": len(attempts),
		"decision": at(c, "decision"), "evaluation_passed": passed, "evaluation_total": len(evaluation),
		"local_passed": at(last, "local_passed"), "local_total": at(last, "local_total"),
		"fill_holdout_passed": at(fill, "holdout_cases_passed"), "fill_holdout_total": at(fill, "holdout_cases_total"),
		"fresh_model_predictions": fresh, "historical_model_predictions": historical,
		"replayed": at(r, "evaluation", "construction_replayed")}
}

var names = []string{"budget-2", "budget-3", "budget-replay", "mixed-fixed", "mixed-model", "mixed-replay"}

func comparePriorRuns(root, priorRoot string) {
	for _, name := range names {
		a, b := read(filepath.Join(root, name+".json")), read(filepath.Join(priorRoot, name+".json"))
		for _, key := range []string{"selected_source", "selected_attempt", "candidate_kinds", "candidate_space", "decision", "stop_reason"} {
			require(reflect.DeepEqual(at(a, "construction", key), at(b, "construction", key)), "prior construction selection differs: "+key)
		}
		require(reflect.DeepEqual(at(a, "evaluation", "runtime", "traces"), at(b, "evaluation", "runtime", "traces")), "prior exact final values differ")
		x, y := at(a, "construction", "attempts").([]any), at(b, "construction", "attempts").([]any)
		require(len(x) == len(y), "prior attempt count differs")
		for i := range x {
			for _, key := range []string{"masks", "fill_candidates", "candidates", "local_passed", "local_total"} {
				require(reflect.DeepEqual(at(x[i], key), at(y[i], key)), "prior candidate history differs: "+key)
			}
			require(reflect.DeepEqual(at(x[i], "runtime", "traces"), at(y[i], "runtime", "traces")), "prior exact caller values differ")
		}
	}
}

func main() {
	if len(os.Args) != 4 {
		panic("provide candidate observations, original fixtures and prior observations")
	}
	root, fixture, prior := os.Args[1], os.Args[2], os.Args[3]
	rows := []any{}
	for _, name := range names {
		rows = append(rows, recount(root, fixture, name))
	}
	comparePriorRuns(root, prior)
	build := read(filepath.Join(root, "build.json"))
	platform := read(filepath.Join(root, "platform", "darwin-arm64.receipt.json"))
	require(build["version"] == "0.6.14-dev" && build["vcs_modified"] == "false" && build["go_version"] == "go1.27.1", "candidate build identity differs")
	require(platform["decision"] == "PASS" && platform["resolution"] == "EXACT" && platform["head_sha"] == build["vcs_revision"], "platform identity differs")
	require(at(platform, "archive", "replay_equal") == true && number(at(platform, "archive", "builds")) == 2, "two native archives differ")
	for _, pair := range [][2]string{{"partial", "budget-2"}, {"construct", "budget-3"}, {"replay", "budget-replay"}} {
		a := read(filepath.Join(root, "platform", "darwin-arm64-caller-fill-"+pair[0]+".json"))
		b := read(filepath.Join(root, pair[1]+".json"))
		require(at(a, "construction", "selected_source") == at(b, "construction", "selected_source") &&
			reflect.DeepEqual(at(a, "evaluation", "runtime", "traces"), at(b, "evaluation", "runtime", "traces")), "platform source-fill values differ")
	}
	fixed, model := read(filepath.Join(root, "mixed-fixed.json")), read(filepath.Join(root, "mixed-model.json"))
	require(at(fixed, "construction", "selected_source") == at(model, "construction", "selected_source"), "fixed/model selected sources differ")
	result := map[string]any{"schema": "gooo/release-0614-source-fill-regression/v1", "compiler": build, "native_platform": platform, "rows": rows,
		"prior_values_and_assignments_equal": true, "fixed_model_selected_source_equal": true, "host_cpu_utilization": "NOT_SAMPLED",
		"scope": "Local candidate regression of two related frozen programs: four fresh constructions and two saved replays. Same weights and expectations as the feature observation; no new training, generalization or controlled performance claim. The runner label names a configured target; this witness ran on the local macOS host."}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	must(e.Encode(result))
	fmt.Fprintln(os.Stderr, "PASS: original cases, assignments, local/holdout separation, caller and final values match the prior observation; native archive reproduction and saved replay match.")
}
