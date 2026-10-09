package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/parser"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func validateRecordIRBodyFillPlan(plan IRBodyFillPlan) error {
	if strings.TrimSpace(plan.Intent) == "" || utf8.RuneCountInString(plan.Intent) > 2000 ||
		len(plan.ValueCases) == 0 || len(plan.ValueCases)+len(plan.ValueHoldoutCases) > 128 || len(plan.TestCases) != 0 ||
		len(plan.HoldoutTestCases) != 0 || plan.HoleID != "" || len(plan.Holes) < 1 || len(plan.Holes) > 8 ||
		len(plan.Candidates) < 2 || len(plan.Candidates) > 16 {
		return fmt.Errorf("record IR body-fill plan requires intent, 1..8 holes, 2..16 candidates and 1..128 value cases")
	}
	seenHoles := map[string]bool{}
	for _, hole := range plan.Holes {
		if !validBodyFillIdentifier(hole.ID) || seenHoles[hole.ID] {
			return fmt.Errorf("record IR body-fill hole id %q is invalid or duplicated", hole.ID)
		}
		seenHoles[hole.ID] = true
	}
	seenCandidates := map[string]bool{}
	for _, candidate := range plan.Candidates {
		if !validBodyFillIdentifier(candidate.ID) || seenCandidates[candidate.ID] || candidate.Expression != "" || len(candidate.Fills) != len(plan.Holes) {
			return fmt.Errorf("record IR body-fill candidate %q must have a unique id and fill every hole exactly once", candidate.ID)
		}
		seenCandidates[candidate.ID] = true
		for _, hole := range plan.Holes {
			expression, ok := candidate.Fills[hole.ID]
			if !ok || strings.TrimSpace(expression) == "" || utf8.RuneCountInString(expression) > 512 {
				return fmt.Errorf("record IR body-fill candidate %q requires a bounded expression for hole %q", candidate.ID, hole.ID)
			}
			for _, other := range plan.Holes {
				if countIdentifier(expression, bodyFillHoleToken(other.ID)) != 0 {
					return fmt.Errorf("record IR body-fill candidate %q cannot introduce hole %q", candidate.ID, other.ID)
				}
			}
			parsed, err := parser.ParseExpr(expression)
			if err != nil {
				return fmt.Errorf("parse record IR body-fill candidate %q hole %q: %w", candidate.ID, hole.ID, err)
			}
			if err := validateExpression(parsed); err != nil {
				return fmt.Errorf("record IR body-fill candidate %q hole %q uses an unsupported expression: %w", candidate.ID, hole.ID, err)
			}
		}
	}
	for _, testCase := range plan.ValueCases {
		inputs, err := assemblyspec.CanonicalValue(testCase.Inputs)
		if err != nil || inputs != testCase.Inputs {
			return fmt.Errorf("record IR body-fill inputs must be bounded canonical JSON")
		}
		expected, err := assemblyspec.CanonicalValue(testCase.Expected)
		if err != nil || expected != testCase.Expected {
			return fmt.Errorf("record IR body-fill expected values must be bounded canonical JSON")
		}
	}
	trainingInputs := make(map[string]bool, len(plan.ValueCases))
	for _, testCase := range plan.ValueCases {
		trainingInputs[testCase.Inputs] = true
	}
	for _, testCase := range plan.ValueHoldoutCases {
		inputs, err := assemblyspec.CanonicalValue(testCase.Inputs)
		if err != nil || inputs != testCase.Inputs {
			return fmt.Errorf("record IR body-fill holdout inputs must be bounded canonical JSON")
		}
		expected, err := assemblyspec.CanonicalValue(testCase.Expected)
		if err != nil || expected != testCase.Expected {
			return fmt.Errorf("record IR body-fill holdout expected values must be bounded canonical JSON")
		}
		if trainingInputs[testCase.Inputs] {
			return fmt.Errorf("record IR body-fill holdout input %s also appears in training cases", testCase.Inputs)
		}
	}
	return nil
}

