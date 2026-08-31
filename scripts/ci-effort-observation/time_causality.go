package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	githubActionsProvider        = "github-actions"
	githubActionsRunClockDomain  = "github.actions.run.api.v1"
	githubActionsJobClockDomain  = "github.actions.job.api.v1"
	timeCausalitySchema          = "gooo/ci-time-causality/v0.1.1"
	timeCausalityContract        = "ci-time-causality-v1"
	timeCausalitySourceDigest    = "sha256:a45675079c7e80c8324578ed999e73ccdb6f553f2c1ce35cac223e2809fb4fc1"
	timeCausalityFixtureDigest   = "sha256:d41f4c6fdc33a5442339d83fdc1f4e68f8fcb317393188aa6b0fe67004d662cd"
	timeCausalityIRDigest        = "sha256:1292a28d338696ff8245a309ca7abcf18ed01991b9eaebba0a4c1ba18c43e5f0"
	timeCausalityEvaluatorDigest = "sha256:cd837bed10b30d0d42d5990d39535c168904f3ce261f03787574666d6be12d6f"
	timeCausalityAggregationRule = "Only same operation_id, run_id, job_id, provider, and clock_domain may form one duration; source-ci and opentofu remain separate observations."
	timeCausalityNegativeReason  = "REFUTED_CLOCK_ORDER"
	timeCausalityClampPolicy     = "FORBIDDEN"
)

type operationInterval struct {
	OperationID string
	RunID       int64
	JobID       int64
	Provider    string
	ClockDomain string
	StartedAt   string
	CompletedAt string
}

func operationIntervalForJob(job APIJob) operationInterval {
	return operationInterval{OperationID: jobOperationID(job), RunID: job.RunID, JobID: job.ID,
		Provider: githubActionsProvider, ClockDomain: githubActionsJobClockDomain,
		StartedAt: job.StartedAt, CompletedAt: job.CompletedAt}
}

func operationIntervalForStep(job APIJob, step APIStep) operationInterval {
	return operationInterval{OperationID: stepOperationID(job, step.Name), RunID: job.RunID, JobID: job.ID,
		Provider: githubActionsProvider, ClockDomain: githubActionsJobClockDomain,
		StartedAt: step.StartedAt, CompletedAt: step.CompletedAt}
}

func observeOperationInterval(interval operationInterval) timestampObservation {
	if interval.OperationID == "" || interval.Provider == "" || interval.ClockDomain == "" {
		return timestampObservation{missing: true}
	}
	return observeTimestamp(interval.StartedAt, interval.CompletedAt)
}

func jobOperationID(job APIJob) string {
	if job.RunID > 0 {
		return fmt.Sprintf("github-actions:source-ci:run:%d:job:%d", job.RunID, job.ID)
	}
	return fmt.Sprintf("github-actions:source-ci:job:%d", job.ID)
}

func stepOperationID(job APIJob, name string) string {
	return jobOperationID(job) + ":step:" + name
}

func sourceRunOperationID(source sourceRunInput) string {
	if source.ID > 0 {
		return fmt.Sprintf("github-actions:source-ci:run:%d", source.ID)
	}
	return "github-actions:source-ci:run"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

type TimeCausalityRelease struct {
	Repository   string `json:"repository"`
	Version      string `json:"version"`
	ReleaseID    int64  `json:"release_id"`
	Immutable    bool   `json:"immutable"`
	TagObjectSHA string `json:"tag_object_sha"`
	TargetCommit string `json:"target_commit"`
	MainRunID    int64  `json:"main_run_id"`
	TagRunID     int64  `json:"tag_run_id"`
}

type TimeCausalityAsset struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	SizeBytes int64  `json:"size_bytes"`
	Digest    string `json:"sha256"`
}

type TimeCausalitySummary struct {
	Total      int `json:"total"`
	Closed     int `json:"closed"`
	Unknown    int `json:"unknown"`
	Refuted    int `json:"refuted"`
	Cells      int `json:"cells"`
	Activities int `json:"activities"`
}

