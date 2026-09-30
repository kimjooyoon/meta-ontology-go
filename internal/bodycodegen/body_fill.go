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
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const bodyFillPlanSchema = "gooo/body-codegen-ir-fill-plan/v1"
const bodyFillStateSchema = "gooo/body-codegen-ir-fill-state/v1"
const bodyFillEvaluator = "gooo/bodycodegen-int64-ast-interpreter/v2"
const irBodyFillDecisionBudget = 8 * time.Second

// IRBodyFillPlan supplies finite, typed expression candidates for one explicit
// hole in a Gooo activity body. Gooo owns the body skeleton and emitter; an
// optional decision provider may choose only one of these expressions.
type IRBodyFillPlan struct {
	Schema        string                `json:"schema"`
	Intent        string                `json:"intent"`
	HoleID        string                `json:"hole_id"`
	ProviderModel string                `json:"provider_model,omitempty"`
	Candidates    []IRBodyFillCandidate `json:"candidates"`
	TestCases     []IRBodyFillTestCase  `json:"test_cases"`
}

type IRBodyFillCandidate struct {
	ID         string `json:"id"`
	Expression string `json:"expression"`
}

// IRBodyFillTestCase is a finite, explicit integer contract used to measure
// the selected body's observed functional accuracy. It is not a proof over all
// int64 values.
type IRBodyFillTestCase struct {
	Input    int64 `json:"input"`
	Expected int64 `json:"expected"`
}

