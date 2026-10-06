package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const bodyFillPlanSchema = "gooo/body-codegen-ir-fill-plan/v1"
const bodyFillMultiPlanSchema = "gooo/body-codegen-ir-fill-plan/v2"
const bodyFillStateSchema = "gooo/body-codegen-ir-fill-state/v2"
const bodyFillEvaluator = "gooo/bodycodegen-int64-ast-interpreter/v2"
const irBodyFillDecisionBudget = 8 * time.Second

// IRBodyFillPlan supplies finite, typed expression candidates for explicit
// holes in a Gooo activity body. V1 has one hole; V2 selects one complete,
// declared assignment across several holes as a single model decision.
type IRBodyFillPlan struct {
	Schema           string                `json:"schema"`
	Intent           string                `json:"intent"`
	HoleID           string                `json:"hole_id"`
	Holes            []IRBodyFillHole      `json:"holes,omitempty"`
	ProviderModel    string                `json:"provider_model,omitempty"`
	Candidates       []IRBodyFillCandidate `json:"candidates"`
	TestCases        []IRBodyFillTestCase  `json:"test_cases"`
	HoldoutTestCases []IRBodyFillTestCase  `json:"holdout_test_cases,omitempty"`
}

type IRBodyFillHole struct {
	ID string `json:"id"`
}

type IRBodyFillCandidate struct {
	ID         string            `json:"id"`
	Expression string            `json:"expression,omitempty"`
	Fills      map[string]string `json:"fills,omitempty"`
}

// IRBodyFillTestCase is a finite, explicit integer contract used to measure
// the selected body's observed functional accuracy. It is not a proof over all
// int64 values.
type IRBodyFillTestCase struct {
	Input    int64   `json:"input"`
	Inputs   []int64 `json:"inputs,omitempty"`
	Expected int64   `json:"expected"`
}

