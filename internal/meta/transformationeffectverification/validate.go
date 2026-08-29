package transformationeffectverification

import (
	"fmt"
	"path/filepath"
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
	if err := validateEffectOutcomes(bundle.Plan, bundle.Ledger, bundle.Receipts); err != nil {
		return err
	}
	if err := validateProvenance(bundle.Plan, bundle.Execution, bundle.Receipts, bundle.Provenance); err != nil {
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
		len(plan.Selected) == 0 || len(plan.Registry) == 0 {
		return bindingFailure("plan", "PLAN with selected operations", "invalid plan shape")
	}
	seenIndicators := map[string]bool{}
	for _, action := range plan.Selected {
		if action.IndicatorID == "" || action.Operation == "" || seenIndicators[action.IndicatorID] {
			return bindingFailure("plan.selected", "unique non-empty indicator IDs", action.IndicatorID)
		}
		seenIndicators[action.IndicatorID] = true
		if err := validateActionBinding(plan.Registry, action); err != nil {
			return err
		}
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
	if err := compareBinding(expected, candidates[0]); err != nil {
		return err
	}
	return compareBinding(expected, actionBinding(action))
}

func actionBinding(action generation.Action) generation.Binding {
	return generation.Binding{
		Operation: action.Operation, Activity: action.Activity, Output: action.Output,
		IndependenceGroupID: action.IndependenceGroupID, ProofChoice: action.ProofChoice,
		Executor: action.Executor, Evaluator: action.Evaluator,
		RequiredIndicatorIDs: append([]string{}, action.RequiredIndicatorIDs...),
		ReceiptRequired: action.ReceiptRequired, Priority: action.Priority,
	}
}

func compareBinding(expected, observed generation.Binding) error {
	if expected.Operation != observed.Operation {
		return bindingFailure(bindingPath(expected.Operation)+".operation", string(expected.Operation), string(observed.Operation))
	}
	checks := []struct {
		path, expected, observed string
	}{
		{"activity", expected.Activity, observed.Activity}, {"output", expected.Output, observed.Output},
		{"executor", expected.Executor, observed.Executor}, {"evaluator", expected.Evaluator, observed.Evaluator},
		{"independence_group_id", expected.IndependenceGroupID, observed.IndependenceGroupID},
		{"proof_choice", string(expected.ProofChoice), string(observed.ProofChoice)},
	}
	for _, check := range checks {
		if check.expected != check.observed {
			return bindingFailure(bindingPath(expected.Operation)+"."+check.path, check.expected, check.observed)
		}
	}
	if !reflect.DeepEqual(sorted(expected.RequiredIndicatorIDs), sorted(observed.RequiredIndicatorIDs)) ||
		!observed.ReceiptRequired {
		return bindingFailure(bindingPath(expected.Operation)+".required_indicator_ids", strings.Join(sorted(expected.RequiredIndicatorIDs), ","), strings.Join(sorted(observed.RequiredIndicatorIDs), ","))
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
		ledger.HeadSHA != plan.HeadSHA || ledger.SelectedPlanOperations != len(plan.Selected) ||
		ledger.BoundExecutorOperations != len(plan.Selected) || ledger.UnboundExecutorOperations != 0 || len(ledger.Effects) != len(plan.Selected) {
		return bindingFailure("ledger", "selected=bound, unbound=0, one effect per selected", "ledger counters or context mismatch")
	}
	actions := make(map[string]generation.Action, len(plan.Selected))
	for _, action := range plan.Selected {
		actions[action.IndicatorID] = action
	}
	seen := make(map[string]bool, len(ledger.Effects))
	for _, effect := range ledger.Effects {
		action, ok := actions[effect.ActionIndicatorID]
		if !ok || seen[effect.ActionIndicatorID] {
			return bindingFailure("ledger.effects["+effect.ActionIndicatorID+"]", "one selected action effect", "unknown or duplicate effect")
		}
		seen[effect.ActionIndicatorID] = true
		if err := validateEffectEntry(effect, action); err != nil {
			return err
		}
	}
	for _, action := range plan.Selected {
		if !seen[action.IndicatorID] {
			return bindingFailure("ledger.effects["+action.IndicatorID+"]", "one selected action effect", "effect missing")
		}
	}
	return nil
}


func validateEffectEntry(effect effect, action generation.Action) error {
	if effect.MetricID != string(action.MetricID) || effect.Subject != action.Subject ||
		effect.Operation != string(action.Operation) || effect.Activity != action.Activity ||
		effect.Output != action.Output || effect.Executor != action.Executor ||
		effect.Evaluator != action.Evaluator || effect.ProofChoice != string(action.ProofChoice) ||
		!validEffectStatus(effect.Status) {
		return bindingFailure("ledger.effects["+action.IndicatorID+"]", "selected action effect binding", effect.Status)
	}
	return nil
}

func validEffectStatus(status string) bool {
	return status == "APPLIED" || status == "REFUTED" || status == "UNKNOWN"
}

func validateEffectOutcomes(plan generation.Plan, ledger ledger, report generation.ReceiptReport) error {
	statuses := make(map[string]string, len(report.Receipts)+len(report.Failures))
	for _, receipt := range report.Receipts {
		if _, exists := statuses[receipt.ActionIndicatorID]; exists {
			return bindingFailure("ledger.effects["+receipt.ActionIndicatorID+"]", "one operation outcome", "duplicate receipt outcome")
		}
		statuses[receipt.ActionIndicatorID] = "APPLIED"
	}
	for _, failure := range report.Failures {
		if _, exists := statuses[failure.ActionIndicatorID]; exists {
			return bindingFailure("ledger.effects["+failure.ActionIndicatorID+"]", "one operation outcome", "duplicate failure outcome")
		}
		statuses[failure.ActionIndicatorID] = failure.Decision
	}
	for _, effect := range ledger.Effects {
		expected, exists := statuses[effect.ActionIndicatorID]
		if !exists || effect.Status != expected {
			return bindingFailure("ledger.effects["+effect.ActionIndicatorID+"]", expected, effect.Status)
		}
	}
	if len(statuses) != len(plan.Selected) {
		return bindingFailure("ledger.effects", fmt.Sprintf("%d operation outcomes", len(plan.Selected)), fmt.Sprintf("%d", len(statuses)))
	}
	return nil
}

func validateReceipts(plan generation.Plan, report generation.ReceiptReport) error {
	if report.SchemaVersion != generation.ReceiptReportSchemaVersion || report.PlanDigest != plan.PlanDigest || report.PromotionAuthorized ||
		len(report.MissingIndicatorIDs) != 0 || len(report.RejectedIndicatorIDs) != 0 ||
		len(report.Receipts)+len(report.Failures) != len(plan.Selected) {
		return bindingFailure("receipts", "one observation per selected action", "receipt set or context mismatch")
	}
	actions := make(map[string]generation.Action, len(plan.Selected))
	for _, action := range plan.Selected {
		actions[action.IndicatorID] = action
	}
	seen := make(map[string]bool, len(plan.Selected))
	for _, receipt := range report.Receipts {
		action, ok := actions[receipt.ActionIndicatorID]
		if !ok || seen[receipt.ActionIndicatorID] {
			return bindingFailure("receipts.receipts["+receipt.ActionIndicatorID+"]", "one selected action", "unknown or duplicate receipt")
		}
		seen[receipt.ActionIndicatorID] = true
		if err := validateClosedReceipt(receipt, action, plan.HeadSHA); err != nil {
			return err
		}
	}
	for _, failure := range report.Failures {
		action, ok := actions[failure.ActionIndicatorID]
		if !ok || seen[failure.ActionIndicatorID] {
			return bindingFailure("receipts.failures["+failure.ActionIndicatorID+"]", "one selected action", "unknown or duplicate failure")
		}
		seen[failure.ActionIndicatorID] = true
		if err := validateObservationFailure(failure, action, plan.HeadSHA); err != nil {
			return err
		}
	}
	for _, action := range plan.Selected {
		if !seen[action.IndicatorID] {
			return bindingFailure("receipts["+action.IndicatorID+"]", "one receipt or failure", "observation missing")
		}
	}
	if err := validateUnknownLinkage(plan, report); err != nil {
		return err
	}
	return validateReceiptDecision(report, len(report.Failures) > 0)
}

func validateClosedReceipt(receipt generation.OperationReceipt, action generation.Action, head string) error {
	if receipt.ActionIndicatorID != action.IndicatorID || receipt.Operation != action.Operation ||
		receipt.Activity != action.Activity || receipt.Output != action.Output || receipt.Executor != action.Executor ||
		receipt.Evaluator != action.Evaluator || receipt.ProofChoice != action.ProofChoice || receipt.InstanceEvidence == nil ||
		len(receipt.Indicators) != len(action.RequiredIndicatorIDs) {
		return bindingFailure("receipts.receipt", "closed selected action", "receipt binding mismatch")
	}
	if err := validateProcessEvidence(receipt.InstanceEvidence, action, head); err != nil {
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

func validateObservationFailure(failure generation.ObservationFailure, action generation.Action, head string) error {
	if failure.Decision != "REFUTED" && failure.Decision != "UNKNOWN" || failure.Stage == "" ||
		failure.Step == "" || failure.Reason == "" || failure.NextOperation == "" || failure.BlockedBy == nil ||
		!validFailureUnknownClass(failure) || !validProcessObservation(failure.Executor) ||
		len(failure.FailureEvidence) != len(action.RequiredIndicatorIDs) ||
		!validCounterexampleRelations(failure.DerivedRelations) {
		return bindingFailure("receipts.failures["+action.IndicatorID+"]", "typed selected-action failure", failure.Reason)
	}
	if failure.Decision == "REFUTED" && !commandMatchesExecutor(failure.Executor.Command, action.Executor) {
		return bindingFailure("receipts.failures["+action.IndicatorID+"].executor.command", action.Executor, strings.Join(failure.Executor.Command, " "))
	}
	if failure.Decision == "REFUTED" &&
		(!commandContainsSubject(failure.Executor.Command, action.Subject) ||
			action.Operation == sourcepolicy.OperationExtractFunction && failure.Executor.ExitCode == 0) {
		return bindingFailure("receipts.failures["+action.IndicatorID+"].executor", "failed command bound to selected subject", "failure process is not bound")
	}
	if err := validateFailureEvidence(failure, action); err != nil {
		return err
	}
	_ = head
	return nil
}

func validFailureUnknownClass(failure generation.ObservationFailure) bool {
	if failure.Decision == "REFUTED" {
		return failure.UnknownClass == ""
	}
	switch failure.UnknownClass {
	case generation.ReceiptUnknownClassDirectMissing, generation.ReceiptUnknownClassMalformedEvidence,
		generation.ReceiptUnknownClassUnexpectedEvidence, generation.ReceiptUnknownClassDependencyBlocked:
		return true
	default:
		return false
	}
}

func validProcessObservation(observation generation.ProcessObservation) bool {
	return validCanonicalCommand(observation.Command) && observation.StdoutBytes >= 0 &&
		observation.StderrBytes >= 0 && validEvidenceDigest(observation.RawStdoutDigest) &&
		validEvidenceDigest(observation.StdoutDigest) && validEvidenceDigest(observation.RawStderrDigest) &&
		validEvidenceDigest(observation.StderrDigest)
}

func validCanonicalCommand(command []string) bool {
	if len(command) == 0 {
		return false
	}
	for _, argument := range command {
		if argument == "" || absoluteCommandArgument(argument) {
			return false
		}
	}
	return true
}

func absoluteCommandArgument(argument string) bool {
	if filepath.IsAbs(argument) || strings.HasPrefix(argument, "//") || strings.HasPrefix(argument, `\\`) {
		return true
	}
	if len(argument) >= 3 && argument[1] == ':' &&
		((argument[0] >= 'a' && argument[0] <= 'z') || (argument[0] >= 'A' && argument[0] <= 'Z')) &&
		(argument[2] == '/' || argument[2] == '\\') {
		return true
	}
	if _, after, ok := strings.Cut(argument, "="); ok {
		return absoluteCommandArgument(after)
	}
	return false
}

func validCounterexampleRelations(relations []generation.CounterexampleRelation) bool {
	seen := make(map[string]bool, len(relations))
	for _, relation := range relations {
		if relation.Counterexample == "" || relation.DerivedFrom == "" ||
			relation.Relation != "DERIVED_FROM" || seen[relation.Counterexample] {
			return false
		}
		seen[relation.Counterexample] = true
	}
	return true
}

func validateFailureEvidence(failure generation.ObservationFailure, action generation.Action) error {
	allowed := map[string]bool{}
	for _, id := range action.RequiredIndicatorIDs {
		allowed[id] = true
	}
	for _, evidence := range failure.FailureEvidence {
		if !allowed[evidence.IndicatorID] || evidence.Decision == "PASS" || evidence.Decision == "" {
			return bindingFailure("receipts.failure_evidence", "typed non-PASS evidence", evidence.IndicatorID)
		}
		delete(allowed, evidence.IndicatorID)
	}
	if len(allowed) != 0 {
		return bindingFailure("receipts.failure_evidence", "all required indicators", "indicator missing")
	}
	return nil
}

func validateUnknownLinkage(plan generation.Plan, report generation.ReceiptReport) error {
	actions := make(map[string]generation.Action, len(plan.Selected))
	for _, action := range plan.Selected {
		actions[action.IndicatorID] = action
	}
	expectedIDs := make([]string, 0, len(report.Failures))
	expectedUnknowns := 0
	for _, failure := range report.Failures {
		action := actions[failure.ActionIndicatorID]
		expectedIDs = append(expectedIDs, action.IndicatorID+"::dependency:"+failure.Reason)
		expectedUnknowns += len(action.RequiredIndicatorIDs)
		if err := validateDependencyUnknowns(report.Unknowns, failure, action); err != nil {
			return err
		}
	}
	if len(report.Unknowns) != expectedUnknowns {
		return bindingFailure("receipts.unknowns", fmt.Sprintf("%d dependency unknowns", expectedUnknowns), fmt.Sprintf("%d", len(report.Unknowns)))
	}
	if !reflect.DeepEqual(sorted(expectedIDs), sorted(report.UnknownIndicatorIDs)) {
		return bindingFailure("receipts.unknown_indicator_ids", strings.Join(sorted(expectedIDs), ","), strings.Join(sorted(report.UnknownIndicatorIDs), ","))
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
		if unknown.ActionIndicatorID != action.IndicatorID {
			continue
		}
		if unknown.ActionIndicatorID != action.IndicatorID || !allowed[unknown.RequiredIndicatorID] ||
			unknown.Stage != failure.Stage || unknown.Step != failure.Step || unknown.Reason != generation.ReceiptReason(failure.Reason) ||
			unknown.UnknownClass != generation.ReceiptUnknownClassDependencyBlocked || unknown.NextOperation != failure.NextOperation ||
			unknown.BlockedBy == nil || len(unknown.BlockedBy) != 1 || unknown.BlockedBy[0] != root {
			return bindingFailure("receipts.unknowns", root, "dependency frontier mismatch")
		}
		delete(allowed, unknown.RequiredIndicatorID)
	}
	if len(allowed) != 0 {
		return bindingFailure("receipts.unknowns["+action.IndicatorID+"]", "one dependency unknown per required indicator", "unknown missing")
	}
	return nil
}

func validateReceiptDecision(report generation.ReceiptReport, hasFailures bool) error {
	if hasFailures {
		for _, failure := range report.Failures {
			if failure.Decision == "REFUTED" {
				if report.Decision != generation.ReceiptDecisionRefuted || report.Reason != generation.ReceiptReasonRefutedOperation {
					return bindingFailure("receipts.decision", string(generation.ReceiptDecisionRefuted), string(report.Decision))
				}
				return nil
			}
		}
		if report.Decision != generation.ReceiptDecisionUnknown {
			return bindingFailure("receipts.decision", string(generation.ReceiptDecisionUnknown), string(report.Decision))
		}
		return nil
	}
	if report.Decision != generation.ReceiptDecisionConformant && report.Decision != generation.ReceiptDecisionFixedPoint {
		return bindingFailure("receipts.decision", "CONFORMANT or FIXED_POINT", string(report.Decision))
	}
	return nil
}

func validateProcessEvidence(evidence *generation.OperationInstanceEvidence, action generation.Action, head string) error {
	if evidence.Schema != generation.OperationInstanceEvidenceSchema || evidence.ActionIndicatorID == "" ||
		evidence.ActionIndicatorID != action.IndicatorID || evidence.Subject != action.Subject ||
		evidence.HeadSHA != head || evidence.OperationID == "" || !evidence.ReplayMatch ||
		evidence.ReplayComparisons < 1 || evidence.ExecutorObservation.Command == nil ||
		evidence.EvaluatorObservation.Command == nil || evidence.VerifierObservation == nil ||
		evidence.VerifierObservation.Command == nil ||
		!commandMatchesExecutor(evidence.ExecutorObservation.Command, action.Executor) ||
		!validEvidenceReuse(*evidence) {
		return bindingFailure("receipts.instance_evidence", "complete process observations", "instance evidence missing")
	}
	if evidence.ExecutorObservation.ExitCode != 0 || evidence.EvaluatorObservation.ExitCode != 0 ||
		evidence.VerifierObservation.ExitCode != 0 {
		return bindingFailure("receipts.instance_evidence.process", "successful selected operation observations", "nonzero process exit")
	}
	if !validProcessObservation(evidence.ExecutorObservation) ||
		!validProcessObservation(evidence.EvaluatorObservation) ||
		!validProcessObservation(*evidence.VerifierObservation) ||
		!commandMatchesEvaluator(evidence.EvaluatorObservation.Command, action.Evaluator) ||
		!hasTestCommand(evidence.VerifierObservation.Command) {
		return bindingFailure("receipts.instance_evidence.process", "canonical process observations", "invalid process observation")
	}
	return nil
}

func commandMatchesExecutor(command []string, executor string) bool {
	if executor == "" {
		return false
	}
	for _, argument := range command {
		if strings.Contains(strings.TrimPrefix(argument, "./"), executor) {
			return true
		}
	}
	return false
}

func commandMatchesEvaluator(command []string, evaluator string) bool {
	if evaluator == "" {
		return false
	}
	for _, argument := range command {
		if strings.Contains(strings.TrimPrefix(argument, "./"), evaluator) {
			return true
		}
	}
	return false
}

func commandContainsSubject(command []string, subject string) bool {
	if subject == "" {
		return false
	}
	for index, argument := range command[:max(0, len(command)-1)] {
		if argument == "-subject" && command[index+1] == subject {
			return true
		}
	}
	return false
}

func validEvidenceReuse(evidence generation.OperationInstanceEvidence) bool {
	if evidence.EvidenceOrigin == "" && evidence.SourceReceiptDigest == "" {
		return true
	}
	return evidence.EvidenceOrigin == generation.EvidenceOriginInputReceipt &&
		validEvidenceDigest(evidence.SourceReceiptDigest)
}

func validEvidenceDigest(value string) bool {
	const prefix = "sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}
	for _, character := range value[len(prefix):] {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func validateProvenance(plan generation.Plan, execution generation.ExecutionManifest, receipts generation.ReceiptReport, provenance provenance) error {
	if provenance.Schema != generation.ArtifactProvenanceSchemaVersion || provenance.BaseSHA != plan.BaseSHA ||
		provenance.HeadSHA != plan.HeadSHA || provenance.PlanDigest != plan.PlanDigest ||
		provenance.ExecutionManifestDigest != execution.ManifestDigest ||
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
	commands, tests, evidenceReuse := processCounts(bundle.Receipts)
	outcome := operationOutcome(bundle.Receipts.Receipts, bundle.Receipts.Failures)
	return Report{Schema: verifierSchema, Decision: "PASS", Resolution: "EXACT", Stage: "verify-bundle",
		Step: "adjudicate-mixed-operation", Reason: "TRANSFORMATION_EXECUTOR_BINDINGS_VERIFIED", NextOperation: "none",
		BlockedBy: []string{}, SelectedPlanOperations: len(bundle.Plan.Selected), BoundExecutorOperations: bundle.Ledger.BoundExecutorOperations,
		UnboundExecutorOperations: bundle.Ledger.UnboundExecutorOperations, ReceiptCount: len(bundle.Receipts.Receipts),
		FailureCount: len(bundle.Receipts.Failures), UnknownCount: len(bundle.Receipts.Unknowns), OperationExecutorCommands: commands,
		OperationExecutorTests: tests, CommandScope: "operation-receipt-executors",
		TestScope: "operation-receipt-verifiers", ReusedCommands: 0, ReusedTests: 0,
		ReusedEvidenceRecords: evidenceReuse, EvidenceReuseScope: "input-receipt-instance-evidence",
		RepositoryWrites: bundle.Runtime.RepositoryWrites, LocalTestExecutions: bundle.Runtime.LocalTestExecutions,
		CrossProjectRequiredGates: bundle.Runtime.CrossProjectRequiredGates, Improvement: "UNKNOWN",
		OperationOutcome: outcome, PromotionAuthorized: false, Runtime: bundle.Runtime}
}

func processCounts(report generation.ReceiptReport) (int, int, int) {
	commands := 0
	tests := 0
	evidenceReuse := 0
	for _, receipt := range report.Receipts {
		if receipt.InstanceEvidence == nil {
			continue
		}
		commands++
		if receipt.InstanceEvidence.EvidenceOrigin == generation.EvidenceOriginInputReceipt {
			evidenceReuse++
		}
		if hasTestCommand(receipt.InstanceEvidence.VerifierObservation.Command) {
			tests++
		}
	}
	for _, failure := range report.Failures {
		if len(failure.Executor.Command) != 0 {
			commands++
		}
	}
	return commands, tests, evidenceReuse
}

func operationOutcome(receipts []generation.OperationReceipt, failures []generation.ObservationFailure) string {
	if len(failures) == 0 {
		return "CLOSED"
	}
	for _, failure := range failures {
		if failure.Decision == "REFUTED" {
			if len(receipts) != 0 {
				return "MIXED_CLOSED_REFUTED"
			}
			return "REFUTED"
		}
	}
	if len(receipts) != 0 {
		return "MIXED_CLOSED_UNKNOWN"
	}
	return "UNKNOWN"
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
