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
	if manifest.Schema != manifestSchema || manifest.Workflow != "CI" || len(manifest.Operations) == 0 {
		return fmt.Errorf("operation manifest is malformed")
	}
	seen := map[string]bool{}
	for _, operation := range manifest.Operations {
		if operation.ID == "" || operation.JobName == "" || operation.StepName == "" || operation.Kind == "" || len(operation.Command) == 0 || seen[operation.ID] {
			return fmt.Errorf("operation manifest identity is invalid")
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
		if !strings.Contains(string(program), "activity "+cell.Activity+"(") || !strings.Contains(string(program), "entity "+entityName(cell.InputID)+" id \""+cell.InputID+"\"") || !strings.Contains(string(program), "entity "+entityName(cell.OutputID)+" id \""+cell.OutputID+"\"") {
			return fmt.Errorf("Gooo activity binding is incomplete for %s", cell.ID)
		}
	}
	return nil
}

func entityName(id string) string {
	parts := strings.Split(id, "/")
	if len(parts) == 0 {
		return ""
	}
	value := parts[len(parts)-1]
	value = strings.ReplaceAll(value, "-", " ")
	words := strings.Fields(value)
	for index := range words {
		words[index] = strings.ToUpper(words[index][:1]) + words[index][1:]
	}
	suffix := "Input"
	if strings.Contains(id, "/output/") {
		suffix = "Output"
	}
	return strings.Join(words, "") + suffix
}

