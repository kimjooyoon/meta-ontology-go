package bodyexecution

import (
	"encoding/json"
	"runtime"
	"runtime/debug"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
)

// RuntimeProfileV1 records fresh execution, including pre-artifact failures.
const RuntimeProfileV1 = "gooo/typed-path-runtime-v1"

// RuntimeProfileV2 records an owned build separately from current-call work.
const RuntimeProfileV2 = "gooo/typed-path-runtime-v2"

// RuntimeProfileV3 additionally separates an owned Go version check from work
// performed by the current call. The Go executable bytes are checked each time.
const RuntimeProfileV3 = "gooo/typed-path-runtime-v3"

func producerSourceSHA() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "UNBOUND_LOCAL_SOURCE"
	}
	revision, clean := "", false
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			revision = s.Value
		}
		if s.Key == "vcs.modified" {
			clean = s.Value == "false"
		}
	}
	if !clean || len(revision) != 40 {
		return "UNBOUND_LOCAL_SOURCE"
	}
	return revision
}

func runtimeCompleteness(prior bodycodegen.Result, result Result) *completeness.CompletenessReceipt {
	r, err := completeness.Decode(result.ParentReceipt)
	if err != nil {
		r = bodycodegen.FailureCompletenessReceipt(prior.Report.Activity, nil, result.Observation.Failure)
	}
	o := result.Observation
	r.ProfileID = RuntimeProfileV1
	buildPassed := o.Build.Completed && o.ExecutableSHA256 != ""
	buildReason := "The Go tool and compiled executable bytes are bound separately."
	if o.Artifact != nil {
		r.ProfileID = RuntimeProfileV2
		buildPassed = o.Artifact.ExecutableVerified && o.Artifact.SourceBuild.Completed && o.ExecutableSHA256 != ""
		r.Scope["owned_artifact"] = o.Artifact
		buildReason = "The source-bound successful original build is recorded in owned_artifact; current Build records only work actually executed by this call."
	}
	if o.ToolchainReference != nil {
		r.ProfileID = RuntimeProfileV3
		r.Scope["owned_toolchain"] = o.ToolchainReference
	}
	r.DecisionBasis = "Original compiler dimensions are preserved; this producer adds exact source replay, native build, ordered compiled executions, finite caller expectations and reverse source links. Unobserved permissions, model-training generalization and full-domain behavior remain explicit."
	r.Scope["parent_receipt_sha256"] = o.ParentReceiptSHA256
	r.Scope["producer_source_sha"] = o.ProducerSourceSHA
	r.Scope["producer_toolchain"] = runtime.Version()
	r.Scope["parent_observation_authority"] = "caller-supplied prior receipt; source/body/finite observations replayed; prior model execution and declared compiler revision not independently attested"
	encoded, _ := json.Marshal(o)
	r.Scope["runtime_observation_sha256"] = digest(encoded)
	r.Scope["runtime_scope"] = map[string]any{"original_source_sha256": o.OriginalSourceSHA256,
		"selected_source_sha256": o.SelectedSourceSHA256, "activity_id": o.ActivityID, "generated_sha256": o.GeneratedSHA256,
		"runtime_suite_sha256": o.RuntimeSuiteSHA256, "executable_sha256": o.ExecutableSHA256,
		"local_model_predictions": 0, "external_provider_calls": 0, "temporary_workspace": true,
		"go_build_cache_writes": "possible in the configured Go cache; location inherited", "host_permission_profile": "inherited; unobserved",
		"suite_authority": "explicit caller expectations; disjointness is only relative to this selection suite, not model training"}
	if o.ToolchainReference != nil {
		r.Scope["runtime_scope"].(map[string]any)["go_tool_sha256"] = o.GoToolSHA256
		r.Scope["runtime_scope"].(map[string]any)["go_version"] = o.GoVersion
	}
	r.Scope["boundary"] = map[string]any{"non_executing": false, "non_authorizing": true, "repository_writes": "unobserved; temporary/cache locations inherited",
		"execution_profile": "source-replayed closed pure Go body, stdlib-only wrapper, bounded temporary build and two runs"}
	for i, v := range r.NotClaimed {
		if v == "generated runtime behavior" {
			r.NotClaimed[i] = "unobserved generated runtime behavior"
		}
	}
	if excluded, ok := r.Scope["excluded_scope"].([]any); ok {
		for i, v := range excluded {
			if v == "generated runtime behavior" {
				excluded[i] = "unobserved generated runtime behavior"
			}
		}
	}
	completed, passed := 0, 0
	for _, run := range o.Runs {
		if run.Completed {
			completed++
		}
	}
	for _, c := range o.Cases {
		if c.Passed {
			passed++
		}
	}
	linkCount := 0
	if o.ProjectionReplayed && o.ActivityID != "" && o.ExecutableSHA256 != "" {
		linkCount = len(o.Cases)
	}
	added := []completeness.CompletenessDimension{
		axis("runtime_completion", btoi(o.Stage == "COMPLETE"), 1, "completed bounded runtime observations", "Build, execution, decoding and replay errors are recorded at the actual stage.", o.Stage),
		axis("runtime_source_replay", btoi(o.ProjectionReplayed), 1, "exact source/document/selection/emission replays", "Only the closed selected projection may reach the build child.", o.OriginalSourceSHA256),
		axis("runtime_build", btoi(buildPassed), 1, "successful pinned Go builds", buildReason, o.GoToolSHA256),
		axis("execution_boundary", completed, 2, "completed compiled-program executions", "This producer actually executes the source-replayed generated package twice.", o.ExecutableSHA256),
		axis("runtime_finite_accuracy", passed, o.DeclaredCases, "ordered finite caller expectations matched", "These cases measure supplied expectations only; actual outputs and failures are retained.", o.RuntimeSuiteSHA256),
		axis("runtime_deterministic_replay", btoi(o.RuntimeReplayed), 1, "matching compiled output sequences", "Both independent process executions must return the same ordered int64 values.", o.ExecutableSHA256),
		axis("reverse_observation_coverage", linkCount, o.DeclaredCases, "runtime case observations linked to source identity", "Binds observed inputs/actuals through executable, generated bytes, selected/original source and stable activity ID.", digest(encoded)),
		axis("runtime_selection_disjointness", min(o.SelectionDisjointInputs, len(o.Cases)), o.DeclaredCases, "observed inputs absent from the declared selection suite", "This is not evidence that inputs were unseen during model training.", o.RuntimeSuiteSHA256),
		axis("runtime_provider_boundary", 1, 1, "runtime stages with no model or provider operation", "Replay, build and execution load no model and configure no provider.", "runtime local predictions:0; external provider calls:0"),
		runtimeResourceAxis(o),
	}
	for _, d := range added {
		if d.ID == "runtime_completion" && o.Failure != "" {
			d.Status = "FAIL_CLOSED"
			d.Reason = o.Stage + ": " + o.Failure
		}
		if len(o.Cases) > 0 && d.Numerator == 0 && (d.ID == "runtime_finite_accuracy" || d.ID == "runtime_selection_disjointness") {
			d.Status = "PROGRESS"
		}
		replaceAxis(r, d)
	}
	for _, id := range []string{"runtime_completion", "runtime_source_replay", "runtime_build", "execution_boundary", "runtime_finite_accuracy", "runtime_deterministic_replay", "reverse_observation_coverage"} {
		if !slices.Contains(r.CoreDimensions, id) {
			r.CoreDimensions = append(r.CoreDimensions, id)
		}
	}
	finishReceipt(r, o.Failure)
	return r
}

