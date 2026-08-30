package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type expectedCell struct {
	ID, Operation, Proof, Indicator, Activity, Input, Output string
}

func expectedCells() []expectedCell {
	return []expectedCell{
		{"WORKFLOW_WINDOW", "ObserveCIWorkflowWindow", "FOUNDATION", "DRIVER", "ObserveCIWorkflowWindow", "gooo://ci-effort-observation/input/workflow-window", "gooo://ci-effort-observation/output/workflow-window"},
		{"OPERATION_MANIFEST", "BindVerificationOperationManifest", "FOUNDATION", "GUARDRAIL", "BindVerificationOperationManifest", "gooo://ci-effort-observation/input/operation-manifest", "gooo://ci-effort-observation/output/operation-manifest"},
		{"VERIFICATION_RUNTIME", "ObserveVerificationRuntime", "FOUNDATION", "OUTCOME", "ObserveVerificationRuntime", "gooo://ci-effort-observation/input/verification-runtime", "gooo://ci-effort-observation/output/verification-runtime"},
		{"EXECUTED_VERIFICATIONS", "AccountExecutedVerifications", "COHERENCE", "OUTCOME", "AccountExecutedVerifications", "gooo://ci-effort-observation/input/executed-verifications", "gooo://ci-effort-observation/output/executed-verifications"},
		{"SKIPPED_VERIFICATIONS", "AccountSkippedVerifications", "COHERENCE", "DRIVER", "AccountSkippedVerifications", "gooo://ci-effort-observation/input/skipped-verifications", "gooo://ci-effort-observation/output/skipped-verifications"},
		{"REUSE_DECISION", "EvaluateExactEvidenceReuse", "REGRESSION", "GUARDRAIL", "EvaluateExactEvidenceReuse", "gooo://ci-effort-observation/input/reuse-decision", "gooo://ci-effort-observation/output/reuse-decision"},
		{"REUSE_REJECTIONS_UNKNOWN", "RecordReuseRejectionsAndUnknowns", "REGRESSION", "GUARDRAIL", "RecordReuseRejectionsAndUnknowns", "gooo://ci-effort-observation/input/reuse-rejections-unknown", "gooo://ci-effort-observation/output/reuse-rejections-unknown"},
		{"OPENTOFU_BINDING", "BindOpenTofuObservation", "REGRESSION", "OUTCOME", "BindOpenTofuObservation", "gooo://ci-effort-observation/input/opentofu-observation", "gooo://ci-effort-observation/output/opentofu-observation"},
	}
}

func validateStaticInputs(manifest Manifest, contract Contract, program []byte) error {
	if manifest.Schema != manifestSchema || manifest.Workflow != "CI" || manifest.WorkflowSource == "" || len(manifest.Operations) == 0 {
		return fmt.Errorf("operation manifest is malformed")
	}
	seen := map[string]bool{}
	for _, operation := range manifest.Operations {
		if operation.ID == "" || operation.JobName == "" || operation.StepName == "" || operation.Kind == "" || len(operation.Command) == 0 || operation.ProofObligationID == "" || seen[operation.ID] {
			return fmt.Errorf("operation manifest identity is invalid")
		}
		for event, step := range operation.EventStepNames {
			if event == "" || step == "" {
				return fmt.Errorf("operation event step binding is invalid")
			}
		}
		seen[operation.ID] = true
	}
	if contract.Schema != contractSchema || contract.ID == "" || len(contract.Cells) != len(expectedCells()) || contract.GraphProgram == "" {
		return fmt.Errorf("CI effort contract is malformed")
	}
	for index, expected := range expectedCells() {
		cell := contract.Cells[index]
		if cell.ID != expected.ID || cell.MetaOperation != expected.Operation || cell.ProofChoice != expected.Proof || cell.Indicator != expected.Indicator || cell.Activity != expected.Activity || cell.InputID != expected.Input || cell.OutputID != expected.Output {
			return fmt.Errorf("cell %d is not canonical", index)
		}
		if !strings.Contains(string(program), "activity "+cell.Activity+"(") || !hasEntityBinding(program, cell.InputID) || !hasEntityBinding(program, cell.OutputID) {
			return fmt.Errorf("Gooo activity binding is incomplete for %s", cell.ID)
		}
	}
	return nil
}

