package bodycodegen

import (
	"encoding/json"
	"strconv"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

const typedPathCompletenessProfile = "gooo/body-codegen-typed-path-v1"

func completenessProfile(report Report) string {
	if report.BodyPaths != nil {
		return typedPathCompletenessProfile
	}
	return "gooo/body-codegen-pure-v1"
}

func typedPathSearchConfig(step int, feedback *typedPathFeedback) json.RawMessage {
	rounds, unfixed := 0, false
	var ci *pathplan.CIHint
	if feedback != nil {
		rounds, unfixed, ci = feedback.rounds, feedback.unfixed, feedback.ci
	}
	raw, _ := json.Marshal(map[string]any{
		"schema": "gooo/typed-path-search-config/v1", "step_attempts": step,
		"feedback_rounds": rounds, "feedback_unfixed": unfixed, "ci_hint": ci,
		"timeout_ms": bodyPathBudget.Milliseconds(),
	})
	return raw
}

// Failed emission still carries the original request and any actual search
// observations. No candidate score is substituted for an absent final score.
func PathFailureCompletenessReceipt(activity string, source []byte, failure string,
	path *BodyPathReceipt) *CompletenessReceipt {
	report := Report{Decision: "FAIL_CLOSED", Activity: activity, SourceDigest: digest(source), BodyPaths: path}
	populateCompletenessReceipt(&report, failure)
	return report.CompletenessReceipt
}

func bodyPathCompletenessDimensions(report Report) ([]CompletenessDimension, []string) {
	path := report.BodyPaths
	source := bodyPathSourceDimension(report)
	finite := bodyPathFiniteDimension(report)
	accounting := bodyPathAccountingDimension(path)
	search := path.Search
	declared := max(0, search.DeclaredCombinations)
	observation := completenessDimension("typed_path_candidate_observation", min(len(search.Attempts), declared), declared,
		"declared combinations attempted", "Attempts include rejected typed combinations; unexplored combinations remain explicit.",
		[]string{"body_paths.search.attempts", "body_paths.document_sha256:" + path.DocumentSHA256},
		len(search.Attempts) > declared || search.DeclaredCombinations < 0)
	scoring := completenessDimension("typed_path_candidate_scoring", min(max(0, search.Evaluated), declared), declared,
		"declared combinations with finite scores", "Type-rejected candidates have no functional score; search coverage is not intent completeness.",
		[]string{"body_paths.search.evaluated_candidates", "body_paths.search.type_rejected_candidates"},
		search.Evaluated < 0 || search.Evaluated > len(search.Attempts))
	if scoring.Status == "UNKNOWN" && len(search.Attempts) > 0 {
		scoring.Status = "PROGRESS"
	}
	dimensions := []CompletenessDimension{source, finite, accounting, observation, scoring}
	required := []string{source.ID, finite.ID, accounting.ID}
	if path.Observation != nil {
		d := pathObservationDimension(path)
		dimensions, required = append(dimensions, d), append(required, d.ID)
	}
	if path.Resolution != nil || path.Observation != nil && path.Observation.Options.ResolveUniqueCandidate {
		d := pathResolutionDimension(path)
		dimensions, required = append(dimensions, d), append(required, d.ID)
	}
	return dimensions, required
}

func bodyPathSourceDimension(report Report) CompletenessDimension {
	p := report.BodyPaths
	checks := []bool{validDigest(p.OriginalSourceSHA256), validDigest(p.DocumentSHA256),
		validDigest(p.TestSuiteSHA256), len(p.SearchConfig) > 0 && digest(p.SearchConfig) == p.SearchConfigSHA256,
		p.SourceBaseMatched && p.SourceBinding.Decision == "PASS" && p.SourceBinding.Equivalent &&
			validDigest(p.SourceBinding.SourceSemanticDigest) &&
			p.SourceBinding.SourceSemanticDigest == p.SourceBinding.GeneratedSemanticDigest,
		validDigest(p.SelectedSourceSHA256) && p.SelectedSourceSHA256 == report.SourceDigest,
	}
	observed := 0
	for _, check := range checks {
		observed += boolCount(check)
	}
	mismatch := p.SourceBinding.Decision == "FAIL_CLOSED" ||
		p.SelectedSourceSHA256 != "" && p.SelectedSourceSHA256 != report.SourceDigest ||
		len(p.SearchConfig) > 0 && digest(p.SearchConfig) != p.SearchConfigSHA256
	return completenessDimension("typed_path_source_binding", observed, len(checks),
		"original source, document, tests, search controls, fallback semantics and selected source bindings",
		"The declared document and original source authorize bounded alternatives; selected-source emission retains that chain.",
		[]string{"body_paths.original_source_sha256:" + p.OriginalSourceSHA256,
			"body_paths.document_sha256:" + p.DocumentSHA256, "body_paths.test_suite_sha256:" + p.TestSuiteSHA256,
			"body_paths.search_config_sha256:" + p.SearchConfigSHA256, "body_paths.source_binding",
			"body_paths.selected_source_sha256:" + p.SelectedSourceSHA256, "source_digest:" + report.SourceDigest}, mismatch)
}

func bodyPathFiniteDimension(report Report) CompletenessDimension {
	p := report.BodyPaths
	testSHA, expectedCount, observationMismatch := observedPathContract(p)
	total, passed := max(0, expectedCount), 0
	mismatch := p.DeclaredTestCases < 0 || observationMismatch
	cases := make([]pathplan.TestCase, len(p.NativeCases))
	for i, observed := range p.NativeCases {
		matched := observed.Actual == observed.Expected
		passed += boolCount(matched)
		mismatch = mismatch || observed.Passed != matched
		cases[i] = pathplan.TestCase{Input: observed.Input, Expected: observed.Expected}
	}
	if len(p.NativeCases) > 0 {
		raw, _ := json.Marshal(cases)
		selectedPassed, selectedTotal := bodyPathSelectedScore(p)
		mismatch = mismatch || digest(raw) != testSHA || len(cases) != total ||
			selectedTotal != total || selectedPassed != passed ||
			!validDigest(report.GeneratedDigest) || p.SelectedSourceSHA256 != report.SourceDigest
	}
	evidenceSuite := "body_paths.test_suite_sha256:" + testSHA
	if p.Observation != nil {
		evidenceSuite = "body_paths.observation.effective_cases_sha256:" + testSHA
	}
	d := completenessDimension("typed_path_finite_accuracy", min(passed, total), total,
		"declared selection cases matched by the emitted body in the bounded Go AST evaluator",
		"Observed actuals are checked against the declared suite and typed interpreter; selection tests do not prove unseen behavior or generated-package execution.",
		[]string{"body_paths.native_case_results", evidenceSuite,
			"generated_digest:" + report.GeneratedDigest, "evaluator:bounded_integer_go_ast"}, mismatch)
	if !mismatch && len(p.NativeCases) > 0 && passed == 0 {
		d.Status = "PROGRESS"
	}
	return d
}

func bodyPathCalls(p *BodyPathReceipt) (local, external int, known bool) {
	s := bodyPathSelection(p)
	if !p.SearchStarted && s.Schema == "" && s.ModelCalls == 0 && s.ExternalCalls == 0 {
		return 0, 0, true // Source/context/model-load failures precede the prediction entry point.
	}
	return s.ModelCalls, s.ExternalCalls, s.Schema != "" && s.ExternalCallsKnown
}

func bodyPathAccountingDimension(p *BodyPathReceipt) CompletenessDimension {
	local, external, known := bodyPathCalls(p)
	s := bodyPathSelection(p)
	pinned := local == 0 || validDigest("sha256:"+s.MetadataSHA256) && validDigest("sha256:"+s.WeightsSHA256)
	return completenessDimension("typed_path_provider_accounting", boolCount(known && pinned), 1,
		"local prediction counts, model pins and external request counts",
		"Counts executed predictions including feedback, separately from choices and candidate tests; an unobserved search result remains unknown.",
		[]string{"body_paths.search_started:" + strconv.FormatBool(p.SearchStarted),
			"local_model_predictions:" + strconv.Itoa(local), "external_provider_calls:" + strconv.Itoa(external),
			"external_provider_calls_known:" + strconv.FormatBool(known),
			"model_metadata_sha256:" + s.MetadataSHA256, "model_weights_sha256:" + s.WeightsSHA256},
		local < 0 || external < 0)
}

func bodyPathNetworkDimension(p *BodyPathReceipt) CompletenessDimension {
	local, external, known := bodyPathCalls(p)
	return completenessDimension("external_network_boundary", boolCount(known && external == 0), 1,
		"typed-path executions with no external provider requests",
		"This path invokes only local models and deterministic evaluators; prediction counts do not imply network calls.",
		[]string{"local_model_predictions:" + strconv.Itoa(local), "external_provider_calls:" + strconv.Itoa(external),
			"external_provider_calls_known:" + strconv.FormatBool(known)}, known && external != 0)
}

func bindPathCompletenessScope(receipt *CompletenessReceipt, report Report) {
	p := report.BodyPaths
	local, external, known := bodyPathCalls(p)
	s := bodyPathSelection(p)
	mode, provider := "deterministic_finite_tdd", "deterministic"
	if pathResolved(p) {
		mode = "unique_source_observation"
	} else if !p.SearchStarted {
		mode, provider = "not_started", "none"
	} else if !known {
		mode, provider = "unobserved_search_result", "UNKNOWN"
	} else if local > 0 {
		mode, provider = "local_prediction_then_finite_tdd", "tiny_go"
	} else if p.ModelContext != nil && p.ModelContext.Status == "DECLINED_TO_DETERMINISTIC" {
		mode = "context_declined_to_deterministic"
	}
	raw, _ := json.Marshal(p)
	contextSHA, contextStatus := "", "UNOBSERVED"
	if p.ModelContext != nil {
		contextBytes, _ := json.Marshal(p.ModelContext)
		contextSHA, contextStatus = digest(contextBytes), p.ModelContext.Status
	}
	receipt.ProfileID = typedPathCompletenessProfile
	receipt.Scope["domain_scope"] = "one source-bound Integer -> Integer activity, declared typed alternatives and finite selection cases"
	receipt.DecisionBasis += "; typed paths additionally require original/selected source binding, final finite-suite accuracy and provider accounting"
	receipt.Scope["emission_decision_mode"] = report.RouteDecision.Mode
	receipt.Scope["emission_decision_provider"] = report.RouteDecision.Provider
	receipt.Scope["laya_mode"], receipt.Scope["laya_provider"] = "", ""
	receipt.Scope["typed_path"] = map[string]any{
		"observation_sha256": digest(raw), "document_sha256": p.DocumentSHA256,
		"original_source_sha256": p.OriginalSourceSHA256, "selected_source_sha256": p.SelectedSourceSHA256,
		"test_suite_sha256": p.TestSuiteSHA256, "search_config_sha256": p.SearchConfigSHA256,
		"search_config":      p.SearchConfig,
		"search_plan_sha256": s.PlanSHA256, "decision_mode": mode, "decision_provider": provider,
		"local_model_requested": p.LocalModelRequested, "search_started": p.SearchStarted,
		"local_model_predictions": local, "external_provider_calls": external, "provider_counts_known": known,
		"model_metadata_sha256": s.MetadataSHA256, "model_weights_sha256": s.WeightsSHA256,
		"model_variant": s.ModelVariant, "context_sha256": contextSHA, "context_status": contextStatus,
		"case_evaluator": "bounded_integer_go_ast", "generated_package_executed": false,
	}
}
