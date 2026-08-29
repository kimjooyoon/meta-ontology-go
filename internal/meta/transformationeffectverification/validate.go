package transformationeffectverification

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const verifierSchema = "gooo/transformation-executor-binding-verifier/v1"

func Verify(opts Options) (Report, error) {
	if opts.RuntimePath == "" {
		err := unknownFailure("read-runtime", "RUNTIME_OBSERVATION_MISSING", "restore-runtime-evidence")
		return failureReport(err), err
	}
	bundle, err := loadBundle(opts)
	if err != nil {
		return failureReport(err), err
	}
	if err := validateBundle(bundle, opts.ExpectedHead); err != nil {
		return failureReport(err), err
	}
	return successReport(bundle), nil
}

func validateBundle(bundle bundle, expectedHead string) error {
	if !validHead(expectedHead) {
		return bindingFailure("plan.head_sha", "valid expected head", expectedHead)
	}
	if bundle.Plan.HeadSHA != expectedHead || bundle.Execution.HeadSHA != expectedHead ||
		bundle.Receipts.HeadSHA != expectedHead || bundle.Provenance.HeadSHA != expectedHead ||
		bundle.Ledger.HeadSHA != expectedHead || bundle.Patch.HeadSHA != expectedHead {
		return bindingFailure("bundle.head_sha", expectedHead, "mismatched artifact head")
	}
	if err := validatePlan(bundle.Plan); err != nil {
		return err
	}
	if err := validateExecution(bundle.Plan, bundle.Execution); err != nil {
		return err
	}
	if err := validateLedger(bundle.Plan, bundle.Ledger); err != nil {
		return err
	}
	if err := validateReceipts(bundle.Plan, bundle.Receipts); err != nil {
		return err
	}
	if err := validateProvenance(bundle.Plan, bundle.Receipts, bundle.Provenance); err != nil {
		return err
	}
	if err := validatePatch(bundle.Plan, bundle.Patch); err != nil {
		return err
	}
	if err := validateRuntime(bundle.Runtime); err != nil {
		return err
	}
	return validateCrossDigests(bundle.Ledger, bundle.Receipts, bundle.Provenance, bundle.Patch)
}

func validatePlan(plan generation.Plan) error {
	if plan.SchemaVersion != generation.SchemaVersion || plan.Decision != generation.DecisionPlan ||
		len(plan.Selected) != 2 || len(plan.Registry) == 0 {
		return bindingFailure("plan", "PLAN with two selected operations", "invalid plan shape")
	}
	seenOperations := map[sourcepolicy.Operation]bool{}
	seenIndicators := map[string]bool{}
	for _, action := range plan.Selected {
		if seenOperations[action.Operation] || seenIndicators[action.IndicatorID] {
			return bindingFailure("plan.selected", "unique operation and indicator IDs", string(action.Operation))
		}
		seenOperations[action.Operation] = true
		seenIndicators[action.IndicatorID] = true
		if err := validateActionBinding(plan.Registry, action); err != nil {
			return err
		}
	}
	if !seenOperations[sourcepolicy.OperationSplitGo] || !seenOperations[sourcepolicy.OperationExtractFunction] {
		return bindingFailure("plan.selected", "split-go-declarations and extract-function", "different operation set")
	}
	return nil
}

func validateActionBinding(registry []generation.Binding, action generation.Action) error {
	candidates := make([]generation.Binding, 0, 1)
	for _, binding := range registry {
		if binding.Operation == action.Operation {
			candidates = append(candidates, binding)
		}
	}
	if len(candidates) != 1 {
		return bindingFailure(bindingPath(action.Operation), "exactly one registry binding", fmt.Sprintf("%d", len(candidates)))
	}
	expected, ok := generation.BindingForOperation(generation.DefaultRegistry(), action.Operation)
	if !ok {
		return bindingFailure(bindingPath(action.Operation), "registered operation", string(action.Operation))
	}
	if err := compareBinding(action, candidates[0]); err != nil {
		return err
	}
	return compareBinding(action, expected)
}

func compareBinding(action generation.Action, binding generation.Binding) error {
	if action.Operation != binding.Operation {
		return bindingFailure(bindingPath(action.Operation)+".operation", string(binding.Operation), string(action.Operation))
	}
	checks := []struct {
		path, expected, observed string
	}{
		{"activity", binding.Activity, action.Activity}, {"output", binding.Output, action.Output},
		{"executor", binding.Executor, action.Executor}, {"evaluator", binding.Evaluator, action.Evaluator},
		{"independence_group_id", binding.IndependenceGroupID, action.IndependenceGroupID},
		{"proof_choice", string(binding.ProofChoice), string(action.ProofChoice)},
	}
	for _, check := range checks {
		if check.expected != check.observed {
			return bindingFailure(bindingPath(action.Operation)+"."+check.path, check.expected, check.observed)
		}
	}
	if !reflect.DeepEqual(sorted(binding.RequiredIndicatorIDs), sorted(action.RequiredIndicatorIDs)) ||
		!action.ReceiptRequired {
		return bindingFailure(bindingPath(action.Operation)+".required_indicator_ids", strings.Join(sorted(binding.RequiredIndicatorIDs), ","), strings.Join(sorted(action.RequiredIndicatorIDs), ","))
	}
	return nil
}