func axis(id string, n, d int, unit, reason, evidence string) completeness.CompletenessDimension {
	status := "UNKNOWN"
	if d > 0 && n > 0 {
		status = "PROGRESS"
		if n == d {
			status = "PASS"
		}
	}
	return completeness.CompletenessDimension{ID: id, Status: status, Numerator: n, Denominator: d, Unit: unit, Reason: reason, Evidence: []string{evidence}}
}
func btoi(v bool) int {
	if v {
		return 1
	}
	return 0
}
func replaceAxis(r *completeness.CompletenessReceipt, d completeness.CompletenessDimension) {
	for i, old := range r.Dimensions {
		if old.ID == d.ID {
			r.Dimensions[i] = d
			return
		}
	}
	r.Dimensions = append(r.Dimensions, d)
}
func runtimeResourceAxis(o Observation) completeness.CompletenessDimension {
	processes := append([]ProcessObservation{o.Toolchain, o.Build}, o.Runs...)
	if o.Artifact != nil && o.Artifact.Reused {
		processes = append([]ProcessObservation{o.Toolchain}, o.Runs...)
	}
	observed := 0
	for _, p := range processes {
		if p.ExitCode != nil && p.PeakRSSBytes != nil && p.WallNS > 0 {
			observed++
		}
	}
	denominator := 4
	unit := "toolchain/build/two-runtime child CPU, wall and peak RSS profiles"
	if o.Artifact != nil && o.Artifact.Reused {
		denominator = 3
		unit = "current toolchain/two-runtime child CPU, wall and peak RSS profiles; retained build excluded"
	}
	if o.ToolchainReference != nil && o.ToolchainReference.Reused {
		processes = append([]ProcessObservation{o.Build}, o.Runs...)
		denominator, unit = 3, "current build/two-runtime child CPU, wall and peak RSS profiles; retained toolchain check excluded"
		if o.Artifact != nil && o.Artifact.Reused {
			processes = o.Runs
			denominator, unit = 2, "current two-runtime child CPU, wall and peak RSS profiles; retained toolchain check and build excluded"
		}
		observed = 0
		for _, p := range processes {
			if p.ExitCode != nil && p.PeakRSSBytes != nil && p.WallNS > 0 {
				observed++
			}
		}
	}
	return axis("runtime_child_resources", observed, denominator, unit,
		"CPU is process user+system time; peak RSS is known on Linux/macOS. Parent compiler resources and whole-host utilization remain unobserved.", "runtime observation toolchain/build/runs")
}