type TimeAuditFailure struct {
	Sequence            int    `json:"sequence"`
	Audit               string `json:"audit"`
	Attempt             int    `json:"attempt"`
	FailureCode         string `json:"failure_code"`
	Failure             string `json:"failure"`
	SuccessCounted      bool   `json:"success_counted"`
	ProductVerification bool   `json:"product_verification"`
	LocalTestExecution  int    `json:"local_test_execution"`
}

type TimeCausalityBinding struct {
	Schema                         string               `json:"schema"`
	ContractID                     string               `json:"contract_id"`
	Release                        TimeCausalityRelease `json:"release"`
	SourcePath                     string               `json:"source_path"`
	SourceInputPath                string               `json:"source_input_path"`
	SourceDigest                   string               `json:"source_digest"`
	ImmutableFixtureDigest         string               `json:"immutable_fixture_digest"`
	SemanticIRDigest               string               `json:"semantic_ir_digest"`
	GeneratedEvaluatorDigest       string               `json:"generated_evaluator_digest"`
	ManifestDigest                 string               `json:"manifest_digest"`
	OutputAssets                   []TimeCausalityAsset `json:"output_assets"`
	Summary                        TimeCausalitySummary `json:"summary"`
	RetryAttempts                  []int                `json:"retry_attempts"`
	AggregationRule                string               `json:"aggregation_rule"`
	NegativeDurationReason         string               `json:"negative_duration_reason"`
	ClampToZeroPolicy              string               `json:"clamp_to_zero_policy"`
	SourceOpenTofuSeparate         bool                 `json:"source_ci_opentofu_separate"`
	AuditPath                      string               `json:"audit_path"`
	AuditDigest                    string               `json:"audit_digest"`
	AuditRecordCount               int                  `json:"audit_record_count"`
	AuditFailureCount              int                  `json:"audit_failure_count"`
	AuditFailures                  []TimeAuditFailure   `json:"audit_failures"`
	V010Preserved                  bool                 `json:"v0_1_0_preserved"`
	FailedAttemptsCountedAsSuccess bool                 `json:"failed_attempts_counted_as_success"`
	GeneratedBindingPath           string               `json:"generated_binding_path"`
	GeneratedBindingDigest         string               `json:"generated_binding_digest"`
	BindingDigest                  string               `json:"binding_digest"`
}

type timeManifestInput struct {
	Schema                    string                                        `json:"schema"`
	ContractID                string                                        `json:"contract_id"`
	SourcePath                string                                        `json:"source_path"`
	SourceDigest              string                                        `json:"source_digest"`
	ImmutableFixtureDigest    string                                        `json:"immutable_fixture_digest"`
	IRDigest                  string                                        `json:"ir_digest"`
	GeneratedEvaluatorDigest  string                                        `json:"generated_evaluator_digest"`
	ActivityCount             int                                           `json:"activity_count"`
	CellCount                 int                                           `json:"cell_count"`
	ActivityCellOneToOne      bool                                          `json:"activity_cell_one_to_one"`
	ArtifactCount             int                                           `json:"artifact_count"`
	Summary                   struct{ Total, Closed, Unknown, Refuted int } `json:"summary"`
	RetryAttempts             []int                                         `json:"retry_attempts"`
	RepositoryWrites          int                                           `json:"repository_writes"`
	LocalTestExecutions       int                                           `json:"local_test_executions"`
	CrossProjectRequiredGates int                                           `json:"cross_project_required_gates"`
	VerificationAuthority     string                                        `json:"verification_authority"`
}

