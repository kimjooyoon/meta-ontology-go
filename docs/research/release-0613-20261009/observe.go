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
	if len(os.Args) != 4 {
		panic("provide candidate observations, independent fixtures and prior observations")
	}
	root, fixture := os.Args[1], os.Args[2]
	var rows []any
	for _, name := range []string{"scalar", "partial", "scalar-replay", "mixed-fixed", "mixed-model", "mixed-model-replay"} {
		rows = append(rows, recount(root, fixture, name))
	}
	comparePriorRuns(root, os.Args[3])
	platform := checkReleasePlatform(root)
	fixed, model, replay := read(filepath.Join(root, "mixed-fixed.json")), read(filepath.Join(root, "mixed-model.json")), read(filepath.Join(root, "mixed-model-replay.json"))
	require(at(fixed, "construction", "selected_source") == at(model, "construction", "selected_source"), "fixed/model source differs")
	require(reflect.DeepEqual(at(model, "evaluation", "runtime", "traces"), at(replay, "evaluation", "runtime", "traces")), "saved mixed values differ")
	simple, saved := read(filepath.Join(root, "scalar.json")), read(filepath.Join(root, "scalar-replay.json"))
	require(reflect.DeepEqual(at(simple, "evaluation", "runtime", "traces"), at(saved, "evaluation", "runtime", "traces")), "saved scalar values differ")
	result := map[string]any{"schema": "gooo/release-0613-rejection-regression/v1", "compiler": read(filepath.Join(root, "build.json")),
		"native_platform": platform, "prior_values_and_selections_equal": true,
		"rows": rows, "fixed_model_selected_source_equal": true, "saved_values_equal": true,
		"host_cpu_utilization": "NOT_SAMPLED", "scope": "Local release regression of the same two fixed programs, four construction runs and two saved replays; unchanged weights and expectations; no new training or generalization measurement. The platform runner field names the configured target profile; this run occurred on the local macOS host."}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	must(e.Encode(result))
	fmt.Fprintln(os.Stderr, "PASS: exact original caller/evaluation values recounted; rejections remain unscored; selected source and saved values agree.")
}

func comparePriorRuns(root, priorRoot string) {
	for _, name := range []string{"scalar", "partial", "scalar-replay", "mixed-fixed", "mixed-model", "mixed-model-replay"} {
		current, prior := read(filepath.Join(root, name+".json")), read(filepath.Join(priorRoot, name+".json"))
		require(at(current, "construction", "selected_source") == at(prior, "construction", "selected_source"), "release selected source differs from prior observation")
		require(reflect.DeepEqual(at(current, "evaluation", "runtime", "traces"), at(prior, "evaluation", "runtime", "traces")), "release exact evaluation values differ")
		a, b := at(current, "construction", "attempts").([]any), at(prior, "construction", "attempts").([]any)
		require(len(a) == len(b), "release attempt count differs")
		for i := range a {
			require(reflect.DeepEqual(at(a[i], "masks"), at(b[i], "masks")) && reflect.DeepEqual(at(a[i], "rejection"), at(b[i], "rejection")), "release candidate order or rejection changed")
		}
	}
}

func checkReleasePlatform(root string) map[string]any {
	b := read(filepath.Join(root, "build.json"))
	p := read(filepath.Join(root, "platform", "darwin-arm64.receipt.json"))
	require(b["version"] == "0.6.13-dev" && b["vcs_modified"] == "false" && b["go_version"] == "go1.27.1", "release build identity differs")
	require(p["decision"] == "PASS" && p["resolution"] == "EXACT" && p["head_sha"] == b["vcs_revision"], "platform source differs")
	require(at(p, "build", "vcs_modified") == false && at(p, "archive", "replay_equal") == true && number(at(p, "archive", "builds")) == 2, "native archive replay differs")
	for _, mode := range []string{"construct", "replay"} {
		platform := read(filepath.Join(root, "platform", "darwin-arm64-caller-search-"+mode+".json"))
		manual := read(filepath.Join(root, "scalar.json"))
		require(at(platform, "construction", "selected_source") == at(manual, "construction", "selected_source") &&
			reflect.DeepEqual(at(platform, "evaluation", "runtime", "traces"), at(manual, "evaluation", "runtime", "traces")), "platform and manual caller search differ")
	}
	return p
}
