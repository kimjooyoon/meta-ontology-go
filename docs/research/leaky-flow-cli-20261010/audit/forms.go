package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type form struct {
	Name, ID             string
	Deterministic, Model int
	First, Final         uint16
	Normalized           bool
}

var forms = []form{
	{"max-direct", "max-k11-r1-w1-direct", 1, 1, 0, 0, true},
	{"max-copy", "max-k11-r1-w1-copy_after_branch", 1, 1, 0, 0, true},
	{"min-direct", "min-k11-r0-w0-direct", 1, 2, 2, 0, true},
	{"min-copy", "min-k11-r0-w0-copy_after_branch", 1, 2, 2, 0, true},
	{"nested", "control-nested", 3, 2, 0, 2, false},
}

func checkForm(root string, f form, original sdkJudgment) (bodycodegen.TypedPathContextExport, map[string]any) {
	sourceSHA := digest(read(root, "source.gooo.gz"))
	pre := decode[bodycodegen.TypedPathContextExport](read(root, "preflight.json.gz"))
	must(pre.OriginalSourceSHA256 == sourceSHA && pre.ModelPredictions == 0 && pre.CandidateTests == 0 && len(pre.Inputs) == 2, "preflight")
	must(pre.ModelCompatibility.Model.ModelSchema == "gooo/flow-candidate-decision/v2", "preflight activation schema")
	d := checkSearch(decode[bodycodegen.Result](read(root, "deterministic.json.gz")), f.Deterministic, 0, sourceSHA)
	initial := checkSearch(decode[bodycodegen.Result](read(root, "initial.json.gz")), f.Model, 1, sourceSHA)
	fb := checkSearch(decode[bodycodegen.Result](read(root, "feedback.json.gz")), f.Model, f.Model, sourceSHA)
	for _, p := range []*bodycodegen.BodyPathReceipt{d, initial, fb} {
		checkSourceOutputs(root, p)
	}
	for _, p := range []*bodycodegen.BodyPathReceipt{initial, fb} {
		must(p.Search.Attempts[0].Mask == f.First && p.Search.Attempts[len(p.Search.Attempts)-1].Mask == f.Final, "recorded first and final candidates")
		checkInputs(pre, p, f.Normalized)
		checkOriginal(root, pre, p, original)
	}
	must(len(fb.ConditionFeedback) == f.Model-1, "committed feedback count")
	extraNS := int64(0)
	for _, feedback := range fb.ConditionFeedback {
		must(feedback.Calls == 1 && feedback.OutputFailure != nil && !feedback.HasFailure && feedback.Proposed == f.First && feedback.Applied && !feedback.AddedMask, "unchanged failed proposal after feedback")
		extraNS += feedback.PredictNS
	}
	checkSavedConstruction(root, f.Model)
	return pre, map[string]any{"deterministic_attempts": f.Deterministic, "initial_attempts": f.Model, "feedback_attempts": f.Model,
		"first_mask": f.First, "final_mask": f.Final, "first_passed": f.First == f.Final, "initial_calls": 1, "feedback_calls": f.Model,
		"deterministic_total_ms": d.Timing.TotalMS, "initial_total_ms": initial.Timing.TotalMS, "feedback_total_ms": fb.Timing.TotalMS,
		"initial_predict_ns": initial.ConditionProgress[0].Ranking.PredictNS, "feedback_initial_predict_ns": fb.ConditionProgress[0].Ranking.PredictNS, "extra_feedback_predict_ns": extraNS,
		"source_outputs": 8, "source_conditions": 3, "construction_native_cases": 8, "model_free_replay_cases": 8, "repeated_candidate_executions": 0}
}

func audit(root string) map[string]any {
	build := decode[map[string]any](read(root, "producer-version.json"))
	must(build["compiler_source_sha"] == producer && build["source_status"] == "CLEAN_VCS" && build["go_version"] == "go1.27.2", "clean producer")
	sdk := build["decision_runtime"].(map[string]any)
	must(sdk["version"] == "v0.2.35-experimental" && sdk["replaced"] == false, "public SDK")
	must(strings.TrimSpace(string(read(root, "model.sha256"))) == model, "public model hash")
	original := originalJudgments(root)
	var completed strings.Builder
	all := map[string]bodycodegen.TypedPathContextExport{}
	results := map[string]any{}
	for _, f := range forms {
		for _, name := range []string{"preflight", "deterministic", "initial", "feedback", "construction", "replay"} {
			fmt.Fprintf(&completed, "%s\t%s\n", f.Name, name)
		}
		pre, report := checkForm(filepath.Join(root, f.Name), f, original[f.ID])
		all[f.Name], results[f.Name] = pre, report
	}
	must(string(read(root, "completed-commands.txt")) == completed.String(), "thirty original commands")
	for _, family := range []string{"max", "min"} {
		a, b := all[family+"-direct"], all[family+"-copy"]
		must(a.OriginalSourceSHA256 != b.OriginalSourceSHA256, "distinct source spelling")
		for i, input := range a.Inputs {
			must(input.InputSHA == b.Inputs[i].InputSHA && *input.FlowFeatures == *b.Inputs[i].FlowFeatures, "same actual arrays")
		}
	}
	return map[string]any{"status": "PASS", "producer": producer, "forms": results, "process": processResources(root),
		"paired_choices": 4, "native_processes": 20, "original_model_calls": 18,
		"new_audit_model_calls": 0, "new_audit_native_runs": 0}
}

func main() {
	must(len(os.Args) == 2, "usage: audit RESULT_DIRECTORY")
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	must(e.Encode(audit(os.Args[1])) == nil, "summary")
}
