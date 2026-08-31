package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func buildReuseKey(config Config, source sourceRunInput, manifest, contract, run, jobs []byte, openTofu ExternalOpenTofu, timeCausality TimeCausalityBinding, workflow []byte, operations []OperationObservation) (ReuseKey, error) {
	summary, err := os.ReadFile(config.SummaryPath)
	if err != nil {
		return ReuseKey{}, fmt.Errorf("read CI summary: %w", err)
	}
	evidence, err := os.ReadFile(config.EvidencePath)
	if err != nil {
		return ReuseKey{}, fmt.Errorf("read CI evidence: %w", err)
	}
	parts := map[string][]byte{"contract": contract, "evidence": evidence, "jobs": jobs, "manifest": manifest, "run": run, "summary": summary}
	inputDigest := digestNamed(parts)
	commandDigest := digestJSON(struct {
		ManifestDigest string
		WorkflowDigest string
		Operations     []commandContext
	}{digestBytes(manifest), digestIfPresent(workflow), commandContexts(operations)})
	environmentDigest := digestString(config.Environment)
	dependencyInputs, err := readDependencyInputs(config.DependencyFiles)
	if err != nil {
		return ReuseKey{}, err
	}
	expectedDigest := digestJSON(struct {
		Workflow   string
		Conclusion string
		Operations []OperationSpec
	}{"CI", source.Conclusion, mustManifest(manifest).Operations})
	return ReuseKey{
		HeadSHA: source.HeadSHA, SourceEvent: source.Event, InputDigest: inputDigest,
		ToolchainDigest: digestString(goToolchain), CommandContextDigest: commandDigest,
		EnvironmentAllowlistDigest: environmentDigest,
		DependencyGraphDigest:      digestJSON(dependencyInputs), DependencyInputs: dependencyInputs, ExpectedResultDigest: expectedDigest,
		OpenTofuReleaseDigest: openTofu.ReleaseAssetDigest, TimeCausalityDigest: timeCausality.BindingDigest,
	}, nil
}

func readDependencyInputs(paths []string) ([]DependencyInput, error) {
	result := make([]DependencyInput, 0, len(paths))
	seen := make(map[string]bool, len(paths))
	for _, path := range paths {
		if path == "" || seen[path] {
			return nil, fmt.Errorf("dependency path is empty or duplicated")
		}
		seen[path] = true
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			result = append(result, DependencyInput{Path: path, State: "ABSENT", Digest: "ABSENT"})
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("inspect dependency %s: %w", path, err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read dependency %s: %w", path, err)
		}
		result = append(result, DependencyInput{Path: path, State: "PRESENT", Digest: digestBytes(data)})
	}
	return result, nil
}

type commandContext struct {
	ID, ProofObligationID, SourceEvent, JobName, StepName, WorkflowPath, WorkflowDigest, ContextDigest string
	Command                                                                                            []string
}

func commandContexts(operations []OperationObservation) []commandContext {
	result := make([]commandContext, 0, len(operations))
	for _, operation := range operations {
		result = append(result, commandContext{ID: operation.ID, ProofObligationID: operation.ProofObligationID, SourceEvent: operation.SourceEvent,
			JobName: operation.JobName, StepName: operation.StepName, WorkflowPath: operation.WorkflowSourcePath,
			WorkflowDigest: operation.WorkflowSourceDigest, ContextDigest: operation.CommandContextDigest,
			Command: operation.Command})
	}
	return result
}

func mustManifest(data []byte) Manifest {
	var manifest Manifest
	_ = json.Unmarshal(data, &manifest)
	return manifest
}

func buildReuse(path string, key ReuseKey) (ReuseObservation, error) {
	base := ReuseObservation{Requests: 1, Key: key, Cases: fixedReuseCases()}
	if path == "" {
		base.Decision, base.Resolution, base.Reason = "EXECUTE", "EXACT", "NO_PRIOR_RECEIPT"
		base.RequiresExecution, base.NextOperation = true, "EXECUTE_VERIFIED_OPERATIONS"
		return base, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		base.Decision, base.Resolution, base.Reason = "UNKNOWN", "LOWER", "PRIOR_RECEIPT_MISSING"
		base.Unknown = 1
		base.UnknownEvidence = priorMissingUnknown()
		base.RequiresExecution, base.NextOperation = true, base.UnknownEvidence.NextOperation
		return base, nil
	}
	base.PriorReceiptDigest = digestBytes(data)
	var prior PriorRecord
	if err := json.Unmarshal(data, &prior); err != nil {
		base.Decision, base.Resolution, base.Reason = "FAIL_CLOSED", "LOWER", "PRIOR_RECEIPT_MALFORMED"
		base.RequiresExecution, base.NextOperation = true, "REBUILD_PRIOR_RECEIPT"
		return base, nil
	}
	base.PriorCandidates, base.PriorReceiptsValid = 1, 1
	base.PriorEvidenceDigest = prior.EvidenceDigest
	if prior.Decision == "REFUTED" || prior.Decision == "UNKNOWN" || prior.Decision == "FAIL_CLOSED" {
		base.Decision, base.Resolution, base.Reason = "REFUTED", "EXACT", "PRIOR_RECEIPT_NOT_REUSABLE"
		base.Rejected = 1
		base.RequiresExecution, base.NextOperation = true, "EXECUTE_VERIFIED_OPERATIONS"
		return base, nil
	}
	if prior.Schema != reportSchema || prior.HeadSHA != key.HeadSHA || !validDigest(prior.EvidenceDigest) || !validDigest(prior.ResultDigest) || prior.Decision != "PASS" || prior.Resolution != "EXACT" || prior.RepositoryWrites != 0 || !sameReuseKey(prior.Key, key) {
		base.Decision, base.Resolution, base.Reason = "REFUTED", "EXACT", "REUSE_INPUT_DIGEST_MISMATCH"
		base.Rejected = 1
		base.RequiresExecution, base.NextOperation = true, "EXECUTE_VERIFIED_OPERATIONS"
		return base, nil
	}
	base.Decision, base.Resolution, base.Reason = "REUSED", "EXACT", "EXACT_IMMUTABLE_RECEIPT"
	base.Reused = 1
	base.ReusedCommands = 0
	base.ReusedTests = 0
	base.RequiresExecution = false
	return base, nil
}