func validateExecution(plan generation.Plan, execution generation.ExecutionManifest) error {
	if execution.SchemaVersion != generation.ExecutionManifestSchemaVersion ||
		execution.PlanDigest != plan.PlanDigest || execution.HeadSHA != plan.HeadSHA ||
		len(execution.Steps) != len(plan.Selected) || execution.Decision != generation.ExecutionDecisionProposed {
		return bindingFailure("execution", "bound proposed manifest", "execution contract mismatch")
	}
	for _, action := range plan.Selected {
		found := 0
		for _, step := range execution.Steps {
			if step.ActionIndicatorID == action.IndicatorID {
				found++
				if step.Operation != action.Operation || step.Executor != action.Executor ||
					step.Evaluator != action.Evaluator || step.Activity != action.Activity ||
					step.Output != action.Output || step.ProofChoice != action.ProofChoice {
					return bindingFailure("execution.steps["+action.IndicatorID+"]", "selected action binding", "stale execution binding")
				}
			}
		}
		if found != 1 {
			return bindingFailure("execution.steps["+action.IndicatorID+"]", "one step", fmt.Sprintf("%d", found))
		}
	}
	return nil
}

func validateLedger(plan generation.Plan, ledger ledger) error {
	if ledger.Schema != "gooo/transformation-effect-ledger/v1" || ledger.BaseSHA != plan.BaseSHA ||
		ledger.HeadSHA != plan.HeadSHA || ledger.SelectedPlanOperations != 2 ||
		ledger.BoundExecutorOperations != 2 || ledger.UnboundExecutorOperations != 0 || len(ledger.Effects) != 2 {
		return bindingFailure("ledger", "2 selected, 2 bound, 0 unbound", "ledger counters or context mismatch")
	}
	for _, action := range plan.Selected {
		if err := validateEffect(ledger.Effects, action); err != nil {
			return err
		}
	}
	return nil
}

func validateEffect(effects []effect, action generation.Action) error {
	found := 0
	for _, effect := range effects {
		if effect.ActionIndicatorID != action.IndicatorID {
			continue
		}
		found++
		wantStatus := "APPLIED"
		if action.Operation == sourcepolicy.OperationExtractFunction {
			wantStatus = "REFUTED"
		}
		if effect.MetricID != string(action.MetricID) || effect.Subject != action.Subject ||
			effect.Operation != string(action.Operation) || effect.Activity != action.Activity ||
			effect.Output != action.Output || effect.Executor != action.Executor ||
			effect.Evaluator != action.Evaluator || effect.ProofChoice != string(action.ProofChoice) ||
			effect.Status != wantStatus {
			return bindingFailure("ledger.effects["+action.IndicatorID+"]", wantStatus, effect.Status)
		}
	}
	if found != 1 {
		return bindingFailure("ledger.effects["+action.IndicatorID+"]", "one effect", fmt.Sprintf("%d", found))
	}
	return nil
}

func validateReceipts(plan generation.Plan, report generation.ReceiptReport) error {
	if report.SchemaVersion != generation.ReceiptReportSchemaVersion || report.PlanDigest != plan.PlanDigest ||
		report.Decision != generation.ReceiptDecisionRefuted || report.Reason != generation.ReceiptReasonRefutedOperation ||
		len(report.Receipts) != 1 || len(report.Failures) != 1 || report.PromotionAuthorized ||
		len(report.MissingIndicatorIDs) != 0 || len(report.RejectedIndicatorIDs) != 0 || len(report.Unknowns) != 5 {
		return bindingFailure("receipts", "one closed receipt and one refuted failure", "mixed receipt contract mismatch")
	}
	split, extract, ok := selectedActions(plan.Selected)
	if !ok {
		return bindingFailure("receipts", "selected operation set", "missing split or extract action")
	}
	if err := validateClosedReceipt(report.Receipts[0], split, plan.HeadSHA); err != nil {
		return err
	}
	return validateExtractFailure(report, extract)
}

func selectedActions(actions []generation.Action) (generation.Action, generation.Action, bool) {
	var split, extract generation.Action
	for _, action := range actions {
		switch action.Operation {
		case sourcepolicy.OperationSplitGo:
			split = action
		case sourcepolicy.OperationExtractFunction:
			extract = action
		}
	}
	return split, extract, split.IndicatorID != "" && extract.IndicatorID != ""
}