func validateReport(report Report, manifest Manifest, contract Contract, program []byte) error {
	if err := validateStaticInputs(manifest, contract, program); err != nil {
		return err
	}
	if report.Schema != reportSchema || report.ContractID != contract.ID || !validSHA(report.HeadSHA) || report.SourceWorkflow != "CI" || !validDigest(report.OperationManifestDigest) || !validDigest(report.Graph.ProgramDigest) {
		return fmt.Errorf("report identity is malformed")
	}
	if len(report.Operations) != len(manifest.Operations) || report.Accounting.ManifestOperations != len(manifest.Operations) {
		return fmt.Errorf("operation denominator is invalid")
	}
	if report.Window.WallMS <= 0 || report.Window.JobWallMSSum <= 0 || report.Window.StepWallMSSum <= 0 {
		return fmt.Errorf("workflow runtime is incomplete")
	}
	if report.RepositoryStatus.Writes != report.RepositoryWrites || report.RepositoryStatus.Before != "" || report.RepositoryStatus.After != "" || report.LocalTestExecutions != 0 || report.CrossProjectRequiredGates != 0 || report.Improvement != "UNKNOWN" {
		return fmt.Errorf("authority or improvement boundary is invalid")
	}
	if err := validateJobs(report.Jobs, report.HeadSHA); err != nil {
		return err
	}
	if err := validateOperations(report.Operations); err != nil {
		return err
	}
	if report.Accounting.Executed+report.Accounting.Skipped+report.Accounting.Unknown+report.Accounting.Rejected != report.Accounting.ManifestOperations {
		return fmt.Errorf("operation accounting is inconsistent")
	}
	if !validSHA(report.Reuse.Key.HeadSHA) || !validDigest(report.Reuse.Key.InputDigest) || !validDigest(report.Reuse.Key.ToolchainDigest) || !validDigest(report.Reuse.Key.CommandContextDigest) || !validDigest(report.Reuse.Key.EnvironmentAllowlistDigest) || !validDigest(report.Reuse.Key.DependencyGraphDigest) || !validDigest(report.Reuse.Key.ExpectedResultDigest) || !validDigest(report.Reuse.Key.OpenTofuReleaseDigest) {
		return fmt.Errorf("reuse key is incomplete")
	}
	if err := validateReuseCases(report.Reuse.Cases); err != nil {
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
	if report.Decision != "PASS" && report.Decision != "UNKNOWN" && report.Decision != "REFUTED" {
		return fmt.Errorf("unknown top-level decision")
	}
	if report.ReportDigest != sealReport(report) {
		return fmt.Errorf("report digest is not sealed")
	}
	return nil
}

func validateJobs(jobs []JobObservation, head string) error {
	seen := map[int64]bool{}
	for _, job := range jobs {
		if job.ID <= 0 || seen[job.ID] || job.HeadSHA != "" && job.HeadSHA != head || job.WallMS <= 0 {
			return fmt.Errorf("job evidence is invalid")
		}
		seen[job.ID] = true
		for _, step := range job.Steps {
			if step.Name == "" || step.Conclusion == "skipped" {
				continue
			}
			if step.Status == "completed" && step.WallMS <= 0 {
				return fmt.Errorf("step runtime is invalid")
			}
		}
	}
	return nil
}

func validateOperations(operations []OperationObservation) error {
	seen := map[string]bool{}
	for _, operation := range operations {
		if operation.ID == "" || seen[operation.ID] || len(operation.Command) == 0 || operation.EvidenceDigest != operationDigest(operation) {
			return fmt.Errorf("operation evidence is invalid")
		}
		seen[operation.ID] = true
		if operation.State == "" || (operation.State != "EXECUTED" && operation.State != "SKIPPED" && operation.State != "UNKNOWN" && operation.State != "REJECTED") {
			return fmt.Errorf("operation state is invalid")
		}
	}
	return nil
}

func validOpenTofu(value ExternalOpenTofu, head string) bool {
	return value.Workflow == "OpenTofu released-CLI observation" && value.RunID > 0 && value.ArtifactName == "opentofu-observation-"+head && validDigest(value.ArtifactDigest) && value.ArtifactSize > 0 && value.ReportSchema == "gooo/opentofu-observation-report/v1" && value.SubjectSHA == head && value.Decision == "PASS" && value.Resolution == "EXACT" && value.CellsClosed == 12 && value.CellsTotal == 12 && value.ReportDigest != ""
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

func validUnknown(value *Unknown) bool {
	return value != nil && value.Stage != "" && value.Step != "" && value.Reason != "" && value.UnknownClass != "" && value.NextOperation != "" && value.BlockedBy != nil
}

func buildCells(contract Contract, report Report) []CellObservation {
	values := []struct{ observed, expected string; ok bool }{
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
	if report.SourceRunConclusion != "success" || report.Accounting.Rejected > 0 || report.Reuse.Decision == "REFUTED" {
		return "REFUTED", "EXACT", "KNOWN_VERIFICATION_CONTRADICTION"
	}
	if report.Accounting.Unknown > 0 || report.Reuse.Decision == "UNKNOWN" || report.OpenTofu.ArtifactID == 0 {
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
func validDigest(value string) bool { return strings.HasPrefix(value, "sha256:") && len(value) == 71 && validHex(value[7:]) }
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
	fmt.Fprintf(&builder, "observed workflow window=%d ms (%s -> %s)\n", report.Window.WallMS, report.Window.StartAt, report.Window.EndAt)
	fmt.Fprintf(&builder, "job wall sum=%d ms; step wall sum=%d ms\n", report.Window.JobWallMSSum, report.Window.StepWallMSSum)
	fmt.Fprintf(&builder, "operations manifest=%d executed=%d skipped=%d unknown=%d rejected=%d\n", report.Accounting.ManifestOperations, report.Accounting.Executed, report.Accounting.Skipped, report.Accounting.Unknown, report.Accounting.Rejected)
	for _, operation := range report.Operations {
		fmt.Fprintf(&builder, "operation %s kind=%s state=%s job=%q step=%q wall_ms=%d command=%q\n", operation.ID, operation.Kind, operation.State, operation.JobName, operation.StepName, operation.WallMS, operation.Command)
	}
	fmt.Fprintf(&builder, "reuse decision=%s/%s reason=%s requests=%d prior_candidates=%d valid_prior=%d reused=%d rejected=%d unknown=%d skipped=%d reused_commands=%d reused_tests=%d\n", report.Reuse.Decision, report.Reuse.Resolution, report.Reuse.Reason, report.Reuse.Requests, report.Reuse.PriorCandidates, report.Reuse.PriorReceiptsValid, report.Reuse.Reused, report.Reuse.Rejected, report.Reuse.Unknown, report.Reuse.Skipped, report.Reuse.ReusedCommands, report.Reuse.ReusedTests)
	fmt.Fprintf(&builder, "reuse key head=%s input=%s toolchain=%s command_context=%s environment=%s dependency_graph=%s expected_result=%s opentofu_release=%s\n", report.Reuse.Key.HeadSHA, report.Reuse.Key.InputDigest, report.Reuse.Key.ToolchainDigest, report.Reuse.Key.CommandContextDigest, report.Reuse.Key.EnvironmentAllowlistDigest, report.Reuse.Key.DependencyGraphDigest, report.Reuse.Key.ExpectedResultDigest, report.Reuse.Key.OpenTofuReleaseDigest)
	fmt.Fprintf(&builder, "external OpenTofu workflow=%s run=%d artifact=%s/%d digest=%s report=%s cells=%d/%d reuse=%s/%d\n", report.OpenTofu.Workflow, report.OpenTofu.RunID, report.OpenTofu.ArtifactName, report.OpenTofu.ArtifactID, report.OpenTofu.ArtifactDigest, report.OpenTofu.ReportDigest, report.OpenTofu.CellsClosed, report.OpenTofu.CellsTotal, report.OpenTofu.ReuseDecision, report.OpenTofu.ReuseCount)
	fmt.Fprintf(&builder, "repository_writes=%d local_test_executions=%d cross_project_required_gates=%d improvement=%s\n", report.RepositoryWrites, report.LocalTestExecutions, report.CrossProjectRequiredGates, report.Improvement)
	return builder.String()
}