func sameReuseKey(left, right ReuseKey) bool {
	return left.HeadSHA == right.HeadSHA && left.SourceEvent == right.SourceEvent && left.InputDigest == right.InputDigest &&
		left.ToolchainDigest == right.ToolchainDigest && left.CommandContextDigest == right.CommandContextDigest &&
	left.EnvironmentAllowlistDigest == right.EnvironmentAllowlistDigest && left.DependencyGraphDigest == right.DependencyGraphDigest &&
		sameDependencyInputs(left.DependencyInputs, right.DependencyInputs) &&
		left.ExpectedResultDigest == right.ExpectedResultDigest && left.OpenTofuReleaseDigest == right.OpenTofuReleaseDigest && left.TimeCausalityDigest == right.TimeCausalityDigest
}

func sameDependencyInputs(left, right []DependencyInput) bool {
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

func fixedReuseCases() []ReuseCase {
	return []ReuseCase{
		{ID: "first-run-no-prior", Decision: "EXECUTE", Resolution: "EXACT", Reason: "NO_PRIOR_RECEIPT"},
		{ID: "exact-immutable-prior", Decision: "REUSED", Resolution: "EXACT", Reason: "EXACT_IMMUTABLE_RECEIPT"},
		{ID: "digest-mismatch", Decision: "REFUTED", Resolution: "EXACT", Reason: "REUSE_INPUT_DIGEST_MISMATCH"},
		{ID: "cache-marker-only", Decision: "REFUTED", Resolution: "EXACT", Reason: "CACHE_MARKER_NOT_EVIDENCE"},
		{ID: "missing-prior", Decision: "UNKNOWN", Resolution: "LOWER", Reason: "PRIOR_RECEIPT_MISSING", Unknown: priorMissingUnknown()},
	}
}

func priorMissingUnknown() *Unknown {
	return &Unknown{Stage: "REUSE", Step: "READ_PRIOR_RECEIPT", Reason: "PRIOR_RECEIPT_MISSING",
		UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_PRIOR_RECEIPT", BlockedBy: []string{}}
}

func observeOpenTofu(reportPath, metaPath, expectedHead string) (ExternalOpenTofu, error) {
	if reportPath == "" || metaPath == "" {
		return ExternalOpenTofu{Unknown: openTofuUnknown("OPENTOFU_EVIDENCE_MISSING")}, nil
	}
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		return ExternalOpenTofu{Unknown: openTofuUnknown("OPENTOFU_REPORT_MISSING")}, nil
	}
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return ExternalOpenTofu{Unknown: openTofuUnknown("OPENTOFU_ARTIFACT_METADATA_MISSING")}, nil
	}
	var report OpenTofuReportInput
	var meta ArtifactMeta
	if json.Unmarshal(reportData, &report) != nil || json.Unmarshal(metaData, &meta) != nil {
		return ExternalOpenTofu{}, nil
	}
	if report.SubjectSHA != expectedHead || report.Schema != "gooo/opentofu-observation-report/v1" || meta.ID <= 0 || meta.Name == "" || meta.Digest == "" || meta.Size <= 0 || meta.Expired {
		return ExternalOpenTofu{}, nil
	}
	return ExternalOpenTofu{Workflow: "OpenTofu released-CLI observation", RunID: meta.RunID, ArtifactID: meta.ID,
		ArtifactName: meta.Name, ArtifactDigest: meta.Digest, ArtifactSize: meta.Size,
		ReportDigest: report.ReportDigest, ReleaseAssetDigest: report.Release.AssetSHA256,
		ReportSchema: report.Schema, SubjectSHA: report.SubjectSHA,
		Decision: report.Decision, Resolution: report.Resolution, CellsClosed: report.Summary.ClosedCells,
		CellsTotal: report.Summary.CellsTotal, ReuseDecision: report.Reuse.Decision, ReuseCount: report.Reuse.Reused}, nil
}

func canonicalEnvironment(value string) string { return strings.TrimSpace(value) }

func openTofuUnknown(reason string) *Unknown {
	return &Unknown{Stage: "OBSERVE", Step: "READ_OPENTOFU_EVIDENCE", Reason: reason,
		UnknownClass: "DIRECT_MISSING", NextOperation: "RECAPTURE_OPENTOFU_OBSERVATION", BlockedBy: []string{}}
}