type timeDurationResult struct {
	Category string `json:"category"`
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type timeDurationInput struct {
	Schema          string                                        `json:"schema"`
	ContractID      string                                        `json:"contract_id"`
	Summary         struct{ Total, Closed, Unknown, Refuted int } `json:"summary"`
	Results         []timeDurationResult                          `json:"results"`
	AggregationRule string                                        `json:"aggregation_rule"`
	RetryAttempts   []int                                         `json:"retry_attempts"`
}

type timeClockInput struct {
	Schema  string `json:"schema"`
	Rule    string `json:"rule"`
	Domains []struct {
		ID, Source, Comparable string
		ResolutionMS           int `json:"resolution_ms"`
	} `json:"domains"`
}

type timeReplayInput struct {
	Schema        string `json:"schema"`
	ContractID    string `json:"contract_id"`
	Deterministic bool   `json:"deterministic"`
	ReplayCount   int    `json:"replay_count"`
	Decision      string `json:"decision"`
}

type timeGeneratedBindingInput struct {
	Schema                   string `json:"schema"`
	SourcePath               string `json:"source_path"`
	SourceDigest             string `json:"source_digest"`
	SemanticIRDigest         string `json:"semantic_ir_digest"`
	GeneratedEvaluatorDigest string `json:"generated_evaluator_digest"`
	CIImplementationPath     string `json:"ci_implementation_path"`
	DurationRule             string `json:"duration_rule"`
	AggregationRule          string `json:"aggregation_rule"`
	NegativeDurationReason   string `json:"negative_duration_reason"`
	ClampToZeroPolicy        string `json:"clamp_to_zero_policy"`
}

type timeOperationRecord struct {
	RecordType        string `json:"record_type"`
	CaseID            string `json:"case_id"`
	ObservationID     string `json:"observation_id"`
	OperationID       string `json:"operation_id"`
	RunID             string `json:"run_id"`
	JobID             string `json:"job_id"`
	Provider          string `json:"provider"`
	Scope             string `json:"scope"`
	ClockDomain       string `json:"clock_domain"`
	ArtifactID        string `json:"artifact_id"`
	ArtifactDigest    string `json:"artifact_digest"`
	Attempt           int    `json:"attempt"`
	Decision          string `json:"decision"`
	ArtifactCreatedAt string `json:"artifact_created_at"`
	ArtifactUpdatedAt string `json:"artifact_updated_at"`
}

func loadTimeCausality(root string) (TimeCausalityBinding, error) {
	if root == "" {
		return TimeCausalityBinding{}, fmt.Errorf("time-causality input root is empty")
	}
	sourcePath := filepath.Join(root, "main.gooo")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return TimeCausalityBinding{}, fmt.Errorf("read time-causality source: %w", err)
	}
	if digestBytes(source) != timeCausalitySourceDigest {
		return TimeCausalityBinding{}, fmt.Errorf("time-causality source digest is not exact")
	}
	expected := []TimeCausalityAsset{
		{ID: 537649700, Name: "clock-domains.json", SizeBytes: 770, Digest: "sha256:4d858df4758881a329b6e4c52d76c2d4d2e54f114419b873ba90659fc5c9988b"},
		{ID: 537649697, Name: "duration-receipt.json", SizeBytes: 9519, Digest: "sha256:58b2236d930da8f6fd8724f766b3ca0a26a4d90de51b163194ecc8350d765a67"},
		{ID: 537649701, Name: "operations.ndjson", SizeBytes: 8742, Digest: "sha256:c5e704fa5542a69f9e950982f4fdd492e469c900d3ace3ce0408e05113aa5707"},
		{ID: 537649699, Name: "replay-receipt.json", SizeBytes: 505, Digest: "sha256:ee2b19165fb9d56d2ddbd3f96416ed80c4642508321cf24db36fe2b510733a71"},
		{ID: 537649698, Name: "time-manifest.json", SizeBytes: 4877, Digest: "sha256:67ec377f670f925dbb346bd83eee233ab53a01b05be4b56368517e2fef7554a3"},
		{ID: 537649709, Name: "time-report.md", SizeBytes: 3414, Digest: "sha256:b44c42f2188747fff4dee88458ea04a2e2b54bff32367f24fe2af19226cecf17"},
	}
	files := map[string][]byte{"source": source}
	for _, asset := range expected {
		data, readErr := os.ReadFile(filepath.Join(root, asset.Name))
		if readErr != nil {
			return TimeCausalityBinding{}, fmt.Errorf("read time-causality asset %s: %w", asset.Name, readErr)
		}
		if int64(len(data)) != asset.SizeBytes || digestBytes(data) != asset.Digest {
			return TimeCausalityBinding{}, fmt.Errorf("time-causality asset %s is not exact", asset.Name)
		}
		files[asset.Name] = data
	}
	auditPath := filepath.Join(root, "release-history", "post-release-audit.ndjson")
	auditData, err := os.ReadFile(auditPath)
	if err != nil {
		return TimeCausalityBinding{}, fmt.Errorf("read time-causality audit: %w", err)
	}
	files["post-release-audit.ndjson"] = auditData
	generationPath := filepath.Join(root, "generated", "evaluator-binding.json")
	generationData, err := os.ReadFile(generationPath)
	if err != nil {
		return TimeCausalityBinding{}, fmt.Errorf("read time-causality generated binding: %w", err)
	}
	files["evaluator-binding.json"] = generationData

	var manifest timeManifestInput
	if err := json.Unmarshal(files["time-manifest.json"], &manifest); err != nil {
		return TimeCausalityBinding{}, fmt.Errorf("decode time manifest: %w", err)
	}
	var duration timeDurationInput
	if err := json.Unmarshal(files["duration-receipt.json"], &duration); err != nil {
		return TimeCausalityBinding{}, fmt.Errorf("decode duration receipt: %w", err)
	}
	var clocks timeClockInput
	if err := json.Unmarshal(files["clock-domains.json"], &clocks); err != nil {
		return TimeCausalityBinding{}, fmt.Errorf("decode clock domains: %w", err)
	}
	var replay timeReplayInput
	if err := json.Unmarshal(files["replay-receipt.json"], &replay); err != nil {
		return TimeCausalityBinding{}, fmt.Errorf("decode replay receipt: %w", err)
	}
	var generated timeGeneratedBindingInput
	if err := json.Unmarshal(generationData, &generated); err != nil {
		return TimeCausalityBinding{}, fmt.Errorf("decode generated evaluator binding: %w", err)
	}
	operations, err := decodeTimeOperations(files["operations.ndjson"])
	if err != nil {
		return TimeCausalityBinding{}, err
	}
	retryAttempts, err := validateTimeOperations(operations)
	if err != nil {
		return TimeCausalityBinding{}, err
	}
	auditFailures, auditCount, failedAsSuccess, preserved, err := decodeTimeAudit(auditData)
	if err != nil {
		return TimeCausalityBinding{}, err
	}
	if manifest.Schema != "gooo/ci-time-causality/time-manifest/v1" || manifest.ContractID != timeCausalityContract || manifest.SourcePath != "examples/time-causality/main.gooo" || manifest.SourceDigest != timeCausalitySourceDigest || manifest.ImmutableFixtureDigest != timeCausalityFixtureDigest || manifest.IRDigest != timeCausalityIRDigest || manifest.GeneratedEvaluatorDigest != timeCausalityEvaluatorDigest || manifest.ActivityCount != 12 || manifest.CellCount != 12 || !manifest.ActivityCellOneToOne || manifest.ArtifactCount != 6 || manifest.Summary.Total != 12 || manifest.Summary.Closed != 3 || manifest.Summary.Unknown != 4 || manifest.Summary.Refuted != 5 || !sameInts(manifest.RetryAttempts, []int{1, 2}) || manifest.VerificationAuthority != "github-actions" || manifest.RepositoryWrites != 0 || manifest.LocalTestExecutions != 0 || manifest.CrossProjectRequiredGates != 0 {
		return TimeCausalityBinding{}, fmt.Errorf("time manifest binding is not exact")
	}
	if duration.Schema != "gooo/ci-time-causality/duration-receipt/v1" || duration.ContractID != timeCausalityContract || duration.Summary.Total != 12 || duration.Summary.Closed != 3 || duration.Summary.Unknown != 4 || duration.Summary.Refuted != 5 || duration.AggregationRule != timeCausalityAggregationRule || !sameInts(duration.RetryAttempts, []int{1, 2}) || len(duration.Results) != 12 || len(closedResults(duration.Results, "CLOSED")) != 3 || len(closedResults(duration.Results, "UNKNOWN")) != 4 || len(closedResults(duration.Results, "REFUTED")) != 5 {
		return TimeCausalityBinding{}, fmt.Errorf("time duration receipt binding is not exact")
	}
	if clocks.Schema != "gooo/ci-time-causality/clock-domains/v1" || clocks.Rule == "" || len(clocks.Domains) != 3 || replay.Schema != "gooo/ci-time-causality/replay/v1" || replay.ContractID != timeCausalityContract || !replay.Deterministic || replay.ReplayCount != 2 || replay.Decision != "CLOSED" || generated.Schema != "gooo/ci-time-causality/generated-evaluator-binding/v1" || generated.SourcePath != manifest.SourcePath || generated.SourceDigest != timeCausalitySourceDigest || generated.SemanticIRDigest != timeCausalityIRDigest || generated.GeneratedEvaluatorDigest != timeCausalityEvaluatorDigest || generated.CIImplementationPath != "scripts/ci-effort-observation/time_causality.go" || generated.DurationRule == "" || generated.AggregationRule != timeCausalityAggregationRule || generated.NegativeDurationReason != timeCausalityNegativeReason || generated.ClampToZeroPolicy != timeCausalityClampPolicy {
		return TimeCausalityBinding{}, fmt.Errorf("time generated evaluator binding is not exact")
	}
	binding := TimeCausalityBinding{Schema: timeCausalitySchema, ContractID: timeCausalityContract,
		Release:    TimeCausalityRelease{Repository: "kimjooyoon/gooo-ci-time-causality", Version: "v0.1.1", ReleaseID: 379586518, Immutable: true, TagObjectSHA: "bdea52f17b0bfe01f5b448c7d4ceffedc7e13540", TargetCommit: "59b72a990b473199af81b8714b107798ab0533aa", MainRunID: 33370026257, TagRunID: 33370074218},
		SourcePath: manifest.SourcePath, SourceInputPath: filepath.ToSlash(sourcePath), SourceDigest: manifest.SourceDigest, ImmutableFixtureDigest: manifest.ImmutableFixtureDigest,
		SemanticIRDigest: manifest.IRDigest, GeneratedEvaluatorDigest: manifest.GeneratedEvaluatorDigest, ManifestDigest: digestBytes(files["time-manifest.json"]), OutputAssets: expected,
		Summary: TimeCausalitySummary{Total: 12, Closed: 3, Unknown: 4, Refuted: 5, Cells: 12, Activities: 12}, RetryAttempts: retryAttempts,
		AggregationRule: timeCausalityAggregationRule, NegativeDurationReason: timeCausalityNegativeReason, ClampToZeroPolicy: timeCausalityClampPolicy, SourceOpenTofuSeparate: true,
		AuditPath: filepath.ToSlash(auditPath), AuditDigest: digestBytes(auditData), AuditRecordCount: auditCount, AuditFailureCount: len(auditFailures), AuditFailures: auditFailures, V010Preserved: preserved, FailedAttemptsCountedAsSuccess: failedAsSuccess,
		GeneratedBindingPath: filepath.ToSlash(generationPath), GeneratedBindingDigest: digestBytes(generationData),
		BindingDigest: digestNamed(files)}
	if err := validateTimeCausalityBinding(binding); err != nil {
		return TimeCausalityBinding{}, err
	}
	return binding, nil
}