func generateWithRecordIRBodyFillOptions(
	ctx context.Context,
	filename string,
	source []byte,
	activityName string,
	plan IRBodyFillPlan,
	endpoint, apiKey string,
	options IRBodyFillOptions,
	tinyProvider tinyGoBodyFillResolver,
) (Result, error) {
	if ctx == nil {
		return Result{}, fmt.Errorf("record body-fill context is required")
	}
	usingTinyGo := bodyFillUsesTinyGo(options, tinyProvider)
	if usingTinyGo && (endpoint != "" || apiKey != "" || plan.ProviderModel != "") {
		return Result{}, fmt.Errorf("tiny_go record body fill cannot be combined with Laya endpoint, API key, or provider model")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	file, diagnostics := ParseBodyFile(filename, source)
	if file == nil || diagnostics.HasErrors() {
		return Result{}, fmt.Errorf("parse .gooo record body-fill source: %w", diagnostics.Error())
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == activityName {
			activity = candidate
			break
		}
	}
	if activity == nil || !activity.ValueProgramPresent || activity.ValueProgram == "" {
		return Result{}, fmt.Errorf("record body-fill activity %q has no computes body", activityName)
	}
	if len(activity.Inputs) < 1 || len(activity.Inputs) > 16 {
		return Result{}, fmt.Errorf("record body fill requires 1..16 typed inputs")
	}
	model, records, err := resolveBodyModel(file)
	if err != nil {
		return Result{}, err
	}
	outputRecord := recordTypeByName(records, activity.Output)
	if outputRecord == nil {
		return Result{}, fmt.Errorf("record IR body fill currently requires a declared record output")
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
	holes := plan.Holes
	for _, hole := range holes {
		if countIdentifier(body, bodyFillHoleToken(hole.ID)) != 1 {
			return Result{}, fmt.Errorf("activity %q must contain exactly one IR body hole %q", activityName, hole.ID)
		}
	}
	parameters := make([]InputParameter, len(activity.Inputs))
	for index, input := range activity.Inputs {
		name := "input"
		if len(activity.Inputs) > 1 {
			name = fmt.Sprintf("input%d", index)
		}
		typeName, ok := bodyEntityType(input.Name, records, file)
		if !ok {
			return Result{}, fmt.Errorf("activity input %q is outside the scalar or declared-record profile", input.Name)
		}
		parameters[index] = InputParameter{Name: name, Type: typeName}
	}

	started := time.Now()
	scores := make([]IRBodyFillCandidateScore, 0, len(plan.Candidates))
	var rejected []IRBodyFillCandidateRejection
	eligible := make([]IRBodyFillCandidate, 0, len(plan.Candidates))
	candidateBodies := make(map[string]string, len(plan.Candidates))
	for _, candidate := range plan.Candidates {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		candidateBody := body
		for _, hole := range holes {
			candidateBody, err = replaceIdentifier(candidateBody, bodyFillHoleToken(hole.ID), candidate.Fills[hole.ID])
			if err != nil {
				return Result{}, fmt.Errorf("fill candidate %q at hole %q: %w", candidate.ID, hole.ID, err)
			}
		}
		generated, err := generateRouteParameters(file.Package.Name, activityName, activityID, parameters,
			activity.Output, candidateBody, preserveRoute, records...)
		if err != nil {
			rejected = append(rejected, rejectedBodyFillCandidate(plan, candidate, "TYPECHECK", err))
			continue
		}
		observed, err := evaluateRecordAssembly(ctx, []byte(generated.source), activityName, records, plan.ValueCases)
		if err != nil {
			if ctx.Err() != nil {
				return Result{}, ctx.Err()
			}
			rejected = append(rejected, rejectedBodyFillCandidate(plan, candidate, "TRAINING_EVALUATION", err))
			continue
		}
		eligible = append(eligible, candidate)
		passed := 0
		for _, result := range observed {
			if result.Passed {
				passed++
			}
		}
		candidateBodies[candidate.ID] = candidateBody
		scores = append(scores, IRBodyFillCandidateScore{
			ID: candidate.ID, Expression: bodyFillCandidateDescription(holes, candidate.Fills), TypecheckPassed: true,
			TestCasesPassed: passed, TestCasesTotal: len(plan.ValueCases),
			AccuracyPercent: float64(passed) * 100 / float64(len(plan.ValueCases)),
		})
	}
	if len(eligible) == 0 {
		return Result{}, noValidBodyFill(ctx, activityName, plan, rejected)
	}
	usingTinyGo = usingTinyGo && len(eligible) > 1
	planBuildMS := float64(time.Since(started)) / float64(time.Millisecond)
	testBytes, _ := json.Marshal(plan.ValueCases)
	stateBytes, err := json.Marshal(irBodyFillState{
		Schema: bodyFillStateSchema, Stage: "ir_ready_before_body_emission", Activity: activityName,
		ActivityID: activityID, InputType: activity.Inputs[0].Name, InputTypes: bodyFillInputTypes(activity.Inputs),
		OutputType: activity.Output, Intent: plan.Intent, HoleID: bodyFillHoleSummary(holes),
		HoleIDs: bodyFillHoleIDs(plan, holes), BodyIR: body, TestCaseCount: len(plan.ValueCases),
		TestSuiteSHA256: digest(testBytes), Candidates: scores, RejectedCandidates: rejected,
	})
	if err != nil {
		return Result{}, fmt.Errorf("encode record Gooo IR body-fill state: %w", err)
	}
	modelCandidates := bodyFillModelCandidates(holes, eligible)
	tinyModelFocusHole := ""
	requestOptions := make([]decisionroute.Option, 0, len(modelCandidates))
	if usingTinyGo {
		tinyModelFocusHole, err = tinyGoBodyFillFocusHole(holes, eligible)
		if err != nil {
			return Result{}, err
		}
		requestOptions, err = tinyGoRecordBodyFillOptionsForHole(eligible, tinyModelFocusHole)
		if err != nil {
			return Result{}, err
		}
	} else {
		for _, candidate := range modelCandidates {
			score := scoreByID(scores, candidate.ID)
			requestOptions = append(requestOptions, decisionroute.Option{
				ID: candidate.ID, Description: fmt.Sprintf("Use this complete body assignment: %s. It matches %d of %d declared record cases (%.2f%%).",
					candidate.Expression, score.TestCasesPassed, score.TestCasesTotal, score.AccuracyPercent),
			})
		}
	}
	instructions := "Fill every typed expression hole in this Gooo body using one complete listed assignment. Choose only a listed assignment. Use the source intent and finite record case scores; do not invent code or edit another IR node."
	if usingTinyGo {
		instructions = "Classify the source intent into one offered operation label for the focused hole. Gooo keeps only complete assignments in that operation class, then selects the highest-scoring compatible assignment. Do not invent code or edit another IR node."
	}
	request := decisionroute.Request{
		Schema: decisionroute.RequestSchema, State: string(stateBytes), ProviderModel: plan.ProviderModel,
		Question: decisionroute.Question{
			ID: "body_ir_fill", Instructions: instructions, Options: requestOptions,
		}, Fallback: eligible[0].ID,
	}
	if usingTinyGo {
		request.Intent = plan.Intent
		request.Fallback = requestOptions[0].ID
	}
	decisionStarted := time.Now()
	decisionContext, cancel := context.WithTimeout(ctx, irBodyFillDecisionBudget)
	decision, err := resolveBodyFillDecision(decisionContext, request, endpoint, apiKey, options, tinyProvider)
	cancel()
	decisionMS := float64(time.Since(decisionStarted)) / float64(time.Millisecond)
	if err != nil {
		return Result{}, fmt.Errorf("select Gooo record body-fill candidate: %w", err)
	}
	if usingTinyGo {
		if decision.Provider != decisionroute.ProviderTinyGo {
			return Result{}, fmt.Errorf("tiny_go chooser returned provider %q", decision.Provider)
		}
		expectedRequestSHA256, validateErr := decisionroute.Validate(request)
		if validateErr != nil {
			return Result{}, fmt.Errorf("validate typed tiny_go record request: %w", validateErr)
		}
		if !validTinyGoDecisionReceipt(decision, expectedRequestSHA256, request) {
			return Result{}, fmt.Errorf("tiny_go chooser returned incomplete or mismatched model provenance")
		}
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
	}
	proposedID := decision.Selected
	if usingTinyGo {
		proposedID, err = tinyGoRecordCandidateForOperation(tinyModelFocusHole, decision.Selected, eligible, scores)
		if err != nil {
			return Result{}, err
		}
	}
	proposed, ok := candidateByID(modelCandidates, proposedID)
	if !ok {
		return Result{}, fmt.Errorf("record body-fill decision selected undeclared candidate %q", proposedID)
	}
	best := bestBodyFillCandidate(scores)
	proposedScore := scoreByID(scores, proposed.ID)
	selected := proposed
	selectionAdjustment := "proposal_retained"
	if proposedScore.TestCasesPassed < best.TestCasesPassed {
		selected, _ = candidateByID(modelCandidates, best.ID)
		selectionAdjustment = "replaced_with_best_scoring_candidate"
	}
	originalSelected, ok := candidateByID(plan.Candidates, selected.ID)
	if !ok {
		return Result{}, fmt.Errorf("selected record body-fill candidate %q is absent from the source plan", selected.ID)
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
		return Result{}, fmt.Errorf("emit selected Gooo record body: %w", err)
	}
	selectedCases, err := evaluateRecordAssembly(ctx, []byte(result.Source), activityName, records, plan.ValueCases)
	if err != nil {
		return Result{}, fmt.Errorf("evaluate selected generated record body: %w", err)
	}
	passed := 0
	for _, caseResult := range selectedCases {
		if caseResult.Passed {
			passed++
		}
	}
	accuracy := float64(passed) * 100 / float64(len(plan.ValueCases))
	var holdoutResults []RecordAssemblyCase
	var holdoutPassed int
	var holdoutAccuracy *float64
	var holdoutSuiteSHA256 string
	if len(plan.ValueHoldoutCases) > 0 {
		holdoutResults, err = evaluateRecordAssembly(ctx, []byte(result.Source), activityName, records, plan.ValueHoldoutCases)
		if err != nil {
			return Result{}, fmt.Errorf("evaluate selected generated record body on holdout cases: %w", err)
		}
		for _, caseResult := range holdoutResults {
			if caseResult.Passed {
				holdoutPassed++
			}
		}
		value := float64(holdoutPassed) * 100 / float64(len(plan.ValueHoldoutCases))
		holdoutAccuracy = &value
		holdoutBytes, _ := json.Marshal(plan.ValueHoldoutCases)
		holdoutSuiteSHA256 = digest(holdoutBytes)
	}
	selectedBodySource, err := sourceWithoutIRBodyFill(filename, []byte(result.GoooSource), activityName)
	if err != nil {
		// External plans do not carry a source-owned fill clause.
		selectedBodySource = completedSource
	}
	result.GoooSource = string(selectedBodySource)
	localPredictions, externalCalls, externalCallsKnown := bodyFillProviderAccounting(decision)
	layaMS, tinyMS := 0.0, 0.0
	var tinyModelLoadMS *float64
	switch decision.Provider {
	case decisionroute.ProviderTinyGo:
		tinyMS = decisionMS
		tinyModelLoadMS = options.TinyModelLoadMS
	case "laya":
		layaMS = decisionMS
	}
	planBytes, _ := json.Marshal(plan)
	result.Report.BodyFill = &IRBodyFillReceipt{
		OriginalSourceDigest: digest(source),
		Schema:               plan.Schema, Intent: plan.Intent, HoleID: bodyFillHoleSummary(holes),
		HoleToken: bodyFillHoleToken(holes[0].ID), IRPlanSHA256: digest(planBytes),
		ProposedCandidateID: proposed.ID, ProposedAccuracyPct: proposedScore.AccuracyPercent,
		SelectedCandidateID: selected.ID, SelectedExpression: selected.Expression,
		HoleFills:          bodyFillHoleResults(holes, bodyFillCandidateFills(plan, originalSelected)),
		TinyModelFocusHole: tinyModelFocusHole,
		BestCandidateID:    best.ID, BestAccuracyPercent: best.AccuracyPercent,
		SelectionRegretPP:   best.AccuracyPercent - proposedScore.AccuracyPercent,
		SelectionAdjustment: selectionAdjustment, Decision: decision, CandidateScores: scores, RejectedCandidates: rejected,
		TestSuiteSHA256: digest(testBytes), TestCasesPassed: passed, TestCasesTotal: len(plan.ValueCases),
		HoldoutSuiteSHA256: holdoutSuiteSHA256, HoldoutCasesPassed: holdoutPassed,
		HoldoutCasesTotal: len(plan.ValueHoldoutCases), HoldoutAccuracyPercent: holdoutAccuracy,
		FunctionalAccuracyPct: accuracy, ExternalProviderCalls: externalCalls,
		ExternalProviderCallsKnown: externalCallsKnown, LocalModelPredictions: localPredictions,
		Evaluator: "gooo/bodycodegen-record-value-interpreter/v1", SelectedValueCaseResults: selectedCases,
		SelectedValueHoldoutCaseResults: holdoutResults,
		AccuracyScope:                   "exact observed record equality over training value cases; holdout value cases are measured separately when present; neither is a full-domain proof",
		Timing: IRBodyFillTiming{IRPlanBuildMS: planBuildMS, TinyModelLoadMS: tinyModelLoadMS, ProviderDecisionMS: decisionMS,
			LayaDecisionMS: layaMS, TinyDecisionMS: tinyMS, FinalEmissionMS: emissionMS,
			TotalMS:        float64(time.Since(started)) / float64(time.Millisecond),
			ExecutionModel: "synchronous_sequential_no_background_codegen_goroutines",
			DecisionStage:  "after_typed_ir_plan_training_scores_before_final_emission"},
	}
	populateCompletenessReceipt(&result.Report, "")
	return result, nil
}