func hasEntityBinding(program []byte, id string) bool {
	marker := ` id "` + id + `"`
	for line := range strings.SplitSeq(string(program), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "entity ") && strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

func validateReport(report Report, manifest Manifest, contract Contract, program, workflow []byte) error {
	if err := validateStaticInputs(manifest, contract, program); err != nil {
		return err
	}
	workflowDigestValid := validDigest(report.WorkflowSourceDigest) || report.UnknownEvidence != nil && report.WorkflowSourceDigest == ""
	if report.Schema != reportSchema || report.ContractID != contract.ID || !validSHA(report.HeadSHA) || report.SourceWorkflow != "CI" || report.WorkflowSourcePath != manifest.WorkflowSource || !workflowDigestValid || !validDigest(report.OperationManifestDigest) || !validDigest(report.Graph.ProgramDigest) {
		return fmt.Errorf("report identity is malformed")
	}
	if len(workflow) > 0 && report.WorkflowSourceDigest != digestBytes(workflow) {
		return fmt.Errorf("workflow source digest is not bound")
	}
	if len(report.Operations) != len(manifest.Operations) || report.Accounting.ManifestOperations != len(manifest.Operations) {
		return fmt.Errorf("operation denominator is invalid")
	}
	if report.RuntimeResolution != "EXACT" && report.RuntimeResolution != "BOUNDED/SOURCE_SECOND" {
		return fmt.Errorf("runtime resolution is invalid")
	}
	if report.Window.TimestampResolutionMS != 1000 {
		return fmt.Errorf("timestamp resolution is invalid")
	}
	if report.UnknownEvidence == nil && (report.Window.WallMS <= 0 || report.Window.JobWallMSSum <= 0 || report.Window.StepWallMSSum < 0 || report.Window.StepWallMSSum == 0 && report.Window.BelowSourceResolutionSteps == 0) {
		return fmt.Errorf("workflow runtime is incomplete")
	}
	if report.RepositoryStatus.Writes != report.RepositoryWrites || report.LocalTestExecutions != 0 || report.CrossProjectRequiredGates != 0 || report.Improvement != "UNKNOWN" {
		return fmt.Errorf("authority or improvement boundary is invalid")
	}
	if report.UnknownEvidence != nil && !validUnknown(report.UnknownEvidence) {
		return fmt.Errorf("report unknown evidence is incomplete")
	}
	if report.Accounting.Unknown > 0 && report.UnknownEvidence == nil {
		return fmt.Errorf("unknown operation evidence has no causal context")
	}
	if err := validateJobs(report.Jobs, report.HeadSHA); err != nil {
		return err
	}
	if err := validateOperations(report.Operations, manifest.Operations, manifest.WorkflowSource, workflow, report.SourceEvent); err != nil {
		return err
	}
	if report.Accounting.Executed+report.Accounting.Skipped+report.Accounting.Unknown+report.Accounting.Rejected != report.Accounting.ManifestOperations {
		return fmt.Errorf("operation accounting is inconsistent")
	}
	if !validSHA(report.Reuse.Key.HeadSHA) || !validDigest(report.Reuse.Key.InputDigest) || !validDigest(report.Reuse.Key.ToolchainDigest) || !validDigest(report.Reuse.Key.CommandContextDigest) || !validDigest(report.Reuse.Key.EnvironmentAllowlistDigest) || !validDigest(report.Reuse.Key.DependencyGraphDigest) || !validDigest(report.Reuse.Key.ExpectedResultDigest) || (report.OpenTofu.ArtifactID > 0 && !validDigest(report.Reuse.Key.OpenTofuReleaseDigest)) {
		return fmt.Errorf("reuse key is incomplete")
	}
	if err := validateDependencyInputs(report.Reuse.Key); err != nil {
		return err
	}
	if err := validateReuseCases(report.Reuse.Cases); err != nil {
		return err
	}
	if err := validateReuseObservation(report.Reuse); err != nil {
		return err
	}
	if report.Reuse.Decision == "UNKNOWN" && !validUnknown(report.Reuse.UnknownEvidence) {
		return fmt.Errorf("reuse unknown decision is incomplete")
	}
	if len(report.Cells) != len(contract.Cells) || report.Graph.ActivityCount != len(contract.Cells) || report.Graph.BindingCount != len(contract.Cells) || len(report.Graph.Activities) != len(contract.Cells) {
		return fmt.Errorf("meta graph denominator is invalid")
	}
	for index, activity := range report.Graph.Activities {
		if activity != contract.Cells[index].Activity {
			return fmt.Errorf("meta graph activity order is invalid")
		}
	}
	for index, cell := range report.Cells {
		spec := contract.Cells[index]
		if cell.ID != spec.ID || cell.MetaOperation != spec.MetaOperation || cell.ProofChoice != spec.ProofChoice || cell.Indicator != spec.Indicator || cell.Activity != spec.Activity || cell.InputID != spec.InputID || cell.OutputID != spec.OutputID || cell.EvidenceDigest != cellDigest(cell) {
			return fmt.Errorf("cell evidence binding is invalid for %s", spec.ID)
		}
	}
	if report.OpenTofu.ArtifactID != 0 && !validOpenTofu(report.OpenTofu, report.HeadSHA) {
		return fmt.Errorf("OpenTofu evidence binding is invalid")
	}
	if report.OpenTofu.ArtifactID == 0 && report.OpenTofu.Unknown == nil {
		return fmt.Errorf("missing OpenTofu evidence has no causal context")
	}
	if report.Decision != "PASS" && report.Decision != "UNKNOWN" && report.Decision != "REFUTED" {
		return fmt.Errorf("unknown top-level decision")
	}
	if report.Decision == "PASS" && report.UnknownEvidence != nil {
		return fmt.Errorf("PASS report carries unresolved evidence")
	}
	if report.Decision == "UNKNOWN" && report.UnknownEvidence == nil {
		return fmt.Errorf("UNKNOWN report has no causal context")
	}
	if report.ReportDigest != sealReport(report) {
		return fmt.Errorf("report digest is not sealed")
	}
	return nil
}

func validateJobs(jobs []JobObservation, head string) error {
	seen := map[int64]bool{}
	for _, job := range jobs {
		if job.ID <= 0 || seen[job.ID] || job.HeadSHA != "" && job.HeadSHA != head || job.WallMS <= 0 && job.Unknown == nil && !job.BelowSourceResolution {
			return fmt.Errorf("job evidence is invalid")
		}
		if job.BelowSourceResolution && job.StartedAt == "" || job.BelowSourceResolution && job.StartedAt != job.CompletedAt {
			return fmt.Errorf("job bounded runtime is invalid")
		}
		seen[job.ID] = true
		if job.Unknown != nil && !validUnknown(job.Unknown) {
			return fmt.Errorf("job unknown evidence is incomplete")
		}
		for _, step := range job.Steps {
			if step.Name == "" || step.Conclusion == "skipped" {
				continue
			}
			if step.Status == "completed" && step.WallMS <= 0 && step.Unknown == nil && !step.BelowSourceResolution && step.RejectionReason == "" {
				return fmt.Errorf("step runtime is invalid")
			}
			if step.BelowSourceResolution && step.StartedAt == "" || step.BelowSourceResolution && step.StartedAt != step.CompletedAt {
				return fmt.Errorf("step bounded runtime is invalid")
			}
			if step.RejectionReason != "" && !validRejectionReason(step.RejectionReason) {
				return fmt.Errorf("step rejection evidence is incomplete")
			}
			if step.Unknown != nil && !validUnknown(step.Unknown) {
				return fmt.Errorf("step unknown evidence is incomplete")
			}
		}
	}
	return nil
}

func validateOperations(operations []OperationObservation, specs []OperationSpec, workflowPath string, workflow []byte, sourceEvent string) error {
	seen := map[string]bool{}
	for _, operation := range operations {
		if operation.ID == "" || seen[operation.ID] || len(operation.Command) == 0 || operation.EvidenceDigest != operationDigest(operation) {
			return fmt.Errorf("operation evidence is invalid")
		}
		seen[operation.ID] = true
		var spec *OperationSpec
		for index := range specs {
			if specs[index].ID == operation.ID {
				spec = &specs[index]
				break
			}
		}
		expectedEvidenceStep := ""
		if spec != nil {
			expectedEvidenceStep = operationEvidenceStep(*spec, sourceEvent)
		}
		if spec == nil || operation.ProofObligationID != spec.ProofObligationID || operation.JobName != spec.JobName || operation.StepName != spec.StepName || operation.EvidenceStepName != expectedEvidenceStep || operation.GuardStepName != spec.GuardStepName || !sameStrings(operation.Command, spec.Command) {
			return fmt.Errorf("operation manifest binding is invalid")
		}
		if operation.State == "" || (operation.State != "EXECUTED" && operation.State != "SKIPPED" && operation.State != "UNKNOWN" && operation.State != "REJECTED") {
			return fmt.Errorf("operation state is invalid")
		}
		if operation.State == "UNKNOWN" {
			if !validUnknown(operation.Unknown) {
				return fmt.Errorf("operation unknown evidence is incomplete")
			}
			continue
		}
		if operation.State == "REJECTED" && !validRejectionReason(operation.RejectionReason) {
			return fmt.Errorf("operation rejection evidence is incomplete")
		}
		contextDigest, err := bindWorkflowCommandForEvent(workflow, workflowPath, *spec, sourceEvent)
		if err != nil || !operation.CommandBound || operation.WorkflowSourcePath != workflowPath || operation.WorkflowSourceDigest != digestBytes(workflow) || operation.CommandContextDigest != contextDigest {
			return fmt.Errorf("operation command context is unbound")
		}
		if spec.GuardStepName != "" && (!operation.GuardBound || operation.GuardStepStatus != "completed" || operation.GuardStepConclusion != "success") {
			return fmt.Errorf("operation guard binding is incomplete")
		}
	}
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func validOpenTofu(value ExternalOpenTofu, head string) bool {
	return value.Workflow == "OpenTofu released-CLI observation" && value.RunID > 0 && value.ArtifactName == "opentofu-observation-"+head && validDigest(value.ArtifactDigest) && value.ArtifactSize > 0 && value.ReportSchema == "gooo/opentofu-observation-report/v1" && value.SubjectSHA == head && value.Decision == "PASS" && value.Resolution == "EXACT" && value.CellsClosed == 12 && value.CellsTotal == 12 && validDigest(value.ReleaseAssetDigest) && value.ReportDigest != ""
}

func validateReuseCases(cases []ReuseCase) error {
	expected := fixedReuseCases()
	if len(cases) != len(expected) {
		return fmt.Errorf("reuse counterexample denominator is invalid")
	}
	for index, value := range cases {
		want := expected[index]
		if value.ID != want.ID || value.Decision != want.Decision || value.Resolution != want.Resolution || value.Reason != want.Reason {
			return fmt.Errorf("reuse counterexample %d is not canonical", index)
		}
		if value.Decision == "UNKNOWN" && !validUnknown(value.Unknown) {
			return fmt.Errorf("reuse unknown counterexample is incomplete")
		}
	}
	return nil
}

func validateReuseObservation(value ReuseObservation) error {
	if value.Requests != 1 || value.PriorCandidates < 0 || value.PriorReceiptsValid < 0 || value.Reused < 0 || value.Rejected < 0 || value.Unknown < 0 || value.Skipped < 0 {
		return fmt.Errorf("reuse accounting is invalid")
	}
	switch value.Decision {
	case "EXECUTE":
		if value.RequiresExecution != true || value.Reused != 0 || value.PriorCandidates != 0 || value.NextOperation == "" {
			return fmt.Errorf("execute reuse decision is unbound")
		}
	case "REUSED":
		if value.RequiresExecution || value.Reused != 1 || value.PriorCandidates != 1 || value.PriorReceiptsValid != 1 || value.NextOperation != "" {
			return fmt.Errorf("reused decision is unbound")
		}
	case "REFUTED", "FAIL_CLOSED":
		if !value.RequiresExecution || value.NextOperation == "" {
			return fmt.Errorf("non-reusable decision is unbound")
		}
	case "UNKNOWN":
		if !value.RequiresExecution || !validUnknown(value.UnknownEvidence) || value.NextOperation != value.UnknownEvidence.NextOperation {
			return fmt.Errorf("unknown reuse decision is unbound")
		}
	default:
		return fmt.Errorf("unknown reuse decision")
	}
	return nil
}

func validUnknown(value *Unknown) bool {
	return value != nil && value.Stage != "" && value.Step != "" && value.Reason != "" && value.UnknownClass != "" && value.NextOperation != "" && value.BlockedBy != nil
}

func validRejectionReason(value string) bool {
	switch value {
	case "DUPLICATE_JOB_OBSERVATION", "DUPLICATE_STEP_OBSERVATION", "DUPLICATE_GUARD_STEP_OBSERVATION", "GUARD_STEP_NOT_SUCCESSFUL", "OPERATION_TIMESTAMP_MALFORMED", "OPERATION_DURATION_NEGATIVE":
		return true
	default:
		return false
	}
}

func validateDependencyInputs(key ReuseKey) error {
	seen := make(map[string]bool, len(key.DependencyInputs))
	for _, input := range key.DependencyInputs {
		if input.Path == "" || seen[input.Path] {
			return fmt.Errorf("dependency evidence identity is invalid")
		}
		seen[input.Path] = true
		switch input.State {
		case "ABSENT":
			if input.Digest != "ABSENT" {
				return fmt.Errorf("absent dependency evidence is invalid")
			}
		case "PRESENT":
			if !validDigest(input.Digest) {
				return fmt.Errorf("present dependency evidence is invalid")
			}
		default:
			return fmt.Errorf("dependency evidence state is invalid")
		}
	}
	if len(key.DependencyInputs) > 0 && key.DependencyGraphDigest != digestJSON(key.DependencyInputs) {
		return fmt.Errorf("dependency evidence digest is not sealed")
	}
	return nil
}

func buildCells(contract Contract, report Report) []CellObservation {
	values := []struct {
		observed, expected string
		ok                 bool
	}{
		{fmt.Sprint(report.Window.WallMS), ">0", report.Window.WallMS > 0},
		{report.OperationManifestDigest, "checked-in manifest digest", report.OperationManifestDigest != ""},
		{fmt.Sprintf("jobs=%d;steps=%d", report.Window.JobWallMSSum, report.Window.StepWallMSSum), "observed job and step sums", report.Window.JobWallMSSum > 0 && report.Window.StepWallMSSum > 0},
		{fmt.Sprintf("executed=%d;unknown=%d;rejected=%d", report.Accounting.Executed, report.Accounting.Unknown, report.Accounting.Rejected), "complete operation accounting", report.Accounting.Unknown == 0 && report.Accounting.Rejected == 0},
		{fmt.Sprint(report.Accounting.Skipped), "observed skipped operations", report.Accounting.Skipped >= 0},
		{report.Reuse.Decision, "EXECUTE or REUSED with exact key", report.Reuse.Decision == "EXECUTE" || report.Reuse.Decision == "REUSED"},
		{fmt.Sprintf("rejected=%d;unknown=%d", report.Reuse.Rejected, report.Reuse.Unknown), "typed counterexamples retained", len(report.Reuse.Cases) == 5},
		{report.OpenTofu.ArtifactName, "exact OpenTofu PASS/EXACT artifact", report.OpenTofu.ArtifactID > 0 && validOpenTofu(report.OpenTofu, report.HeadSHA)},
	}
	result := make([]CellObservation, 0, len(contract.Cells))
	for index, spec := range contract.Cells {
		value := values[index]
		cell := CellObservation{ID: spec.ID, MetaOperation: spec.MetaOperation, ProofChoice: spec.ProofChoice,
			Indicator: spec.Indicator, Activity: spec.Activity, InputID: spec.InputID, OutputID: spec.OutputID,
			Observed: value.observed, Expected: value.expected, Decision: "PASS"}
		if !value.ok {
			cell.Decision = "UNKNOWN"
		}
		cell.EvidenceDigest = cellDigest(cell)
		result = append(result, cell)
	}
	return result
}

func cellDigest(cell CellObservation) string {
	cell.EvidenceDigest = ""
	return digestJSON(cell)
}

func operationDigest(operation OperationObservation) string {
	operation.EvidenceDigest = ""
	return digestJSON(operation)
}

func classifyReport(report Report) (string, string, string) {
	if report.SourceRunConclusion != "success" || report.Accounting.Rejected > 0 || report.RepositoryWrites > 0 || report.Reuse.Decision == "REFUTED" || report.Reuse.Decision == "FAIL_CLOSED" {
		return "REFUTED", "EXACT", "KNOWN_VERIFICATION_CONTRADICTION"
	}
	if report.Accounting.Unknown > 0 || report.Reuse.Decision == "UNKNOWN" || report.OpenTofu.ArtifactID == 0 || report.UnknownEvidence != nil {
		return "UNKNOWN", "LOWER", "OBSERVATION_EVIDENCE_UNAVAILABLE"
	}
	return "PASS", "EXACT", "CI_EFFORT_OBSERVED"
}

func sealReport(report Report) string {
	report.ReportDigest = ""
	data, _ := json.Marshal(report)
	return digestBytes(data)
}

func validSHA(value string) bool { return len(value) == 40 && validHex(value) }
func validDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:") && len(value) == 71 && validHex(value[7:])
}
func validHex(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')) {
			return false
		}
	}
	return true
}