func decodeTimeOperations(data []byte) ([]timeOperationRecord, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	var records []timeOperationRecord
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			return nil, fmt.Errorf("time operations contain a blank record")
		}
		var record timeOperationRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("decode time operation: %w", err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func validateTimeOperations(records []timeOperationRecord) ([]int, error) {
	if len(records) != 16 {
		return nil, fmt.Errorf("time operation denominator is invalid")
	}
	caseIDs := map[string]bool{}
	var retry []timeOperationRecord
	for _, record := range records {
		if record.CaseID == "" || record.Decision == "" {
			return nil, fmt.Errorf("time operation identity is incomplete")
		}
		if record.RecordType == "retry_lineage" {
			retry = append(retry, record)
		} else if record.RecordType == "observation" {
			caseIDs[record.CaseID] = true
		} else {
			return nil, fmt.Errorf("time operation record type is invalid")
		}
	}
	if len(caseIDs) != 12 || len(retry) != 2 {
		return nil, fmt.Errorf("time operation case or retry denominator is invalid")
	}
	sort.Slice(retry, func(i, j int) bool { return retry[i].Attempt < retry[j].Attempt })
	if retry[0].CaseID != "immutable-counterexample" || retry[1].CaseID != retry[0].CaseID || retry[0].Attempt != 1 || retry[1].Attempt != 2 || retry[0].Decision != "REFUTED" || retry[1].Decision != "REFUTED" || retry[0].RunID != "33365730015" || retry[1].RunID != "33365730015" || retry[0].JobID != "99405870188" || retry[1].JobID != "99408612206" || retry[0].ArtifactID != "9748462083" || retry[1].ArtifactID != "9748520364" || retry[0].ArtifactDigest != "sha256:74b24a4c24acd853f5d661aa5ad88b64f56dd8186d98288f87981fd7b4bd3979" || retry[1].ArtifactDigest != "sha256:1d0218ed0945ef76ea9cc396ad7a9e0916bc6243ebf6be0fe04efc8e9878b656" || retry[0].ClockDomain != githubActionsJobClockDomain || retry[1].ClockDomain != githubActionsJobClockDomain {
		return nil, fmt.Errorf("CI-effort retry lineage is not exact")
	}
	return []int{retry[0].Attempt, retry[1].Attempt}, nil
}

func closedResults(results []timeDurationResult, category string) []timeDurationResult {
	result := make([]timeDurationResult, 0)
	for _, value := range results {
		if value.Category == category && value.Decision == category {
			result = append(result, value)
		}
	}
	return result
}

func decodeTimeAudit(data []byte) ([]TimeAuditFailure, int, bool, bool, error) {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	count := 0
	failedAsSuccess := false
	preserved := false
	var failures []TimeAuditFailure
	for scanner.Scan() {
		count++
		var record struct {
			Sequence                       int    `json:"sequence"`
			RecordType                     string `json:"record_type"`
			Audit                          string `json:"audit"`
			Attempt                        int    `json:"attempt"`
			Status                         string `json:"status"`
			SuccessCounted                 bool   `json:"success_counted"`
			ProductVerification            bool   `json:"product_verification"`
			LocalTestExecution             int    `json:"local_test_execution"`
			FailureCode                    string `json:"failure_code"`
			Failure                        string `json:"failure"`
			V010Preserved                  bool   `json:"v0_1_0_deleted_or_recreated"`
			FailedAttemptsCountedAsSuccess bool   `json:"failed_attempts_counted_as_success"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, 0, false, false, fmt.Errorf("decode time audit: %w", err)
		}
		if record.RecordType == "release_audit_attempt" && record.Status == "FAILED" {
			failures = append(failures, TimeAuditFailure{Sequence: record.Sequence, Audit: record.Audit, Attempt: record.Attempt, FailureCode: record.FailureCode, Failure: record.Failure, SuccessCounted: record.SuccessCounted, ProductVerification: record.ProductVerification, LocalTestExecution: record.LocalTestExecution})
		}
		if record.RecordType == "audit_summary" {
			failedAsSuccess = record.FailedAttemptsCountedAsSuccess
			preserved = !record.V010Preserved
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, false, false, err
	}
	return failures, count, failedAsSuccess, preserved, nil
}

func validateTimeCausalityBinding(binding TimeCausalityBinding) error {
	if binding.Schema != timeCausalitySchema || binding.ContractID != timeCausalityContract || binding.Release.Repository != "kimjooyoon/gooo-ci-time-causality" || binding.Release.Version != "v0.1.1" || binding.Release.ReleaseID != 379586518 || !binding.Release.Immutable || binding.Release.TagObjectSHA != "bdea52f17b0bfe01f5b448c7d4ceffedc7e13540" || binding.Release.TargetCommit != "59b72a990b473199af81b8714b107798ab0533aa" || binding.Release.MainRunID != 33370026257 || binding.Release.TagRunID != 33370074218 {
		return fmt.Errorf("time-causality release binding is not exact")
	}
	if binding.SourcePath != "examples/time-causality/main.gooo" || binding.SourceDigest != timeCausalitySourceDigest || binding.ImmutableFixtureDigest != timeCausalityFixtureDigest || binding.SemanticIRDigest != timeCausalityIRDigest || binding.GeneratedEvaluatorDigest != timeCausalityEvaluatorDigest || !validDigest(binding.ManifestDigest) || len(binding.OutputAssets) != 6 || binding.Summary != (TimeCausalitySummary{Total: 12, Closed: 3, Unknown: 4, Refuted: 5, Cells: 12, Activities: 12}) || !sameInts(binding.RetryAttempts, []int{1, 2}) || binding.AggregationRule != timeCausalityAggregationRule || binding.NegativeDurationReason != timeCausalityNegativeReason || binding.ClampToZeroPolicy != timeCausalityClampPolicy || !binding.SourceOpenTofuSeparate || binding.AuditRecordCount != 6 || binding.AuditFailureCount != 2 || !binding.V010Preserved || binding.FailedAttemptsCountedAsSuccess {
		return fmt.Errorf("time-causality semantic binding is not exact")
	}
	if binding.AuditFailures[0].FailureCode != "UNDEFINED_SCHEMA_FIELD" || binding.AuditFailures[1].FailureCode != "SHELL_VARIABLE_PATH_COLLISION" || binding.AuditFailures[0].SuccessCounted || binding.AuditFailures[1].SuccessCounted || !validDigest(binding.AuditDigest) || !validDigest(binding.GeneratedBindingDigest) || !validDigest(binding.BindingDigest) {
		return fmt.Errorf("time-causality append-only audit binding is not exact")
	}
	for index, asset := range binding.OutputAssets {
		if asset.ID <= 0 || asset.Name == "" || asset.SizeBytes <= 0 || !validDigest(asset.Digest) || asset != timeCausalityExpectedAsset(index) {
			return fmt.Errorf("time-causality output asset binding is not exact")
		}
	}
	return nil
}

func timeCausalityExpectedAsset(index int) TimeCausalityAsset {
	return []TimeCausalityAsset{
		{ID: 537649700, Name: "clock-domains.json", SizeBytes: 770, Digest: "sha256:4d858df4758881a329b6e4c52d76c2d4d2e54f114419b873ba90659fc5c9988b"},
		{ID: 537649697, Name: "duration-receipt.json", SizeBytes: 9519, Digest: "sha256:58b2236d930da8f6fd8724f766b3ca0a26a4d90de51b163194ecc8350d765a67"},
		{ID: 537649701, Name: "operations.ndjson", SizeBytes: 8742, Digest: "sha256:c5e704fa5542a69f9e950982f4fdd492e469c900d3ace3ce0408e05113aa5707"},
		{ID: 537649699, Name: "replay-receipt.json", SizeBytes: 505, Digest: "sha256:ee2b19165fb9d56d2ddbd3f96416ed80c4642508321cf24db36fe2b510733a71"},
		{ID: 537649698, Name: "time-manifest.json", SizeBytes: 4877, Digest: "sha256:67ec377f670f925dbb346bd83eee233ab53a01b05be4b56368517e2fef7554a3"},
		{ID: 537649709, Name: "time-report.md", SizeBytes: 3414, Digest: "sha256:b44c42f2188747fff4dee88458ea04a2e2b54bff32367f24fe2af19226cecf17"},
	}[index]
}

func sameInts(left, right []int) bool {
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