func finishReceipt(r *completeness.CompletenessReceipt, failure string) {
	oldNext := map[string]string{}
	for _, c := range r.UnresolvedClaims {
		oldNext[c.ID] = c.NextOperation
	}
	r.UnresolvedClaims = make([]completeness.UnresolvedCompletenessClaim, 0)
	r.StatusCounts = map[string]int{"PASS": 0, "PROGRESS": 0, "UNKNOWN": 0, "FAIL_CLOSED": 0}
	corePass := true
	for _, d := range r.Dimensions {
		r.StatusCounts[d.Status]++
		if d.Status == "PASS" {
			continue
		}
		if slices.Contains(r.CoreDimensions, d.ID) {
			corePass = false
		}
		next := oldNext[d.ID]
		if next == "" {
			next = "CONTINUE_SOURCE_BOUND_RUNTIME_OBSERVATION_WITH_RECORDED_INPUTS_AND_FAILURE"
		}
		r.UnresolvedClaims = append(r.UnresolvedClaims, completeness.UnresolvedCompletenessClaim{ID: d.ID, Status: d.Status, Reason: d.Reason, NextOperation: next})
	}
	r.FirstUnresolved = nil
	if len(r.UnresolvedClaims) > 0 {
		c := r.UnresolvedClaims[0]
		r.FirstUnresolved = &c
	}
	r.AggregateCompletenessScore = nil
	if failure != "" {
		r.Decision = "FAIL_CLOSED"
		r.FailClosedReason = &failure
		return
	}
	if r.StatusCounts["FAIL_CLOSED"] > 0 || r.FailClosedReason != nil {
		r.Decision = "FAIL_CLOSED"
		return
	}
	r.Decision = "PROGRESS_WITHIN_DECLARED_RUNTIME_SCOPE"
	if corePass {
		r.Decision = "PASS_WITHIN_DECLARED_RUNTIME_SCOPE"
	}
	// Preserve only explicit finite claims; never infer success from absence.
	r.NotClaimed = append(r.NotClaimed, "model-training holdout independence", "whole-host CPU utilization")
}