func fixedCounterexamples() []Counterexample {
	return []Counterexample{
		{ID: "cache-marker-only", Decision: "REFUTED", Resolution: "EXACT", Reason: "CACHE_MARKER_NOT_EVIDENCE"},
		{ID: "digest-mismatch", Decision: "REFUTED", Resolution: "EXACT", Reason: "REUSE_INPUT_DIGEST_MISMATCH"},
		{ID: "prior-receipt-missing", Decision: "UNKNOWN", Resolution: "LOWER", Reason: "PRIOR_RECEIPT_MISSING", Unknown: priorMissingUnknown()},
		{ID: "prior-receipt-malformed", Decision: "FAIL_CLOSED", Resolution: "LOWER", Reason: "PRIOR_RECEIPT_MALFORMED"},
		{ID: "unknown-prior-decision", Decision: "FAIL_CLOSED", Resolution: "LOWER", Reason: "PRIOR_DECISION_UNRECOGNIZED"},
	}
}

func humanReport(report Report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "CI effort observation: %s / %s (%s)\n", report.Decision, report.Resolution, report.Reason)
	fmt.Fprintf(&builder, "head=%s source_run=%d workflow=%s event=%s\n", report.HeadSHA, report.SourceRunID, report.SourceWorkflow, report.SourceEvent)
	fmt.Fprintf(&builder, "workflow_source=%s digest=%s\n", report.WorkflowSourcePath, report.WorkflowSourceDigest)
	fmt.Fprintf(&builder, "observed workflow window=%d ms (%s -> %s)\n", report.Window.WallMS, report.Window.StartAt, report.Window.EndAt)
	fmt.Fprintf(&builder, "job wall sum=%d ms; step wall sum=%d ms; runtime_resolution=%s timestamp_resolution_ms=%d below_source_resolution_jobs=%d steps=%d\n", report.Window.JobWallMSSum, report.Window.StepWallMSSum, report.RuntimeResolution, report.Window.TimestampResolutionMS, report.Window.BelowSourceResolutionJobs, report.Window.BelowSourceResolutionSteps)
	fmt.Fprintf(&builder, "operations manifest=%d executed=%d skipped=%d unknown=%d rejected=%d\n", report.Accounting.ManifestOperations, report.Accounting.Executed, report.Accounting.Skipped, report.Accounting.Unknown, report.Accounting.Rejected)
	for _, operation := range report.Operations {
		fmt.Fprintf(&builder, "operation %s proof=%s kind=%s state=%s job=%q step=%q evidence_step=%q guard_step=%q guard_bound=%t guard_status=%q/%q wall_ms=%d command_bound=%t command_context=%s command=%q\n", operation.ID, operation.ProofObligationID, operation.Kind, operation.State, operation.JobName, operation.StepName, operation.EvidenceStepName, operation.GuardStepName, operation.GuardBound, operation.GuardStepStatus, operation.GuardStepConclusion, operation.WallMS, operation.CommandBound, operation.CommandContextDigest, operation.Command)
	}
	fmt.Fprintf(&builder, "reuse decision=%s/%s reason=%s requests=%d prior_candidates=%d valid_prior=%d reused=%d rejected=%d unknown=%d skipped=%d reused_commands=%d reused_tests=%d\n", report.Reuse.Decision, report.Reuse.Resolution, report.Reuse.Reason, report.Reuse.Requests, report.Reuse.PriorCandidates, report.Reuse.PriorReceiptsValid, report.Reuse.Reused, report.Reuse.Rejected, report.Reuse.Unknown, report.Reuse.Skipped, report.Reuse.ReusedCommands, report.Reuse.ReusedTests)
	fmt.Fprintf(&builder, "reuse key head=%s input=%s toolchain=%s command_context=%s environment=%s dependency_graph=%s expected_result=%s opentofu_release_asset=%s\n", report.Reuse.Key.HeadSHA, report.Reuse.Key.InputDigest, report.Reuse.Key.ToolchainDigest, report.Reuse.Key.CommandContextDigest, report.Reuse.Key.EnvironmentAllowlistDigest, report.Reuse.Key.DependencyGraphDigest, report.Reuse.Key.ExpectedResultDigest, report.Reuse.Key.OpenTofuReleaseDigest)
	for _, dependency := range report.Reuse.Key.DependencyInputs {
		fmt.Fprintf(&builder, "dependency path=%s state=%s digest=%s\n", dependency.Path, dependency.State, dependency.Digest)
	}
	fmt.Fprintf(&builder, "external OpenTofu workflow=%s run=%d artifact=%s/%d digest=%s release_asset=%s report=%s cells=%d/%d reuse=%s/%d\n", report.OpenTofu.Workflow, report.OpenTofu.RunID, report.OpenTofu.ArtifactName, report.OpenTofu.ArtifactID, report.OpenTofu.ArtifactDigest, report.OpenTofu.ReleaseAssetDigest, report.OpenTofu.ReportDigest, report.OpenTofu.CellsClosed, report.OpenTofu.CellsTotal, report.OpenTofu.ReuseDecision, report.OpenTofu.ReuseCount)
	if report.UnknownEvidence != nil {
		fmt.Fprintf(&builder, "unknown stage=%s step=%s reason=%s class=%s next=%s blocked_by=%q\n", report.UnknownEvidence.Stage, report.UnknownEvidence.Step, report.UnknownEvidence.Reason, report.UnknownEvidence.UnknownClass, report.UnknownEvidence.NextOperation, report.UnknownEvidence.BlockedBy)
	}
	fmt.Fprintf(&builder, "repository_writes=%d local_test_executions=%d cross_project_required_gates=%d improvement=%s\n", report.RepositoryWrites, report.LocalTestExecutions, report.CrossProjectRequiredGates, report.Improvement)
	return builder.String()
}