func validateClosedReceipt(receipt generation.OperationReceipt, action generation.Action, head string) error {
	if receipt.ActionIndicatorID != action.IndicatorID || receipt.Operation != action.Operation ||
		receipt.Activity != action.Activity || receipt.Output != action.Output || receipt.Executor != action.Executor ||
		receipt.Evaluator != action.Evaluator || receipt.ProofChoice != action.ProofChoice || receipt.InstanceEvidence == nil ||
		len(receipt.Indicators) != len(action.RequiredIndicatorIDs) {
		return bindingFailure("receipts.receipt", "closed selected action", "receipt binding mismatch")
	}
	if err := validateProcessEvidence(receipt.InstanceEvidence); err != nil {
		return err
	}
	allowed := map[string]bool{}
	for _, id := range action.RequiredIndicatorIDs {
		allowed[id] = true
	}
	for _, indicator := range receipt.Indicators {
		if !allowed[indicator.ID] || indicator.Verdict != generation.IndicatorVerdictPass || indicator.Observation == nil ||
			indicator.Observation.Subject != action.Subject || indicator.Observation.HeadSHA != head ||
			indicator.Observation.ActualValue != 1 || indicator.Observation.ExpectedBound != 1 {
			return bindingFailure("receipts.receipt.indicators", "typed PASS observations", indicator.ID)
		}
		delete(allowed, indicator.ID)
	}
	if len(allowed) != 0 {
		return bindingFailure("receipts.receipt.indicators", "all required indicators", "indicator missing")
	}
	return nil
}

func validateExtractFailure(report generation.ReceiptReport, action generation.Action) error {
	failure := report.Failures[0]
	if failure.ActionIndicatorID != action.IndicatorID || failure.Decision != "REFUTED" || failure.Stage != "derive-recipe" ||
		failure.Step != "select-declaration" || failure.Reason != "NO_SAFE_DECLARATION_CAPACITY" ||
		failure.NextOperation != "report-counterexample" || failure.BlockedBy == nil || len(failure.BlockedBy) != 0 ||
		len(failure.FailureEvidence) != len(action.RequiredIndicatorIDs) || failure.Executor.Command == nil {
		return bindingFailure("receipts.failures[0]", "typed NO_SAFE_DECLARATION_CAPACITY", failure.Reason)
	}
	if len(report.UnknownIndicatorIDs) != 1 || report.UnknownIndicatorIDs[0] != action.IndicatorID+"::dependency:"+failure.Reason {
		return bindingFailure("receipts.unknown_indicator_ids", "one dependency obligation", "unknown linkage mismatch")
	}
	if err := validateFailureEvidence(failure, action); err != nil {
		return err
	}
	return validateDependencyUnknowns(report.Unknowns, failure, action)
}

func validateFailureEvidence(failure generation.ObservationFailure, action generation.Action) error {
	allowed := map[string]bool{}
	for _, id := range action.RequiredIndicatorIDs {
		allowed[id] = true
	}
	for _, evidence := range failure.FailureEvidence {
		if !allowed[evidence.IndicatorID] || evidence.Decision != "UNKNOWN" || evidence.Observed != 0 || evidence.Expected != 1 {
			return bindingFailure("receipts.failures[0].failure_evidence", "dependency evidence", evidence.IndicatorID)
		}
		delete(allowed, evidence.IndicatorID)
	}
	if len(allowed) != 0 {
		return bindingFailure("receipts.failures[0].failure_evidence", "all required indicators", "indicator missing")
	}
	return nil
}

func validateDependencyUnknowns(unknowns []generation.ReceiptUnknown, failure generation.ObservationFailure, action generation.Action) error {
	allowed := map[string]bool{}
	for _, id := range action.RequiredIndicatorIDs {
		allowed[id] = true
	}
	root := "operation-failure:" + action.IndicatorID
	for _, unknown := range unknowns {
		if unknown.ActionIndicatorID != action.IndicatorID || !allowed[unknown.RequiredIndicatorID] ||
			unknown.Stage != failure.Stage || unknown.Step != failure.Step || unknown.Reason != generation.ReceiptReason(failure.Reason) ||
			unknown.UnknownClass != generation.ReceiptUnknownClassDependencyBlocked || unknown.NextOperation != failure.NextOperation ||
			unknown.BlockedBy == nil || len(unknown.BlockedBy) != 1 || unknown.BlockedBy[0] != root {
			return bindingFailure("receipts.unknowns", root, "dependency frontier mismatch")
		}
		delete(allowed, unknown.RequiredIndicatorID)
	}
	if len(allowed) != 0 {
		return bindingFailure("receipts.unknowns", "five dependency unknowns", "unknown missing")
	}
	return nil
}