func (testCase *IRBodyFillTestCase) UnmarshalJSON(data []byte) error {
	var fields struct {
		Input    *int64  `json:"input"`
		Inputs   []int64 `json:"inputs"`
		Expected *int64  `json:"expected"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fields); err != nil {
		return err
	}
	if fields.Expected == nil || fields.Input == nil && len(fields.Inputs) == 0 || len(fields.Inputs) > 16 {
		return fmt.Errorf("IR body-fill test case requires explicit integer input(s) and expected value")
	}
	if fields.Input == nil {
		fields.Input = new(int64)
		*fields.Input = fields.Inputs[0]
	} else if len(fields.Inputs) > 0 && fields.Inputs[0] != *fields.Input {
		return fmt.Errorf("IR body-fill test case input must equal the first inputs value")
	}
	*testCase = IRBodyFillTestCase{Input: *fields.Input, Inputs: append([]int64(nil), fields.Inputs...), Expected: *fields.Expected}
	return nil
}

func (testCase IRBodyFillTestCase) MarshalJSON() ([]byte, error) {
	if len(testCase.Inputs) > 0 {
		return json.Marshal(struct {
			Inputs   []int64 `json:"inputs"`
			Expected int64   `json:"expected"`
		}{Inputs: testCase.Inputs, Expected: testCase.Expected})
	}
	return json.Marshal(struct {
		Input    int64 `json:"input"`
		Expected int64 `json:"expected"`
	}{Input: testCase.Input, Expected: testCase.Expected})
}

func (testCase IRBodyFillTestCase) inputValues() []int64 {
	if len(testCase.Inputs) > 0 {
		return testCase.Inputs
	}
	return []int64{testCase.Input}
}

type IRBodyFillCandidateScore struct {
	ID              string  `json:"id"`
	Expression      string  `json:"expression"`
	TypecheckPassed bool    `json:"typecheck_passed"`
	TestCasesPassed int     `json:"test_cases_passed"`
	TestCasesTotal  int     `json:"test_cases_total"`
	AccuracyPercent float64 `json:"accuracy_percent"`
}

type IRBodyFillCaseResult struct {
	Input    int64   `json:"input"`
	Inputs   []int64 `json:"inputs,omitempty"`
	Expected int64   `json:"expected"`
	Actual   int64   `json:"actual"`
	Passed   bool    `json:"passed"`
}

type IRBodyFillTiming struct {
	IRPlanBuildMS      float64  `json:"ir_plan_build_ms"`
	BehavioralProbeMS  float64  `json:"behavioral_probe_ms"`
	TinyModelLoadMS    *float64 `json:"tiny_model_load_ms,omitempty"`
	ProviderDecisionMS float64  `json:"provider_decision_ms"`
	LayaDecisionMS     float64  `json:"laya_decision_ms"`
	TinyDecisionMS     float64  `json:"tiny_decision_ms"`
	FinalEmissionMS    float64  `json:"final_emission_ms"`
	TotalMS            float64  `json:"total_ms"`
	ExecutionModel     string   `json:"execution_model"`
	DecisionStage      string   `json:"decision_stage"`
}

type IRBodyFillBehavioralProbeReceipt struct {
	Schema                                 string                            `json:"schema"`
	ProbeInputs                            []int64                           `json:"probe_inputs"`
	ProbeVectors                           [][]int64                         `json:"probe_vectors,omitempty"`
	ProbeProfileSHA256                     string                            `json:"probe_profile_sha256"`
	CandidateProfiles                      []IRBodyFillCandidateProbeProfile `json:"candidate_profiles"`
	ProbeInputsTotal                       int                               `json:"probe_inputs_total"`
	ProbeInputsOmitted                     int                               `json:"probe_inputs_omitted"`
	CandidateCount                         int                               `json:"candidate_count"`
	CandidateRunsCompleted                 int                               `json:"candidate_runs_completed"`
	CandidateRunsFailed                    int                               `json:"candidate_runs_failed"`
	CandidatePairsDistinguished            int                               `json:"candidate_pairs_distinguished"`
	CandidatePairsEvaluated                int                               `json:"candidate_pairs_evaluated"`
	CandidatePairsTotal                    int                               `json:"candidate_pairs_total"`
	CandidatePairDistinguishabilityPercent float64                           `json:"candidate_pair_distinguishability_percent"`
	ProbeInputsWithDisagreement            int                               `json:"probe_inputs_with_disagreement"`
	Scope                                  string                            `json:"scope"`
}

type IRBodyFillCandidateProbeProfile struct {
	CandidateID string  `json:"candidate_id"`
	Outputs     []int64 `json:"outputs"`
}

type IRBodyFillReceipt struct {
	Schema                     string                                `json:"schema"`
	Intent                     string                                `json:"intent"`
	HoleID                     string                                `json:"hole_id"`
	HoleToken                  string                                `json:"hole_token"`
	IRPlanSHA256               string                                `json:"ir_plan_sha256"`
	ProposedCandidateID        string                                `json:"proposed_candidate_id"`
	ProposedAccuracyPct        float64                               `json:"proposed_accuracy_percent"`
	SelectedCandidateID        string                                `json:"selected_candidate_id"`
	SelectedExpression         string                                `json:"selected_expression"`
	HoleFills                  []IRBodyFillHoleFill                  `json:"hole_fills,omitempty"`
	CandidateGeneration        *IRBodyFillCandidateGenerationReceipt `json:"candidate_generation,omitempty"`
	BestCandidateID            string                                `json:"best_candidate_id"`
	BestAccuracyPercent        float64                               `json:"best_candidate_accuracy_percent"`
	SelectionRegretPP          float64                               `json:"selection_regret_percentage_points"`
	SelectionAdjustment        string                                `json:"selection_adjustment"`
	Decision                   decisionroute.Receipt                 `json:"decision"`
	CandidateScores            []IRBodyFillCandidateScore            `json:"candidate_scores"`
	BehavioralProbes           *IRBodyFillBehavioralProbeReceipt     `json:"behavioral_probes,omitempty"`
	TestSuiteSHA256            string                                `json:"test_suite_sha256"`
	TestCasesPassed            int                                   `json:"test_cases_passed"`
	TestCasesTotal             int                                   `json:"test_cases_total"`
	FunctionalAccuracyPct      float64                               `json:"functional_accuracy_percent"`
	HoldoutSuiteSHA256         string                                `json:"holdout_suite_sha256,omitempty"`
	HoldoutCasesPassed         int                                   `json:"holdout_cases_passed"`
	HoldoutCasesTotal          int                                   `json:"holdout_cases_total"`
	HoldoutAccuracyPercent     *float64                              `json:"holdout_accuracy_percent"`
	HoldoutCaseResults         []IRBodyFillCaseResult                `json:"holdout_case_results,omitempty"`
	LocalModelPredictions      int                                   `json:"local_model_predictions"`
	ExternalProviderCalls      int                                   `json:"external_provider_calls"`
	ExternalProviderCallsKnown bool                                  `json:"external_provider_calls_known"`
	Evaluator                  string                                `json:"evaluator"`
	SelectedCaseResults        []IRBodyFillCaseResult                `json:"selected_case_results"`
	AccuracyScope              string                                `json:"accuracy_scope"`
	Timing                     IRBodyFillTiming                      `json:"timing"`
}

type IRBodyFillHoleFill struct {
	HoleID     string `json:"hole_id"`
	Expression string `json:"expression"`
}

type irBodyFillState struct {
	Schema           string                            `json:"schema"`
	Stage            string                            `json:"stage"`
	Activity         string                            `json:"activity"`
	ActivityID       string                            `json:"activity_id"`
	InputType        string                            `json:"input_type"`
	InputTypes       []string                          `json:"input_types,omitempty"`
	OutputType       string                            `json:"output_type"`
	Intent           string                            `json:"intent"`
	HoleID           string                            `json:"hole_id"`
	HoleIDs          []string                          `json:"hole_ids,omitempty"`
	BodyIR           string                            `json:"body_ir"`
	TestCaseCount    int                               `json:"test_case_count"`
	TestSuiteSHA256  string                            `json:"test_suite_sha256"`
	Candidates       []IRBodyFillCandidateScore        `json:"candidate_scores"`
	BehavioralProbes *IRBodyFillBehavioralProbeReceipt `json:"behavioral_probes,omitempty"`
}

// GenerateWithIRBodyFill builds the typed hole plan synchronously from a Gooo
// activity, calls the existing Laya/default path once after that plan is
// complete, fills its declared hole assignment, and only then emits the final
// Go projection. Both plan versions target one-to-sixteen-input Integer bodies.
func GenerateWithIRBodyFill(
	ctx context.Context,
	filename string,
	source []byte,
	activityName string,
	plan IRBodyFillPlan,
	endpoint, apiKey string,
) (Result, error) {
	return GenerateWithIRBodyFillWithOptions(ctx, filename, source, activityName, plan, endpoint, apiKey, IRBodyFillOptions{})
}

// GenerateWithIRBodyFillWithOptions preserves the legacy Laya path when
// TinyGoProvider is nil and adds an explicit opt-in local model path otherwise.
func GenerateWithIRBodyFillWithOptions(
	ctx context.Context,
	filename string,
	source []byte,
	activityName string,
	plan IRBodyFillPlan,
	endpoint, apiKey string,
	options IRBodyFillOptions,
) (Result, error) {
	var tinyProvider tinyGoBodyFillResolver
	if options.TinyGoProvider != nil {
		tinyProvider = options.TinyGoProvider
	}
	return generateWithIRBodyFillOptions(ctx, filename, source, activityName, plan, endpoint, apiKey,
		options, tinyProvider)
}

func generateWithIRBodyFillOptions(
	ctx context.Context,
	filename string,
	source []byte,
	activityName string,
	plan IRBodyFillPlan,
	endpoint, apiKey string,
	options IRBodyFillOptions,
	tinyProvider tinyGoBodyFillResolver,
) (Result, error) {
	totalStarted := time.Now()
	if ctx == nil {
		return Result{}, fmt.Errorf("body-fill context is required")
	}
	usingTinyGo := tinyProvider != nil
	if usingTinyGo {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
	}
	if err := validateIRBodyFillPlan(plan); err != nil {
		return Result{}, err
	}
	if err := decisionroute.ValidateProviderModel(plan.ProviderModel); err != nil {
		return Result{}, fmt.Errorf("body-fill provider model: %w", err)
	}
	if usingTinyGo && (endpoint != "" || apiKey != "" || plan.ProviderModel != "") {
		return Result{}, fmt.Errorf("tiny_go body fill cannot be combined with Laya endpoint, API key, or provider model")
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Result{}, fmt.Errorf("parse .gooo source: %w", diagnostics.Error())
	}
	if file == nil || file.Package == nil {
		return Result{}, fmt.Errorf(".gooo source has no package declaration")
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == activityName {
			activity = candidate
			break
		}
	}
	if activity == nil {
		return Result{}, fmt.Errorf("activity %q was not found", activityName)
	}
	if !activity.ValueProgramPresent || activity.ValueProgram == "" {
		return Result{}, fmt.Errorf("activity %q has no computes body", activityName)
	}
	if len(activity.Inputs) < 1 || len(activity.Inputs) > 16 || activity.Output != "Integer" {
		return Result{}, fmt.Errorf("IR body fill requires 1..16 Integer inputs and one Integer output")
	}
	for _, input := range activity.Inputs {
		if input.Name != "Integer" {
			return Result{}, fmt.Errorf("IR body fill currently requires Integer inputs and one Integer output")
		}
	}
	for _, testCase := range append(append([]IRBodyFillTestCase(nil), plan.TestCases...), plan.HoldoutTestCases...) {
		if len(testCase.inputValues()) != len(activity.Inputs) {
			return Result{}, fmt.Errorf("IR body-fill case input arity %d does not match activity input arity %d", len(testCase.inputValues()), len(activity.Inputs))
		}
	}
	modelDocument, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return Result{}, fmt.Errorf("lower activity identity: %w", err)
	}
	model, err := bidir.Get(modelDocument)
	if err != nil {
		return Result{}, fmt.Errorf("resolve activity identity: %w", err)
	}
	activityID := ""
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == activityName {
			activityID = string(node.ID)
			break
		}
	}
	if activityID == "" {
		return Result{}, fmt.Errorf("activity %q has no stable semantic identity", activityName)
	}

	body, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil {
		return Result{}, err
	}
	holes := bodyFillPlanHoles(plan)
	for _, hole := range holes {
		holeToken := bodyFillHoleToken(hole.ID)
		if countIdentifier(body, holeToken) != 1 {
			return Result{}, fmt.Errorf("activity %q must contain exactly one IR body hole %q", activityName, holeToken)
		}
	}
	planStarted := time.Now()
	scores := make([]IRBodyFillCandidateScore, 0, len(plan.Candidates))
	candidateBodies := make(map[string]string, len(plan.Candidates))
	candidateSources := make(map[string][]byte, len(plan.Candidates))
	for _, candidate := range plan.Candidates {
		fills := bodyFillCandidateFills(plan, candidate)
		candidateBody := body
		for _, hole := range holes {
			expression := fills[hole.ID]
			candidateBody, err = replaceIdentifier(candidateBody, bodyFillHoleToken(hole.ID), expression)
			if err != nil {
				return Result{}, fmt.Errorf("fill candidate %q at hole %q: %w", candidate.ID, hole.ID, err)
			}
		}
		generated, err := generateRouteParameters(
			file.Package.Name, activityName, activityID, bodyFillInputParameters(activity.Inputs), "int64", candidateBody, preserveRoute,
		)
		if err != nil {
			return Result{}, fmt.Errorf("candidate %q is not a valid typed body: %w", candidate.ID, err)
		}
		_, passed, err := evaluateIntegerCases(generated.source, activityName, plan.TestCases)
		if err != nil {
			return Result{}, fmt.Errorf("evaluate candidate %q: %w", candidate.ID, err)
		}
		candidateBodies[candidate.ID] = candidateBody
		candidateSources[candidate.ID] = []byte(generated.source)
		accuracy := float64(passed) * 100 / float64(len(plan.TestCases))
		candidateExpression := candidate.Expression
		if plan.Schema == bodyFillMultiPlanSchema {
			candidateExpression = bodyFillCandidateDescription(holes, fills)
		}
		scores = append(scores, IRBodyFillCandidateScore{
			ID: candidate.ID, Expression: candidateExpression, TypecheckPassed: true,
			TestCasesPassed: passed, TestCasesTotal: len(plan.TestCases), AccuracyPercent: accuracy,
		})
	}
	behavioralProbeStarted := time.Now()
	behavioralProbes, err := measureIRBodyFillBehavioralProbes(activityName, candidateSources, plan.TestCases, plan.HoldoutTestCases)
	if err != nil {
		return Result{}, fmt.Errorf("measure candidate behavior on synthetic probes: %w", err)
	}
	behavioralProbeMS := float64(time.Since(behavioralProbeStarted)) / float64(time.Millisecond)
	planBuildMS := float64(time.Since(planStarted)) / float64(time.Millisecond)
	testBytes, _ := json.Marshal(plan.TestCases)
	testSuiteSHA256 := digest(testBytes)
	stateBytes, err := json.Marshal(irBodyFillState{
		Schema: bodyFillStateSchema, Stage: "ir_ready_before_body_emission",
		Activity: activityName, ActivityID: activityID, InputType: "Integer", InputTypes: bodyFillInputTypes(activity.Inputs), OutputType: "Integer",
		Intent: plan.Intent, HoleID: bodyFillHoleSummary(holes), HoleIDs: bodyFillHoleIDs(plan, holes), BodyIR: body,
		TestCaseCount: len(plan.TestCases), TestSuiteSHA256: testSuiteSHA256, Candidates: scores,
		BehavioralProbes: behavioralProbes,
	})
	if err != nil {
		return Result{}, fmt.Errorf("encode Gooo IR body-fill state: %w", err)
	}
	requestOptions := make([]decisionroute.Option, 0, len(plan.Candidates))
	modelCandidates := bodyFillModelCandidates(holes, plan.Candidates)
	if usingTinyGo {
		primaryHole := ""
		if plan.Schema == bodyFillMultiPlanSchema && len(holes) > 0 {
			primaryHole = holes[0].ID
		}
		requestOptions, err = tinyGoBodyFillOptionsForHole(plan.Candidates, primaryHole)
		if err != nil {
			return Result{}, err
		}
	} else {
		for _, candidate := range modelCandidates {
			score := scoreByID(scores, candidate.ID)
			requestOptions = append(requestOptions, decisionroute.Option{
				ID: candidate.ID,
				Description: fmt.Sprintf("Emit exactly this expression: %s. It passes %d of %d declared test cases (%.2f%%).",
					candidate.Expression, score.TestCasesPassed, score.TestCasesTotal, score.AccuracyPercent),
			})
		}
	}
	instructions := "Fill the single typed expression hole in the supplied Gooo body IR. Choose only a listed candidate. " +
		"Use the intent, declared training evidence, and generated candidate behavior profiles; probes have no expected answers. " +
		"Do not invent code or modify any other IR node."
	if plan.Schema == bodyFillMultiPlanSchema {
		instructions = "Fill every typed expression hole in the supplied Gooo body IR using one complete listed candidate assignment. " +
			"Choose only a listed assignment. Use the intent, declared training evidence, and generated candidate behavior profiles; " +
			"probes have no expected answers. Do not invent code or modify any other IR node."
	}
	request := decisionroute.Request{
		Schema: decisionroute.RequestSchema, State: string(stateBytes),
		ProviderModel: plan.ProviderModel,
		Question: decisionroute.Question{
			ID:           "body_ir_fill",
			Instructions: instructions,
			Options:      requestOptions,
		},
		Fallback: plan.Candidates[0].ID,
	}
	if usingTinyGo {
		request.Intent = plan.Intent
	}
	if usingTinyGo {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
	}
	decisionStarted := time.Now()
	decisionContext, cancel := context.WithTimeout(ctx, irBodyFillDecisionBudget)
	var decision decisionroute.Receipt
	if usingTinyGo {
		decision, err = tinyProvider.Resolve(decisionContext, request)
	} else {
		decision, err = decisionroute.Resolve(decisionContext, request, endpoint, apiKey)
	}
	cancel()
	decisionMS := float64(time.Since(decisionStarted)) / float64(time.Millisecond)
	if err != nil {
		return Result{}, fmt.Errorf("select Gooo IR body-fill candidate: %w", err)
	}
	if usingTinyGo && decision.Provider != decisionroute.ProviderTinyGo {
		return Result{}, fmt.Errorf("tiny_go chooser returned provider %q", decision.Provider)
	}
	if usingTinyGo {
		expectedRequestSHA256, validateErr := decisionroute.Validate(request)
		if validateErr != nil {
			return Result{}, fmt.Errorf("validate typed tiny_go request: %w", validateErr)
		}
		if !validTinyGoDecisionReceipt(decision, expectedRequestSHA256, request) {
			return Result{}, fmt.Errorf("tiny_go chooser returned incomplete or mismatched model provenance")
		}
	}
	if usingTinyGo {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
	}
	proposed, ok := candidateByID(modelCandidates, decision.Selected)
	if !ok {
		return Result{}, fmt.Errorf("body-fill decision selected undeclared candidate %q", decision.Selected)
	}
	best := bestBodyFillCandidate(scores)
	proposedScore := scoreByID(scores, proposed.ID)
	selected := proposed
	selectionAdjustment := "proposal_retained"
	if proposedScore.TestCasesPassed < best.TestCasesPassed {
		selected, _ = candidateByID(modelCandidates, best.ID)
		selectionAdjustment = "replaced_with_best_scoring_candidate"
	}
	selectedBody := candidateBodies[selected.ID]
	originalSelected, ok := candidateByID(plan.Candidates, selected.ID)
	if !ok {
		return Result{}, fmt.Errorf("selected body-fill candidate %q is absent from the source plan", selected.ID)
	}
	completedSource, err := replaceActivityProgram(source, activity.ValueProgramSpan, selectedBody)
	if err != nil {
		return Result{}, err
	}
	emissionStarted := time.Now()
	result, err := GenerateWithPlanner(ctx, filename, completedSource, activityName, "", "")
	emissionMS := float64(time.Since(emissionStarted)) / float64(time.Millisecond)
	if err != nil {
		return Result{}, fmt.Errorf("emit selected Gooo IR body: %w", err)
	}
	result.GoooSource = string(completedSource)
	caseResults, passed, err := evaluateIntegerCases([]byte(result.Source), activityName, plan.TestCases)
	if err != nil {
		return Result{}, fmt.Errorf("evaluate selected generated body: %w", err)
	}
	accuracy := float64(passed) * 100 / float64(len(plan.TestCases))
	var holdoutResults []IRBodyFillCaseResult
	var holdoutAccuracy *float64
	var holdoutSuiteSHA256 string
	holdoutPassed := 0
	if len(plan.HoldoutTestCases) > 0 {
		holdoutResults, holdoutPassed, err = evaluateIntegerCases([]byte(result.Source), activityName, plan.HoldoutTestCases)
		if err != nil {
			return Result{}, fmt.Errorf("evaluate selected generated body on held-out cases: %w", err)
		}
		holdoutBytes, _ := json.Marshal(plan.HoldoutTestCases)
		holdoutSuiteSHA256 = digest(holdoutBytes)
		value := float64(holdoutPassed) * 100 / float64(len(plan.HoldoutTestCases))
		holdoutAccuracy = &value
	}
	localPredictions, externalCalls, externalCallsKnown := bodyFillProviderAccounting(decision)
	tinyDecisionMS, layaDecisionMS := 0.0, 0.0
	var tinyModelLoadMS *float64
	switch decision.Provider {
	case decisionroute.ProviderTinyGo:
		tinyDecisionMS = decisionMS
		tinyModelLoadMS = options.TinyModelLoadMS
	case "laya":
		layaDecisionMS = decisionMS
	}
	planBytes, _ := json.Marshal(plan)
	var holeFills []IRBodyFillHoleFill
	if plan.Schema == bodyFillMultiPlanSchema {
		holeFills = bodyFillHoleResults(holes, bodyFillCandidateFills(plan, originalSelected))
	}
	result.Report.BodyFill = &IRBodyFillReceipt{
		Schema: plan.Schema, Intent: plan.Intent, HoleID: bodyFillHoleSummary(holes),
		HoleToken: bodyFillHoleToken(holes[0].ID), IRPlanSHA256: digest(planBytes),
		ProposedCandidateID: proposed.ID, ProposedAccuracyPct: proposedScore.AccuracyPercent,
		SelectedCandidateID: selected.ID, SelectedExpression: selected.Expression,
		HoleFills:       holeFills,
		BestCandidateID: best.ID, BestAccuracyPercent: best.AccuracyPercent,
		SelectionRegretPP:   best.AccuracyPercent - proposedScore.AccuracyPercent,
		SelectionAdjustment: selectionAdjustment,
		Decision:            decision, CandidateScores: scores,
		BehavioralProbes: behavioralProbes,
		TestSuiteSHA256:  testSuiteSHA256, TestCasesPassed: passed,
		TestCasesTotal: len(plan.TestCases), FunctionalAccuracyPct: accuracy,
		HoldoutSuiteSHA256: holdoutSuiteSHA256, HoldoutCasesPassed: holdoutPassed,
		HoldoutCasesTotal: len(plan.HoldoutTestCases), HoldoutAccuracyPercent: holdoutAccuracy,
		HoldoutCaseResults:    holdoutResults,
		LocalModelPredictions: localPredictions, ExternalProviderCalls: externalCalls,
		ExternalProviderCallsKnown: externalCallsKnown,
		Evaluator:                  bodyFillEvaluator,
		SelectedCaseResults:        caseResults,
		AccuracyScope:              "exact observed accuracy over the declared finite test suite; not a full-domain proof",
		Timing: IRBodyFillTiming{
			IRPlanBuildMS: planBuildMS, BehavioralProbeMS: behavioralProbeMS, TinyModelLoadMS: tinyModelLoadMS,
			ProviderDecisionMS: decisionMS,
			LayaDecisionMS:     layaDecisionMS, TinyDecisionMS: tinyDecisionMS,
			FinalEmissionMS: emissionMS,
			TotalMS:         float64(time.Since(totalStarted)) / float64(time.Millisecond),
			ExecutionModel:  "synchronous_sequential_no_background_codegen_goroutines",
			DecisionStage:   "after_typed_ir_plan_training_scores_and_behavior_probes_before_final_emission",
		},
	}
	populateCompletenessReceipt(&result.Report, "")
	return result, nil
}

func validateIRBodyFillPlan(plan IRBodyFillPlan) error {
	if plan.Schema != bodyFillPlanSchema && plan.Schema != bodyFillMultiPlanSchema {
		return fmt.Errorf("IR body-fill plan schema must be %q or %q", bodyFillPlanSchema, bodyFillMultiPlanSchema)
	}
	if strings.TrimSpace(plan.Intent) == "" || utf8.RuneCountInString(plan.Intent) > 2000 {
		return fmt.Errorf("IR body-fill intent must contain 1..2000 characters")
	}
	holes := bodyFillPlanHoles(plan)
	if plan.Schema == bodyFillPlanSchema {
		if !validBodyFillIdentifier(plan.HoleID) || len(plan.Holes) != 0 {
			return fmt.Errorf("IR body-fill v1 requires one valid hole_id and no holes array")
		}
	} else if plan.HoleID != "" || len(holes) < 2 || len(holes) > 8 {
		return fmt.Errorf("IR body-fill v2 requires 2..8 holes and no hole_id")
	}
	seenHoles := make(map[string]bool, len(holes))
	for _, hole := range holes {
		if !validBodyFillIdentifier(hole.ID) || seenHoles[hole.ID] {
			return fmt.Errorf("IR body-fill hole id %q is invalid or duplicated", hole.ID)
		}
		seenHoles[hole.ID] = true
	}
	if len(plan.Candidates) < 2 || len(plan.Candidates) > 16 {
		return fmt.Errorf("IR body-fill plan requires 2..16 candidates")
	}
	if len(plan.TestCases) == 0 || len(plan.TestCases) > 4096 || len(plan.HoldoutTestCases) > 4096 {
		return fmt.Errorf("IR body-fill plan requires 1..4096 training cases and at most 4096 holdout cases")
	}
	caseArity := 0
	trainingInputs := make(map[string]bool, len(plan.TestCases))
	for _, testCase := range plan.TestCases {
		inputs := testCase.inputValues()
		if len(inputs) < 1 || len(inputs) > 16 || caseArity != 0 && len(inputs) != caseArity {
			return fmt.Errorf("IR body-fill training cases must provide a consistent 1..16 input values")
		}
		caseArity = len(inputs)
		trainingInputs[bodyFillInputKey(inputs)] = true
	}
	for _, testCase := range plan.HoldoutTestCases {
		inputs := testCase.inputValues()
		if len(inputs) != caseArity {
			return fmt.Errorf("IR body-fill holdout cases must match training input arity %d", caseArity)
		}
		if trainingInputs[bodyFillInputKey(inputs)] {
			return fmt.Errorf("holdout inputs %v also appear in training cases", inputs)
		}
	}
	seen := make(map[string]bool, len(plan.Candidates))
	for _, candidate := range plan.Candidates {
		if !validBodyFillIdentifier(candidate.ID) || seen[candidate.ID] {
			return fmt.Errorf("IR body-fill candidate id %q is invalid or duplicated", candidate.ID)
		}
		seen[candidate.ID] = true
		fills := bodyFillCandidateFills(plan, candidate)
		if len(fills) != len(holes) {
			return fmt.Errorf("IR body-fill candidate %q must fill every declared hole exactly once", candidate.ID)
		}
		if plan.Schema == bodyFillPlanSchema && len(candidate.Fills) != 0 {
			return fmt.Errorf("IR body-fill v1 candidate %q cannot declare a fills map", candidate.ID)
		}
		if plan.Schema == bodyFillMultiPlanSchema && candidate.Expression != "" {
			return fmt.Errorf("IR body-fill v2 candidate %q must use fills, not expression", candidate.ID)
		}
		for _, hole := range holes {
			expressionSource, ok := fills[hole.ID]
			if !ok || strings.TrimSpace(expressionSource) == "" || utf8.RuneCountInString(expressionSource) > 512 {
				return fmt.Errorf("IR body-fill candidate %q expression for hole %q must contain 1..512 characters", candidate.ID, hole.ID)
			}
			for _, declaredHole := range holes {
				if countIdentifier(expressionSource, bodyFillHoleToken(declaredHole.ID)) != 0 {
					return fmt.Errorf("IR body-fill candidate %q expression for hole %q cannot introduce another hole token", candidate.ID, hole.ID)
				}
			}
			expression, err := parser.ParseExpr(expressionSource)
			if err != nil {
				return fmt.Errorf("parse IR body-fill candidate %q hole %q: %w", candidate.ID, hole.ID, err)
			}
			if err := validateExpression(expression); err != nil {
				return fmt.Errorf("IR body-fill candidate %q hole %q uses an unsupported expression: %w", candidate.ID, hole.ID, err)
			}
		}
	}
	return nil
}

func bodyFillInputTypes(inputs []syntax.NameRef) []string {
	if len(inputs) < 2 {
		return nil
	}
	types := make([]string, len(inputs))
	for index, input := range inputs {
		types[index] = input.Name
	}
	return types
}

func bodyFillInputParameters(inputs []syntax.NameRef) []InputParameter {
	parameters := make([]InputParameter, len(inputs))
	for index := range inputs {
		name := "input"
		if len(inputs) > 1 {
			name = fmt.Sprintf("input%d", index)
		}
		parameters[index] = InputParameter{Name: name, Type: "int64"}
	}
	return parameters
}

func bodyFillInputKey(inputs []int64) string {
	encoded, _ := json.Marshal(inputs)
	return string(encoded)
}

func bodyFillCaseKey(testCase IRBodyFillTestCase) string {
	return fmt.Sprintf("%s:%d", bodyFillInputKey(testCase.inputValues()), testCase.Expected)
}

func equalIRBodyFillCaseResults(left, right []IRBodyFillCaseResult) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].Input != right[index].Input || !slices.Equal(left[index].Inputs, right[index].Inputs) ||
			left[index].Expected != right[index].Expected || left[index].Actual != right[index].Actual ||
			left[index].Passed != right[index].Passed {
			return false
		}
	}
	return true
}

func bodyFillPlanHoles(plan IRBodyFillPlan) []IRBodyFillHole {
	if plan.Schema == bodyFillPlanSchema {
		return []IRBodyFillHole{{ID: plan.HoleID}}
	}
	return plan.Holes
}

func bodyFillCandidateFills(plan IRBodyFillPlan, candidate IRBodyFillCandidate) map[string]string {
	if plan.Schema == bodyFillPlanSchema {
		return map[string]string{plan.HoleID: candidate.Expression}
	}
	return candidate.Fills
}

func bodyFillCandidateDescription(holes []IRBodyFillHole, fills map[string]string) string {
	parts := make([]string, 0, len(holes))
	for _, hole := range holes {
		parts = append(parts, hole.ID+"="+fills[hole.ID])
	}
	return strings.Join(parts, "; ")
}

func bodyFillHoleSummary(holes []IRBodyFillHole) string {
	ids := make([]string, 0, len(holes))
	for _, hole := range holes {
		ids = append(ids, hole.ID)
	}
	return strings.Join(ids, ",")
}

func bodyFillHoleIDs(plan IRBodyFillPlan, holes []IRBodyFillHole) []string {
	if plan.Schema == bodyFillPlanSchema {
		return nil
	}
	ids := make([]string, 0, len(holes))
	for _, hole := range holes {
		ids = append(ids, hole.ID)
	}
	return ids
}

func bodyFillModelCandidates(holes []IRBodyFillHole, candidates []IRBodyFillCandidate) []IRBodyFillCandidate {
	result := make([]IRBodyFillCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if len(holes) == 1 && len(candidate.Fills) == 0 {
			result = append(result, candidate)
			continue
		}
		fills := make(map[string]string, len(candidate.Fills))
		maps.Copy(fills, candidate.Fills)
		if len(fills) == 0 && len(holes) == 1 {
			fills[holes[0].ID] = candidate.Expression
		}
		result = append(result, IRBodyFillCandidate{ID: candidate.ID, Expression: bodyFillCandidateDescription(holes, fills)})
	}
	return result
}

func bodyFillHoleResults(holes []IRBodyFillHole, fills map[string]string) []IRBodyFillHoleFill {
	result := make([]IRBodyFillHoleFill, 0, len(holes))
	for _, hole := range holes {
		result = append(result, IRBodyFillHoleFill{HoleID: hole.ID, Expression: fills[hole.ID]})
	}
	return result
}

func bodyFillHoleToken(id string) string { return "__GOOO_BODY_HOLE_" + id + "__" }

func validBodyFillIdentifier(value string) bool {
	if value == "" || len(value) > 32 {
		return false
	}
	for index, char := range value {
		valid := char == '_' || char >= 'a' && char <= 'z' || char >= '0' && char <= '9' && index > 0
		if !valid || index == 0 && (char < 'a' || char > 'z') {
			return false
		}
	}
	return true
}

func countIdentifier(body, name string) int {
	fset := token.NewFileSet()
	file := fset.AddFile("computes", fset.Base(), len(body))
	var sourceScanner scanner.Scanner
	sourceScanner.Init(file, []byte(body), nil, scanner.ScanComments)
	count := 0
	for {
		_, kind, literal := sourceScanner.Scan()
		if kind == token.EOF {
			return count
		}
		if kind == token.IDENT && literal == name {
			count++
		}
	}
}

func replaceIdentifier(body, name, replacement string) (string, error) {
	fset := token.NewFileSet()
	file := fset.AddFile("computes", fset.Base(), len(body))
	var sourceScanner scanner.Scanner
	sourceScanner.Init(file, []byte(body), nil, scanner.ScanComments)
	start, end, count := -1, -1, 0
	for {
		position, kind, literal := sourceScanner.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.IDENT && literal == name {
			start = file.Offset(position)
			end = start + len(literal)
			count++
		}
	}
	if count != 1 || start < 0 || end > len(body) {
		return "", fmt.Errorf("expected one typed IR hole %q, found %d", name, count)
	}
	expression, err := parser.ParseExpr(replacement)
	if err != nil {
		return "", fmt.Errorf("parse replacement expression: %w", err)
	}
	var formatted strings.Builder
	if err := format.Node(&formatted, token.NewFileSet(), expression); err != nil {
		return "", fmt.Errorf("format replacement expression: %w", err)
	}
	replacement = formatted.String()
	switch expression.(type) {
	case *ast.BinaryExpr, *ast.UnaryExpr:
		// A hole denotes one expression, so its caller cannot change the
		// candidate's precedence or merge adjacent unary operator tokens.
		replacement = "(" + replacement + ")"
	}
	return body[:start] + replacement + body[end:], nil
}

func replaceActivityProgram(source []byte, span syntax.Span, body string) ([]byte, error) {
	start, end := span.Start.Offset, span.End.Offset
	if start < 0 || end < start || end > len(source) || start == end {
		return nil, fmt.Errorf("activity computes body has an invalid source span")
	}
	quoted := strconv.Quote(body)
	updated := make([]byte, 0, len(source)-span.Len()+len(quoted))
	updated = append(updated, source[:start]...)
	updated = append(updated, quoted...)
	updated = append(updated, source[end:]...)
	return updated, nil
}

func scoreByID(scores []IRBodyFillCandidateScore, id string) IRBodyFillCandidateScore {
	for _, score := range scores {
		if score.ID == id {
			return score
		}
	}
	return IRBodyFillCandidateScore{}
}

func bestBodyFillCandidate(scores []IRBodyFillCandidateScore) IRBodyFillCandidateScore {
	if len(scores) == 0 {
		return IRBodyFillCandidateScore{}
	}
	best := scores[0]
	for _, score := range scores[1:] {
		if score.AccuracyPercent > best.AccuracyPercent {
			best = score
		}
	}
	return best
}

func candidateByID(candidates []IRBodyFillCandidate, id string) (IRBodyFillCandidate, bool) {
	for _, candidate := range candidates {
		if candidate.ID == id {
			return candidate, true
		}
	}
	return IRBodyFillCandidate{}, false
}