func (testCase *IRBodyFillTestCase) UnmarshalJSON(data []byte) error {
	var fields struct {
		Input    *int64 `json:"input"`
		Expected *int64 `json:"expected"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fields); err != nil {
		return err
	}
	if fields.Input == nil || fields.Expected == nil {
		return fmt.Errorf("IR body-fill test case requires explicit integer input and expected values")
	}
	*testCase = IRBodyFillTestCase{Input: *fields.Input, Expected: *fields.Expected}
	return nil
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
	Input    int64 `json:"input"`
	Expected int64 `json:"expected"`
	Actual   int64 `json:"actual"`
	Passed   bool  `json:"passed"`
}

type IRBodyFillTiming struct {
	IRPlanBuildMS      float64  `json:"ir_plan_build_ms"`
	TinyModelLoadMS    *float64 `json:"tiny_model_load_ms,omitempty"`
	ProviderDecisionMS float64  `json:"provider_decision_ms"`
	LayaDecisionMS     float64  `json:"laya_decision_ms"`
	TinyDecisionMS     float64  `json:"tiny_decision_ms"`
	FinalEmissionMS    float64  `json:"final_emission_ms"`
	TotalMS            float64  `json:"total_ms"`
	ExecutionModel     string   `json:"execution_model"`
	DecisionStage      string   `json:"decision_stage"`
}

type IRBodyFillReceipt struct {
	Schema                     string                     `json:"schema"`
	Intent                     string                     `json:"intent"`
	HoleID                     string                     `json:"hole_id"`
	HoleToken                  string                     `json:"hole_token"`
	IRPlanSHA256               string                     `json:"ir_plan_sha256"`
	ProposedCandidateID        string                     `json:"proposed_candidate_id"`
	ProposedAccuracyPct        float64                    `json:"proposed_accuracy_percent"`
	SelectedCandidateID        string                     `json:"selected_candidate_id"`
	SelectedExpression         string                     `json:"selected_expression"`
	BestCandidateID            string                     `json:"best_candidate_id"`
	BestAccuracyPercent        float64                    `json:"best_candidate_accuracy_percent"`
	SelectionRegretPP          float64                    `json:"selection_regret_percentage_points"`
	SelectionAdjustment        string                     `json:"selection_adjustment"`
	Decision                   decisionroute.Receipt      `json:"decision"`
	CandidateScores            []IRBodyFillCandidateScore `json:"candidate_scores"`
	TestSuiteSHA256            string                     `json:"test_suite_sha256"`
	TestCasesPassed            int                        `json:"test_cases_passed"`
	TestCasesTotal             int                        `json:"test_cases_total"`
	FunctionalAccuracyPct      float64                    `json:"functional_accuracy_percent"`
	LocalModelPredictions      int                        `json:"local_model_predictions"`
	ExternalProviderCalls      int                        `json:"external_provider_calls"`
	ExternalProviderCallsKnown bool                       `json:"external_provider_calls_known"`
	Evaluator                  string                     `json:"evaluator"`
	SelectedCaseResults        []IRBodyFillCaseResult     `json:"selected_case_results"`
	AccuracyScope              string                     `json:"accuracy_scope"`
	Timing                     IRBodyFillTiming           `json:"timing"`
}

type irBodyFillState struct {
	Schema          string                     `json:"schema"`
	Stage           string                     `json:"stage"`
	Activity        string                     `json:"activity"`
	ActivityID      string                     `json:"activity_id"`
	InputType       string                     `json:"input_type"`
	OutputType      string                     `json:"output_type"`
	Intent          string                     `json:"intent"`
	HoleID          string                     `json:"hole_id"`
	BodyIR          string                     `json:"body_ir"`
	TestCaseCount   int                        `json:"test_case_count"`
	TestSuiteSHA256 string                     `json:"test_suite_sha256"`
	Candidates      []IRBodyFillCandidateScore `json:"candidate_scores"`
}

// GenerateWithIRBodyFill builds the typed hole plan synchronously from a Gooo
// activity, calls the existing Laya/default path once after that plan is
// complete, fills the hole from the declared candidate set, and only then emits
// the final Go projection. The v1 experiment is limited to one Integer ->
// Integer hole.
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
	if len(activity.Inputs) != 1 || activity.Inputs[0].Name != "Integer" || activity.Output != "Integer" {
		return Result{}, fmt.Errorf("IR body fill v1 requires one Integer input and one Integer output")
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
	holeToken := bodyFillHoleToken(plan.HoleID)
	if countIdentifier(body, holeToken) != 1 {
		return Result{}, fmt.Errorf("activity %q must contain exactly one IR body hole %q", activityName, holeToken)
	}
	planStarted := time.Now()
	scores := make([]IRBodyFillCandidateScore, 0, len(plan.Candidates))
	candidateBodies := make(map[string]string, len(plan.Candidates))
	for _, candidate := range plan.Candidates {
		candidateBody, err := replaceIdentifier(body, holeToken, candidate.Expression)
		if err != nil {
			return Result{}, fmt.Errorf("fill candidate %q: %w", candidate.ID, err)
		}
		generated, err := generateRoute(
			file.Package.Name, activityName, activityID, "int64", "int64", candidateBody, preserveRoute,
		)
		if err != nil {
			return Result{}, fmt.Errorf("candidate %q is not a valid typed body: %w", candidate.ID, err)
		}
		_, passed, err := evaluateIntegerCases(generated.source, activityName, plan.TestCases)
		if err != nil {
			return Result{}, fmt.Errorf("evaluate candidate %q: %w", candidate.ID, err)
		}
		candidateBodies[candidate.ID] = candidateBody
		accuracy := float64(passed) * 100 / float64(len(plan.TestCases))
		scores = append(scores, IRBodyFillCandidateScore{
			ID: candidate.ID, Expression: candidate.Expression, TypecheckPassed: true,
			TestCasesPassed: passed, TestCasesTotal: len(plan.TestCases), AccuracyPercent: accuracy,
		})
	}
	planBuildMS := float64(time.Since(planStarted)) / float64(time.Millisecond)
	testBytes, _ := json.Marshal(plan.TestCases)
	testSuiteSHA256 := digest(testBytes)
	stateBytes, err := json.Marshal(irBodyFillState{
		Schema: bodyFillStateSchema, Stage: "ir_ready_before_body_emission",
		Activity: activityName, ActivityID: activityID, InputType: "Integer", OutputType: "Integer",
		Intent: plan.Intent, HoleID: plan.HoleID, BodyIR: body,
		TestCaseCount: len(plan.TestCases), TestSuiteSHA256: testSuiteSHA256, Candidates: scores,
	})
	if err != nil {
		return Result{}, fmt.Errorf("encode Gooo IR body-fill state: %w", err)
	}
	requestOptions := make([]decisionroute.Option, 0, len(plan.Candidates))
	if usingTinyGo {
		requestOptions, err = tinyGoBodyFillOptions(plan.Candidates)
		if err != nil {
			return Result{}, err
		}
	} else {
		for _, candidate := range plan.Candidates {
			score := scoreByID(scores, candidate.ID)
			requestOptions = append(requestOptions, decisionroute.Option{
				ID: candidate.ID,
				Description: fmt.Sprintf("Emit exactly this expression: %s. It passes %d of %d declared test cases (%.2f%%).",
					candidate.Expression, score.TestCasesPassed, score.TestCasesTotal, score.AccuracyPercent),
			})
		}
	}
	request := decisionroute.Request{
		Schema: decisionroute.RequestSchema, State: string(stateBytes),
		ProviderModel: plan.ProviderModel,
		Question: decisionroute.Question{
			ID: "body_ir_fill",
			Instructions: "Fill the single typed expression hole in the supplied Gooo body IR. " +
				"Choose only a listed candidate. " +
				"Use the intent and declared test evidence; do not invent code or modify any other IR node.",
			Options: requestOptions,
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
		if !validTinyGoDecisionReceipt(decision, expectedRequestSHA256) {
			return Result{}, fmt.Errorf("tiny_go chooser returned incomplete or mismatched model provenance")
		}
	}
	if usingTinyGo {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
	}
	proposed, ok := candidateByID(plan.Candidates, decision.Selected)
	if !ok {
		return Result{}, fmt.Errorf("body-fill decision selected undeclared candidate %q", decision.Selected)
	}
	best := bestBodyFillCandidate(scores)
	proposedScore := scoreByID(scores, proposed.ID)
	selected := proposed
	selectionAdjustment := "proposal_retained"
	if proposedScore.TestCasesPassed < best.TestCasesPassed {
		selected, _ = candidateByID(plan.Candidates, best.ID)
		selectionAdjustment = "replaced_with_best_scoring_candidate"
	}
	selectedBody := candidateBodies[selected.ID]
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
	caseResults, passed, err := evaluateIntegerCases([]byte(result.Source), activityName, plan.TestCases)
	if err != nil {
		return Result{}, fmt.Errorf("evaluate selected generated body: %w", err)
	}
	accuracy := float64(passed) * 100 / float64(len(plan.TestCases))
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
	result.Report.BodyFill = &IRBodyFillReceipt{
		Schema: bodyFillPlanSchema, Intent: plan.Intent, HoleID: plan.HoleID,
		HoleToken: holeToken, IRPlanSHA256: digest(planBytes),
		ProposedCandidateID: proposed.ID, ProposedAccuracyPct: proposedScore.AccuracyPercent,
		SelectedCandidateID: selected.ID, SelectedExpression: selected.Expression,
		BestCandidateID: best.ID, BestAccuracyPercent: best.AccuracyPercent,
		SelectionRegretPP:   best.AccuracyPercent - proposedScore.AccuracyPercent,
		SelectionAdjustment: selectionAdjustment,
		Decision:            decision, CandidateScores: scores,
		TestSuiteSHA256: testSuiteSHA256, TestCasesPassed: passed,
		TestCasesTotal: len(plan.TestCases), FunctionalAccuracyPct: accuracy,
		LocalModelPredictions: localPredictions, ExternalProviderCalls: externalCalls,
		ExternalProviderCallsKnown: externalCallsKnown,
		Evaluator:                  bodyFillEvaluator,
		SelectedCaseResults:        caseResults,
		AccuracyScope:              "exact observed accuracy over the declared finite test suite; not a full-domain proof",
		Timing: IRBodyFillTiming{
			IRPlanBuildMS: planBuildMS, TinyModelLoadMS: tinyModelLoadMS,
			ProviderDecisionMS: decisionMS,
			LayaDecisionMS:     layaDecisionMS, TinyDecisionMS: tinyDecisionMS,
			FinalEmissionMS: emissionMS,
			TotalMS:         float64(time.Since(totalStarted)) / float64(time.Millisecond),
			ExecutionModel:  "synchronous_sequential_no_background_codegen_goroutines",
			DecisionStage:   "after_typed_ir_plan_and_candidate_test_scores_before_final_emission",
		},
	}
	populateCompletenessReceipt(&result.Report, "")
	return result, nil
}

func validateIRBodyFillPlan(plan IRBodyFillPlan) error {
	if plan.Schema != bodyFillPlanSchema {
		return fmt.Errorf("IR body-fill plan schema must be %q", bodyFillPlanSchema)
	}
	if strings.TrimSpace(plan.Intent) == "" || utf8.RuneCountInString(plan.Intent) > 2000 {
		return fmt.Errorf("IR body-fill intent must contain 1..2000 characters")
	}
	if !validBodyFillIdentifier(plan.HoleID) {
		return fmt.Errorf("IR body-fill hole id %q is invalid", plan.HoleID)
	}
	if len(plan.Candidates) < 2 || len(plan.Candidates) > 16 {
		return fmt.Errorf("IR body-fill plan requires 2..16 candidates")
	}
	if len(plan.TestCases) == 0 || len(plan.TestCases) > 4096 {
		return fmt.Errorf("IR body-fill plan requires 1..4096 integer test cases")
	}
	seen := make(map[string]bool, len(plan.Candidates))
	for _, candidate := range plan.Candidates {
		if !validBodyFillIdentifier(candidate.ID) || seen[candidate.ID] {
			return fmt.Errorf("IR body-fill candidate id %q is invalid or duplicated", candidate.ID)
		}
		seen[candidate.ID] = true
		if strings.TrimSpace(candidate.Expression) == "" || utf8.RuneCountInString(candidate.Expression) > 512 {
			return fmt.Errorf("IR body-fill candidate %q expression must contain 1..512 characters", candidate.ID)
		}
		expression, err := parser.ParseExpr(candidate.Expression)
		if err != nil {
			return fmt.Errorf("parse IR body-fill candidate %q: %w", candidate.ID, err)
		}
		if err := validateExpression(expression); err != nil {
			return fmt.Errorf("IR body-fill candidate %q uses an unsupported expression: %w", candidate.ID, err)
		}
	}
	return nil
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