func validateProcessEvidence(evidence *generation.OperationInstanceEvidence) error {
	if evidence.Schema != generation.OperationInstanceEvidenceSchema || evidence.ActionIndicatorID == "" ||
		evidence.Subject == "" || evidence.HeadSHA == "" || evidence.OperationID == "" || !evidence.ReplayMatch ||
		evidence.ReplayComparisons < 1 || evidence.ExecutorObservation.Command == nil ||
		evidence.EvaluatorObservation.Command == nil || evidence.VerifierObservation == nil ||
		evidence.VerifierObservation.Command == nil {
		return bindingFailure("receipts.instance_evidence", "complete process observations", "instance evidence missing")
	}
	return nil
}

func validateProvenance(plan generation.Plan, receipts generation.ReceiptReport, provenance provenance) error {
	if provenance.Schema != generation.ArtifactProvenanceSchemaVersion || provenance.HeadSHA != plan.HeadSHA ||
		provenance.ReceiptReportDigest != receipts.ReportDigest || provenance.EnvelopeDigest == "" {
		return bindingFailure("provenance", "bound receipt report", "provenance mismatch")
	}
	return nil
}

func validatePatch(plan generation.Plan, output patch) error {
	if output.Schema != "gooo/transformation-content-patch/v1" || output.HeadSHA != plan.HeadSHA || output.PatchDigest == "" {
		return bindingFailure("patch", "bound patch", "patch mismatch")
	}
	return nil
}

func validateCrossDigests(ledger ledger, receipts generation.ReceiptReport, provenance provenance, output patch) error {
	if ledger.GeneratedReceiptReportDigest != receipts.ReportDigest || ledger.ExecutedProvenanceDigest != provenance.EnvelopeDigest ||
		ledger.PatchDigest != output.PatchDigest {
		return bindingFailure("digest_bindings", "receipt/provenance/patch digests", "cross-artifact digest mismatch")
	}
	return nil
}

func validateRuntime(runtime Runtime) error {
	if runtime.First.WallMS <= 0 || runtime.First.PeakRSSKiB <= 0 || runtime.Replay.WallMS <= 0 ||
		runtime.Replay.PeakRSSKiB <= 0 || !runtime.ReplayEqual || runtime.RepositoryWrites != 0 ||
		runtime.LocalTestExecutions != 0 || runtime.CrossProjectRequiredGates != 0 {
		return bindingFailure("runtime", "positive replayed observations and zero authorities", "runtime contract mismatch")
	}
	return nil
}

func successReport(bundle bundle) Report {
	commands, tests := processCounts(bundle.Receipts)
	return Report{Schema: verifierSchema, Decision: "PASS", Resolution: "EXACT", Stage: "verify-bundle",
		Step: "adjudicate-mixed-operation", Reason: "TRANSFORMATION_EXECUTOR_BINDINGS_VERIFIED", NextOperation: "none",
		BlockedBy: []string{}, SelectedPlanOperations: 2, BoundExecutorOperations: 2, ReceiptCount: 1,
		FailureCount: 1, PhysicalCommands: commands, PhysicalTests: tests, ReusedCommands: 0, ReusedTests: 0,
		RepositoryWrites: bundle.Runtime.RepositoryWrites, LocalTestExecutions: bundle.Runtime.LocalTestExecutions,
		CrossProjectRequiredGates: bundle.Runtime.CrossProjectRequiredGates, Improvement: "UNKNOWN",
		OperationOutcome: "MIXED_CLOSED_REFUTED", PromotionAuthorized: false, Runtime: bundle.Runtime}
}

func processCounts(report generation.ReceiptReport) (int, int) {
	commands := 0
	tests := 0
	for _, receipt := range report.Receipts {
		if receipt.InstanceEvidence == nil {
			continue
		}
		commands++
		if hasTestCommand(receipt.InstanceEvidence.VerifierObservation.Command) {
			tests++
		}
	}
	for _, failure := range report.Failures {
		if len(failure.Executor.Command) != 0 {
			commands++
		}
	}
	return commands, tests
}

func hasTestCommand(command []string) bool {
	for _, argument := range command {
		if argument == "test" || strings.HasSuffix(argument, "/test") {
			return true
		}
	}
	return false
}

func sorted(values []string) []string {
	copyOf := append([]string{}, values...)
	sort.Strings(copyOf)
	return copyOf
}

func validHead(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func bindingPath(operation sourcepolicy.Operation) string {
	return "registry[" + string(operation) + "]"
}

func bindingFailure(path, expected, observed string) *validationFailure {
	return &validationFailure{Decision: "REFUTED", Resolution: "EXACT", Stage: "verify-bundle",
		Step: "compare-bindings", Reason: "META_OPERATION_EXECUTOR_BINDING_MISMATCH", Next: "report-counterexample",
		Blocked: []string{}, FieldPath: path, Expected: expected, Observed: observed}
}
