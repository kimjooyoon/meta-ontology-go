package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func buildReuseKey(config Config, source sourceRunInput, manifest, contract, run, jobs []byte, openTofu ExternalOpenTofu) (ReuseKey, error) {
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
		Operations     []OperationSpec
	}{digestBytes(manifest), mustManifest(manifest).Operations})
	environmentDigest := digestString(config.Environment)
	dependencyParts := map[string][]byte{}
	for _, path := range config.DependencyFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			return ReuseKey{}, fmt.Errorf("read dependency %s: %w", path, err)
		}
		dependencyParts[path] = data
	}
	expectedDigest := digestJSON(struct {
		Workflow   string
		Conclusion string
		Operations []OperationSpec
	}{"CI", source.Conclusion, mustManifest(manifest).Operations})
	return ReuseKey{
		HeadSHA: source.HeadSHA, InputDigest: inputDigest,
		ToolchainDigest: digestString(goToolchain), CommandContextDigest: commandDigest,
		EnvironmentAllowlistDigest: environmentDigest,
		DependencyGraphDigest: digestNamed(dependencyParts), ExpectedResultDigest: expectedDigest,
		OpenTofuReleaseDigest: digestString(openTofu.ReportDigest),
	}, nil
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
		return base, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		base.Decision, base.Resolution, base.Reason = "UNKNOWN", "LOWER", "PRIOR_RECEIPT_MISSING"
		base.Unknown = 1
		base.UnknownEvidence = priorMissingUnknown()
		return base, nil
	}
	base.PriorReceiptDigest = digestBytes(data)
	var prior PriorRecord
	if err := json.Unmarshal(data, &prior); err != nil {
		base.Decision, base.Resolution, base.Reason = "FAIL_CLOSED", "LOWER", "PRIOR_RECEIPT_MALFORMED"
		return base, nil
	}
	base.PriorCandidates, base.PriorReceiptsValid = 1, 1
	base.PriorEvidenceDigest = prior.EvidenceDigest
	if prior.Decision == "REFUTED" || prior.Decision == "UNKNOWN" || prior.Decision == "FAIL_CLOSED" {
		base.Decision, base.Resolution, base.Reason = "REFUTED", "EXACT", "PRIOR_RECEIPT_NOT_REUSABLE"
		base.Rejected = 1
		return base, nil
	}
	if prior.Decision != "PASS" || prior.Resolution != "EXACT" || prior.RepositoryWrites != 0 || !sameReuseKey(prior.Key, key) {
		base.Decision, base.Resolution, base.Reason = "REFUTED", "EXACT", "REUSE_INPUT_DIGEST_MISMATCH"
		base.Rejected = 1
		return base, nil
	}
	base.Decision, base.Resolution, base.Reason = "REUSED", "EXACT", "EXACT_IMMUTABLE_RECEIPT"
	base.Reused = 1
	base.ReusedCommands = 0
	base.ReusedTests = 0
	return base, nil
}

func sameReuseKey(left, right ReuseKey) bool {
	return left.HeadSHA == right.HeadSHA && left.InputDigest == right.InputDigest &&
		left.ToolchainDigest == right.ToolchainDigest && left.CommandContextDigest == right.CommandContextDigest &&
		left.EnvironmentAllowlistDigest == right.EnvironmentAllowlistDigest && left.DependencyGraphDigest == right.DependencyGraphDigest &&
		left.ExpectedResultDigest == right.ExpectedResultDigest && left.OpenTofuReleaseDigest == right.OpenTofuReleaseDigest
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
		return ExternalOpenTofu{}, nil
	}
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		return ExternalOpenTofu{}, nil
	}
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return ExternalOpenTofu{}, nil
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
		ReportDigest: report.ReportDigest, ReportSchema: report.Schema, SubjectSHA: report.SubjectSHA,
		Decision: report.Decision, Resolution: report.Resolution, CellsClosed: report.Summary.ClosedCells,
		CellsTotal: report.Summary.CellsTotal, ReuseDecision: report.Reuse.Decision, ReuseCount: report.Reuse.Reused}, nil
}

func canonicalEnvironment(value string) string { return strings.TrimSpace(value) }
