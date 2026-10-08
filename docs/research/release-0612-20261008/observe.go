package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
func read(path string) map[string]any {
	f, err := os.Open(path)
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
func integer(value any) int { n, err := value.(json.Number).Int64(); must(err); return int(n) }
func checkRuntime(value any) {
	r := value.(map[string]any)
	passed, total := 0, 0
	for _, trace := range r["traces"].([]any) {
		for _, item := range at(trace, "deliveries").([]any) {
			d := item.(map[string]any)
			if expected, ok := d["expected"]; ok {
				total++
				match := reflect.DeepEqual(expected, d["actual"])
				require(d["passed"] == match, "reported caller result differs from actual exact values")
				if match {
					passed++
				}
			}
		}
	}
	require(passed == integer(r["finite_passed"]) && total == integer(r["finite_total"]), "caller counts differ")
	require(integer(r["model_calls"]) == 0 && r["projection_replayed"] == true && r["runtime_replayed"] == true, "native execution flags differ")
}

func checkConstruction(value any) {
	for _, attempt := range at(value, "attempts").([]any) {
		checkRuntime(at(attempt, "runtime"))
		localPassed, localTotal := 0, 0
		for _, candidate := range at(attempt, "candidates").([]any) {
			matched := 0
			cases := at(candidate, "cases").([]any)
			for _, value := range cases {
				match := reflect.DeepEqual(at(value, "actual"), at(value, "expected"))
				require(at(value, "passed") == match, "local values disagree with score")
				if match {
					matched++
				}
			}
			require(matched == integer(at(candidate, "attempt", "passed")) &&
				len(cases) == integer(at(candidate, "attempt", "total")), "local case counts differ")
			localPassed, localTotal = localPassed+matched, localTotal+len(cases)
		}
		require(localPassed == integer(at(attempt, "local_passed")) &&
			localTotal == integer(at(attempt, "local_total")), "aggregate local counts differ")
	}
}

func main() {
	if len(os.Args) != 2 {
		panic("provide observation root")
	}
	root := os.Args[1]
	var rows []any
	for _, budget := range []int{1, 8} {
		for _, mode := range []string{"fixed", "model"} {
			name := fmt.Sprintf("%s-%d", mode, budget)
			r := read(filepath.Join(root, name+".json"))
			c := r["construction"].(map[string]any)
			attempts := c["attempts"].([]any)
			checkConstruction(c)
			checkRuntime(at(r, "evaluation", "runtime"))
			prepared := at(c, "initial", "preparations").([]any)
			assembly := at(prepared[0], "generation", "report", "record_assembly")
			modelCalls := integer(at(assembly, "model_calls"))
			require((mode == "model" && modelCalls == 1) || (mode == "fixed" && modelCalls == 0), "unexpected inference count")
			separation := at(r, "evaluation", "input_separation")
			require(integer(at(separation, "construction_inputs")) == 1 && integer(at(separation, "other_inputs")) == 6, "caller exposure lost")
			rows = append(rows, map[string]any{"mode": mode, "program_budget": budget, "program_attempts": len(attempts),
				"decision": c["decision"], "selected_attempt": c["selected_attempt"], "construction_ns": c["elapsed_ns"],
				"evaluation_passed": at(r, "evaluation", "runtime", "finite_passed"), "evaluation_total": at(r, "evaluation", "runtime", "finite_total"),
				"model_calls": modelCalls, "prediction_ns": at(assembly, "predict_ns"), "model": at(c, "initial", "model"),
				"initial_local_attempts": len(at(assembly, "attempts").([]any)), "selected_source_sha256": at(c, "selected", "original_source_sha256")})
		}
	}
	fixed, model, replay := read(filepath.Join(root, "fixed-8.json")), read(filepath.Join(root, "model-8.json")), read(filepath.Join(root, "model-replay.json"))
	require(at(fixed, "construction", "selected_source") == at(model, "construction", "selected_source"), "final source differs")
	require(reflect.DeepEqual(at(fixed, "evaluation", "runtime", "traces"), at(model, "evaluation", "runtime", "traces")), "fixed/model actual results differ")
	require(reflect.DeepEqual(at(model, "evaluation", "runtime", "traces"), at(replay, "evaluation", "runtime", "traces")), "saved actual results differ")
	require(at(replay, "evaluation", "construction_replayed") == true && integer(at(replay, "evaluation", "new_model_calls")) == 0, "saved model history did not replay")
	checkConstruction(at(replay, "construction"))
	checkRuntime(at(replay, "evaluation", "runtime"))
	simple, simpleReplay := read(filepath.Join(root, "simple.json")), read(filepath.Join(root, "simple-replay.json"))
	require(at(simple, "construction", "decision") == "COMPLETE_FINITE" && len(at(simple, "construction", "attempts").([]any)) == 2, "simple caller gap remains")
	checkConstruction(at(simple, "construction"))
	checkConstruction(at(simpleReplay, "construction"))
	require(at(simpleReplay, "evaluation", "construction_replayed") == true &&
		integer(at(simpleReplay, "evaluation", "new_model_calls")) == 0, "simple history did not replay")
	checkRuntime(at(simple, "evaluation", "runtime"))
	checkRuntime(at(simpleReplay, "evaluation", "runtime"))
	require(reflect.DeepEqual(at(simple, "evaluation", "runtime", "traces"), at(simpleReplay, "evaluation", "runtime", "traces")), "simple replay differs")
	platform := read(filepath.Join(root, "platform", "darwin-arm64.receipt.json"))
	require(platform["decision"] == "PASS" && platform["resolution"] == "EXACT" &&
		platform["head_sha"] == at(read(filepath.Join(root, "build.json")), "compiler_source_sha"), "native platform source differs")
	result := map[string]any{"schema": "gooo/release-0612-caller-regression/v1", "compiler": read(filepath.Join(root, "build.json")),
		"native_platform": platform, "model_demonstration_programs": 1, "paired_mode_budget_runs": 4, "evaluation_root_inputs": 7, "consumed_caller_inputs": 1,
		"other_root_inputs": 6, "root_zero_also_local_example": true, "rows": rows, "selected_source_equal": true,
		"saved_replay_new_model_calls": 0, "simple_regression_programs": 1, "simple_program_attempts": 2,
		"host_cpu_utilization": "NOT_SAMPLED", "scope": "Release regression of previously observed programs and unchanged weights; exact finite values and saved native replay; no training, generalization or general performance conclusion."}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	must(e.Encode(result))
	fmt.Fprintln(os.Stderr, "PASS: all recorded local and caller values counted exactly; fixed/model final source and values equal; saved history and evaluation use zero new inference.")
}
