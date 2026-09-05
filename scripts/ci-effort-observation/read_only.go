package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicworkflowlineage"
)

const readOnlyProjectionSchema = "gooo/ci-effort-observation-read-only/v1"

type readOnlyLineageObservation struct {
	Schema       string                           `json:"schema"`
	Decision     string                           `json:"decision"`
	LineageState string                           `json:"lineage_state"`
	Reason       string                           `json:"reason"`
	Trigger      publicworkflowlineage.Trigger    `json:"trigger"`
	Source       publicworkflowlineage.SourceRun  `json:"source_run"`
	Evaluation   publicworkflowlineage.Evaluation `json:"evaluation"`
	PolicyDigest string                           `json:"policy_source_digest"`
}

type readOnlyTimingSummary struct {
	WindowWallMS               int64 `json:"window_wall_ms"`
	ObservedJobIntervals       int   `json:"observed_job_intervals"`
	ObservedStepIntervals      int   `json:"observed_step_intervals"`
	MissingJobIntervals        int   `json:"missing_job_intervals"`
	MissingStepIntervals       int   `json:"missing_step_intervals"`
	JobsDataUnavailable        bool  `json:"jobs_data_unavailable"`
	WindowTimestampsUnknown    bool  `json:"window_timestamps_unknown"`
	BelowSourceResolutionJobs  int   `json:"below_source_resolution_jobs"`
	BelowSourceResolutionSteps int   `json:"below_source_resolution_steps"`
	RuntimeRejectionCount      int   `json:"runtime_rejection_count"`
}

type readOnlyOperationCounts struct {
	Manifest int `json:"manifest"`
	Observed int `json:"observed"`
	Skipped  int `json:"skipped"`
	Missing  int `json:"missing"`
	Rejected int `json:"rejected"`
}

type readOnlyProjection struct {
	Schema                  string                  `json:"schema"`
	Decision                string                  `json:"decision"`
	Resolution              string                  `json:"resolution"`
	Reason                  string                  `json:"reason"`
	LineageState            string                  `json:"lineage_state"`
	LineageReason           string                  `json:"lineage_reason"`
	ObservationReason       string                  `json:"observation_reason"`
	Repository              string                  `json:"repository"`
	SourceWorkflow          string                  `json:"source_workflow"`
	SourceEvent             string                  `json:"source_event"`
	SourceRef               string                  `json:"source_ref"`
	HeadSHA                 string                  `json:"head_sha"`
	SourceRunStatus         string                  `json:"source_run_status"`
	SourceRunConclusion     string                  `json:"source_run_conclusion"`
	SourceRunID             int64                   `json:"source_run_id"`
	SourceRunAttempt        int64                   `json:"source_run_attempt"`
	SourceRunURL            string                  `json:"source_run_url"`
	WorkflowSourcePath      string                  `json:"workflow_source_path"`
	WorkflowSourceDigest    string                  `json:"workflow_source_digest"`
	ContractID              string                  `json:"contract_id"`
	OperationManifestDigest string                  `json:"operation_manifest_digest"`
	ExactSourceIdentity     bool                    `json:"exact_source_identity"`
	SourceFailureKept       bool                    `json:"source_failure_kept"`
	EvidenceReuseAllowed    bool                    `json:"evidence_reuse_allowed"`
	PromotionAllowed        bool                    `json:"promotion_allowed"`
	Timing                  readOnlyTimingSummary   `json:"timing"`
	OperationCounts         readOnlyOperationCounts `json:"operation_counts"`
	Window                  WorkflowWindow          `json:"workflow_window"`
	Jobs                    []JobObservation        `json:"jobs"`
	Operations              []OperationObservation  `json:"operations"`
	Accounting              Accounting              `json:"accounting"`
	Graph                   GraphObservation        `json:"graph"`
	RepositoryWrites        int                     `json:"repository_writes"`
	LocalTestExecutions     int                     `json:"local_test_executions"`
	Improvement             string                  `json:"improvement"`
	HumanReport             string                  `json:"human_report"`
	ReportDigest            string                  `json:"report_digest"`
}

func buildReadOnlyProjection(config Config) (readOnlyProjection, error) {
	if !config.ReadOnly {
		return readOnlyProjection{}, fmt.Errorf("read-only mode is not enabled")
	}
	if config.LineageObservationPath == "" || config.ReadOnlyObservationPath == "" {
		return readOnlyProjection{}, fmt.Errorf("read-only workflow-lineage observations are required")
	}
	var source sourceRunInput
	if _, err := readJSON(config.RunPath, &source); err != nil {
		return readOnlyProjection{}, err
	}
	source.WorkflowName = firstNonEmpty(source.WorkflowName, source.WorkflowPath)
	var jobs JobsInput
	if _, err := readJSON(config.JobsPath, &jobs); err != nil {
		return readOnlyProjection{}, err
	}
	var manifest Manifest
	manifestBytes, err := readJSON(config.ManifestPath, &manifest)
	if err != nil {
		return readOnlyProjection{}, err
	}
	var contract Contract
	if _, err := readJSON(config.ContractPath, &contract); err != nil {
		return readOnlyProjection{}, err
	}
	program, err := os.ReadFile(config.ProgramPath)
	if err != nil {
		return readOnlyProjection{}, err
	}
	if err := validateStaticInputs(manifest, contract, program); err != nil {
		return readOnlyProjection{}, err
	}
	policy, err := publicworkflowlineage.Load(config.ProgramPath, program)
	if err != nil {
		return readOnlyProjection{}, err
	}
	var lineage readOnlyLineageObservation
	if _, err := readJSON(config.LineageObservationPath, &lineage); err != nil {
		return readOnlyProjection{}, err
	}
	var observation publicworkflowlineage.ReadOnlyObservationEvaluation
	if _, err := readJSON(config.ReadOnlyObservationPath, &observation); err != nil {
		return readOnlyProjection{}, err
	}
	if err := validateReadOnlyLineageInputs(source, lineage, observation, policy); err != nil {
		return readOnlyProjection{}, err
	}

	workflow, workflowErr := os.ReadFile(manifest.WorkflowSource)
	var observedJobs []JobObservation
	var window WorkflowWindow
	if len(jobs.Jobs) > 0 {
		observedJobs, window, err = observeJobsWithSource(jobs.Jobs, source)
		if err != nil {
			return readOnlyProjection{}, err
		}
	} else {
		observedJobs = []JobObservation{}
		window = readOnlyWindowForSource(source)
	}
	operations, accounting := observeOperations(manifest.Operations, jobs.Jobs, manifest.WorkflowSource, workflow, workflowErr, source.Event)
	if err := validateJobs(observedJobs, source.HeadSHA); err != nil {
		return readOnlyProjection{}, err
	}
	if err := validateOperations(operations, manifest.Operations, manifest.WorkflowSource, workflow, source.Event); err != nil {
		return readOnlyProjection{}, err
	}
	timing := summarizeReadOnlyTiming(window, observedJobs, len(jobs.Jobs) == 0)
	report := readOnlyProjection{
		Schema: readOnlyProjectionSchema, Decision: publicworkflowlineage.DecisionRefuted, Resolution: "READ_ONLY",
		Reason: "SOURCE_FAILURE_OBSERVED_WITHOUT_REUSE_OR_PROMOTION", LineageState: lineage.LineageState,
		LineageReason: lineage.Reason, ObservationReason: observation.Reason,
		Repository: source.HeadRepository.FullName, SourceWorkflow: manifest.Workflow, SourceEvent: source.Event,
		SourceRef: source.Ref, HeadSHA: source.HeadSHA, SourceRunStatus: source.Status, SourceRunConclusion: source.Conclusion,
		SourceRunID: source.ID, SourceRunAttempt: source.RunAttempt, SourceRunURL: source.HTMLURL,
		WorkflowSourcePath: manifest.WorkflowSource, WorkflowSourceDigest: digestIfPresent(workflow), ContractID: contract.ID,
		OperationManifestDigest: digestBytes(manifestBytes), ExactSourceIdentity: true, SourceFailureKept: true,
		EvidenceReuseAllowed: false, PromotionAllowed: false, Timing: timing, OperationCounts: readOnlyCounts(accounting),
		Window: window, Jobs: observedJobs, Operations: operations, Accounting: accounting,
		Graph: graphObservation(config.ProgramPath, program, contract), RepositoryWrites: 0,
		LocalTestExecutions: 0, Improvement: "UNKNOWN",
	}
	report.HumanReport = humanReadOnlyReport(report)
	report.ReportDigest = sealReadOnlyProjection(report)
	return report, nil
}

func validateReadOnlyLineageInputs(source sourceRunInput, lineage readOnlyLineageObservation, observation publicworkflowlineage.ReadOnlyObservationEvaluation, policy publicworkflowlineage.Policy) error {
	if lineage.PolicyDigest != policy.SourceDigest {
		return fmt.Errorf("strict workflow-lineage policy digest does not match the current Gooo source")
	}
	if lineage.Schema != publicworkflowlineage.ReportSchema || lineage.Decision != publicworkflowlineage.DecisionRefuted || lineage.LineageState != publicworkflowlineage.StateMismatch || lineage.Evaluation.Decision != lineage.Decision || lineage.Evaluation.LineageState != lineage.LineageState || !lineage.Evaluation.ProductFailureKept || lineage.Evaluation.MismatchDetected || lineage.Evaluation.FallbackAttempted || lineage.Evaluation.FallbackRejected {
		return fmt.Errorf("strict workflow-lineage source failure observation is not eligible")
	}
	if source.Status != "completed" || source.Conclusion == "" || source.Conclusion == "success" {
		return fmt.Errorf("read-only observation requires a completed non-success source run")
	}
	if !sourceMatchesLineage(source, lineage.Source) {
		return fmt.Errorf("read-only observation source identity is not exact")
	}
	lineageInput := publicworkflowlineage.Input{
		Trigger: lineage.Trigger, Source: lineage.Source,
		ExpectedArtifactName: fmt.Sprintf("ci-evidence-%d-%d", lineage.Source.ID, lineage.Source.RunAttempt),
		ExpectedRepository:   policy.Repository, ExpectedWorkflow: policy.SourceWorkflow,
		ExpectedSourceAPIKey: policy.SourceAPIKey, ExpectedArtifactSubjectBinding: policy.ArtifactSubjectBinding,
	}
	expectedEvaluation := publicworkflowlineage.Evaluate(lineageInput)
	expectedObservation := policy.EvaluateReadOnlyObservation(lineageInput)
	if !reflect.DeepEqual(lineage.Evaluation, expectedEvaluation) || !reflect.DeepEqual(observation, expectedObservation) {
		return fmt.Errorf("supplied workflow-lineage decisions do not match the current Gooo policy evaluation")
	}
	if observation.Schema != publicworkflowlineage.ObservationSchema || observation.Eligibility != publicworkflowlineage.ObservationAllowed || observation.Decision != lineage.Decision || observation.LineageState != lineage.LineageState || !observation.ExactSourceIdentity || !observation.SourceFailureKept || !observation.TimingObservationEligible || !observation.OperationObservationEligible || observation.EvidenceReuseAllowed || observation.PromotionAllowed {
		return fmt.Errorf("read-only workflow-lineage observation is not eligible")
	}
	return nil
}

func sourceMatchesLineage(source sourceRunInput, lineage publicworkflowlineage.SourceRun) bool {
	return source.ID == lineage.ID && source.Name == lineage.Name && source.WorkflowName == lineage.Workflow && source.Event == lineage.Event && source.Ref == lineage.Ref && source.HeadBranch == lineage.HeadBranch && source.HeadSHA == lineage.HeadSHA && source.HeadRepository.FullName == lineage.Repository && source.Status == lineage.Status && source.Conclusion == lineage.Conclusion && source.RunAttempt == lineage.RunAttempt
}

func readOnlyWindowForSource(source sourceRunInput) WorkflowWindow {
	window := WorkflowWindow{OperationID: sourceRunOperationID(source), RunID: source.ID, Provider: githubActionsProvider, ClockDomain: githubActionsRunClockDomain, StartAt: firstNonEmpty(source.RunStartedAt, source.CreatedAt), EndAt: source.UpdatedAt, TimestampResolutionMS: 1000, IntervalModel: runtimeIntervalModel, IntervalModelDigest: runtimeIntervalModelDigest()}
	if window.StartAt != "" && window.EndAt != "" {
		duration := observeOperationInterval(operationInterval{OperationID: window.OperationID, RunID: window.RunID, Provider: window.Provider, ClockDomain: window.ClockDomain, StartedAt: window.StartAt, CompletedAt: window.EndAt})
		window.WallMS = duration.wall
		if duration.rejection != "" {
			window.RuntimeRejectionCount = 1
			window.RuntimeRejectionReasons = []string{duration.rejection}
		}
	}
	return window
}

func summarizeReadOnlyTiming(window WorkflowWindow, jobs []JobObservation, jobsUnavailable bool) readOnlyTimingSummary {
	summary := readOnlyTimingSummary{WindowWallMS: window.WallMS, ObservedJobIntervals: window.JobIntervalCount, ObservedStepIntervals: window.StepIntervalCount, JobsDataUnavailable: jobsUnavailable, WindowTimestampsUnknown: window.StartAt == "" || window.EndAt == "", BelowSourceResolutionJobs: window.BelowSourceResolutionJobs, BelowSourceResolutionSteps: window.BelowSourceResolutionSteps, RuntimeRejectionCount: window.RuntimeRejectionCount}
	for _, job := range jobs {
		if !job.Skipped && job.Unknown != nil && job.Unknown.Reason == "JOB_TIMESTAMP_MISSING" {
			summary.MissingJobIntervals++
		}
		for _, step := range job.Steps {
			if !step.Skipped && step.Unknown != nil && step.Unknown.Reason == "STEP_TIMESTAMP_MISSING" {
				summary.MissingStepIntervals++
			}
		}
	}
	return summary
}

func readOnlyCounts(accounting Accounting) readOnlyOperationCounts {
	return readOnlyOperationCounts{Manifest: accounting.ManifestOperations, Observed: accounting.Executed, Skipped: accounting.Skipped, Missing: accounting.Unknown, Rejected: accounting.Rejected}
}

func readOnlyAccounting(operations []OperationObservation) Accounting {
	accounting := Accounting{ManifestOperations: len(operations)}
	for _, operation := range operations {
		switch operation.State {
		case "EXECUTED":
			accounting.Executed++
			if operation.Kind == "TEST" {
				accounting.ExecutedTests++
			} else {
				accounting.ExecutedCommands++
			}
		case "SKIPPED":
			accounting.Skipped++
			if operation.Kind == "TEST" {
				accounting.SkippedTests++
			} else {
				accounting.SkippedCommands++
			}
		case "REJECTED":
			accounting.Rejected++
		default:
			accounting.Unknown++
		}
	}
	return accounting
}

func validateReadOnlyProjection(report readOnlyProjection, manifest Manifest, contract Contract, program, workflow, manifestBytes []byte) error {
	if err := validateStaticInputs(manifest, contract, program); err != nil {
		return err
	}
	if report.Schema != readOnlyProjectionSchema || report.Decision != publicworkflowlineage.DecisionRefuted || report.Resolution != "READ_ONLY" || report.LineageState != publicworkflowlineage.StateMismatch || report.SourceWorkflow != manifest.Workflow || report.WorkflowSourcePath != manifest.WorkflowSource || report.ContractID != contract.ID || !validSHA(report.HeadSHA) || report.SourceRunID <= 0 || report.SourceRunAttempt <= 0 || report.SourceRunStatus != "completed" || report.SourceRunConclusion == "" || report.SourceRunConclusion == "success" || !report.ExactSourceIdentity || !report.SourceFailureKept || report.EvidenceReuseAllowed || report.PromotionAllowed || report.Improvement != "UNKNOWN" || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 {
		return fmt.Errorf("read-only report identity or authority boundary is invalid")
	}
	if !validDigest(report.OperationManifestDigest) || report.OperationManifestDigest != digestBytes(manifestBytes) {
		return fmt.Errorf("read-only operation manifest digest is invalid")
	}
	if len(workflow) == 0 {
		if report.WorkflowSourceDigest != "" {
			return fmt.Errorf("missing workflow source has a digest")
		}
	} else if report.WorkflowSourceDigest != digestBytes(workflow) {
		return fmt.Errorf("read-only workflow source digest is not bound")
	}
	if report.Window.TimestampResolutionMS != 1000 || report.Window.IntervalModel != runtimeIntervalModel || report.Window.IntervalModelDigest != runtimeIntervalModelDigest() || report.Window.WallMS < 0 || report.Window.JobWallMSSum < 0 || report.Window.StepWallMSSum < 0 || report.Window.JobWallMSNominal < 0 || report.Window.StepWallMSNominal < 0 || report.Window.JobIntervalCount < 0 || report.Window.StepIntervalCount < 0 || report.Window.BelowSourceResolutionJobs < 0 || report.Window.BelowSourceResolutionSteps < 0 || report.Window.RuntimeRejectionCount < 0 {
		return fmt.Errorf("read-only runtime model is invalid")
	}
	jobIntervals, stepIntervals := runtimeIntervalCounts(report.Jobs)
	if report.Window.JobIntervalCount != jobIntervals || report.Window.StepIntervalCount != stepIntervals || report.Window.JobWallMSNominal != runtimeNominalJob(report.Jobs) || report.Window.StepWallMSNominal != runtimeNominalStep(report.Jobs) || report.Window.JobWallMSSum != report.Window.JobWallMSNominal || report.Window.StepWallMSSum != report.Window.StepWallMSNominal {
		return fmt.Errorf("read-only runtime accounting is inconsistent")
	}
	if count, reasons := runtimeRejectionEvidence(report.Window, report.Jobs); report.Window.RuntimeRejectionCount != count || !sameStrings(report.Window.RuntimeRejectionReasons, reasons) {
		return fmt.Errorf("read-only runtime rejection evidence is inconsistent")
	}
	if len(report.Operations) != len(manifest.Operations) || report.Accounting.ManifestOperations != len(manifest.Operations) || report.OperationCounts.Manifest != len(manifest.Operations) || report.OperationCounts.Observed != report.Accounting.Executed || report.OperationCounts.Skipped != report.Accounting.Skipped || report.OperationCounts.Missing != report.Accounting.Unknown || report.OperationCounts.Rejected != report.Accounting.Rejected || report.OperationCounts.Observed+report.OperationCounts.Skipped+report.OperationCounts.Missing+report.OperationCounts.Rejected != report.OperationCounts.Manifest || report.Accounting.Executed+report.Accounting.Skipped+report.Accounting.Unknown+report.Accounting.Rejected != report.Accounting.ManifestOperations {
		return fmt.Errorf("read-only operation accounting is inconsistent")
	}
	if derived := readOnlyAccounting(report.Operations); derived != report.Accounting {
		return fmt.Errorf("read-only operation states do not match accounting")
	}
	if report.Timing.WindowWallMS != report.Window.WallMS || report.Timing.ObservedJobIntervals != report.Window.JobIntervalCount || report.Timing.ObservedStepIntervals != report.Window.StepIntervalCount || report.Timing.WindowTimestampsUnknown != (report.Window.StartAt == "" || report.Window.EndAt == "") || report.Timing.BelowSourceResolutionJobs != report.Window.BelowSourceResolutionJobs || report.Timing.BelowSourceResolutionSteps != report.Window.BelowSourceResolutionSteps || report.Timing.RuntimeRejectionCount != report.Window.RuntimeRejectionCount {
		return fmt.Errorf("read-only timing summary is inconsistent")
	}
	if expected := summarizeReadOnlyTiming(report.Window, report.Jobs, len(report.Jobs) == 0); expected != report.Timing {
		return fmt.Errorf("read-only timing details do not match observations")
	}
	if err := validateJobs(report.Jobs, report.HeadSHA); err != nil {
		return err
	}
	if err := validateOperations(report.Operations, manifest.Operations, manifest.WorkflowSource, workflow, report.SourceEvent); err != nil {
		return err
	}
	if len(report.Graph.Activities) != len(contract.Cells) || report.Graph.ActivityCount != len(contract.Cells) || report.Graph.BindingCount != len(contract.Cells) || report.Graph.ProgramDigest != digestBytes(program) {
		return fmt.Errorf("read-only meta graph denominator is invalid")
	}
	for index, activity := range report.Graph.Activities {
		if activity != contract.Cells[index].Activity {
			return fmt.Errorf("read-only meta graph activity order is invalid")
		}
	}
	for _, job := range report.Jobs {
		if err := validateReadOnlyDuration(job.WallMS, job.Unknown != nil, job.Skipped, job.BelowSourceResolution, job.RejectionReason); err != nil {
			return fmt.Errorf("job %d: %w", job.ID, err)
		}
		for _, step := range job.Steps {
			if err := validateReadOnlyDuration(step.WallMS, step.Unknown != nil, step.Skipped, step.BelowSourceResolution, step.RejectionReason); err != nil {
				return fmt.Errorf("step %q: %w", step.Name, err)
			}
		}
	}
	for _, operation := range report.Operations {
		if err := validateReadOnlyDuration(operation.WallMS, operation.Unknown != nil, operation.Skipped, operation.BelowSourceResolution, operation.RejectionReason); err != nil {
			return fmt.Errorf("operation %s: %w", operation.ID, err)
		}
	}
	if report.ReportDigest != sealReadOnlyProjection(report) {
		return fmt.Errorf("read-only report digest is not sealed")
	}
	return nil
}

func runtimeNominalJob(jobs []JobObservation) int64 {
	job, _ := runtimeNominalForJobs(jobs)
	return job
}

func runtimeNominalStep(jobs []JobObservation) int64 {
	_, step := runtimeNominalForJobs(jobs)
	return step
}

func validateReadOnlyDuration(wall int64, unknown, skipped, below bool, rejection string) error {
	if wall < 0 || (unknown || skipped || rejection != "") && wall != 0 || below && wall != 0 {
		return fmt.Errorf("runtime duration is not an integer nonnegative observation")
	}
	return nil
}

func validateConfigReadOnlyProjection(config Config, report readOnlyProjection) error {
	var manifest Manifest
	manifestBytes, err := readJSON(config.ManifestPath, &manifest)
	if err != nil {
		return err
	}
	var contract Contract
	if _, err := readJSON(config.ContractPath, &contract); err != nil {
		return err
	}
	program, err := os.ReadFile(config.ProgramPath)
	if err != nil {
		return err
	}
	workflow, _ := os.ReadFile(manifest.WorkflowSource)
	if report.OperationManifestDigest != digestBytes(manifestBytes) {
		return fmt.Errorf("read-only operation manifest input changed during validation")
	}
	return validateReadOnlyProjection(report, manifest, contract, program, workflow, manifestBytes)
}

func writeReadOnlyProjection(config Config, report readOnlyProjection) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if config.OutputPath == "" {
		if _, err := os.Stdout.Write(data); err != nil {
			return err
		}
	} else if err := writeOwned(config.OutputPath, data); err != nil {
		return err
	}
	if config.MarkdownPath != "" {
		if err := writeOwned(config.MarkdownPath, []byte(report.HumanReport)); err != nil {
			return err
		}
	}
	return nil
}

func sealReadOnlyProjection(report readOnlyProjection) string {
	report.ReportDigest = ""
	data, _ := json.Marshal(report)
	return digestBytes(data)
}

func humanReadOnlyReport(report readOnlyProjection) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "CI effort read-only observation: %s / %s (%s)\n", report.Decision, report.Resolution, report.Reason)
	fmt.Fprintf(&builder, "lineage=%s state=%s reason=%s\n", report.Decision, report.LineageState, report.LineageReason)
	fmt.Fprintf(&builder, "source repository=%s workflow=%s event=%s ref=%q head=%s run=%d attempt=%d status=%s conclusion=%s exact_identity=%t source_failure_kept=%t\n", report.Repository, report.SourceWorkflow, report.SourceEvent, report.SourceRef, report.HeadSHA, report.SourceRunID, report.SourceRunAttempt, report.SourceRunStatus, report.SourceRunConclusion, report.ExactSourceIdentity, report.SourceFailureKept)
	fmt.Fprintf(&builder, "workflow_source=%s digest=%s contract=%s manifest_digest=%s\n", report.WorkflowSourcePath, report.WorkflowSourceDigest, report.ContractID, report.OperationManifestDigest)
	fmt.Fprintf(&builder, "timing window_wall_ms=%d window_timestamps_unknown=%t observed_jobs=%d observed_steps=%d missing_jobs=%d missing_steps=%d jobs_data_unavailable=%t below_source_resolution_jobs=%d steps=%d runtime_rejections=%d\n", report.Timing.WindowWallMS, report.Timing.WindowTimestampsUnknown, report.Timing.ObservedJobIntervals, report.Timing.ObservedStepIntervals, report.Timing.MissingJobIntervals, report.Timing.MissingStepIntervals, report.Timing.JobsDataUnavailable, report.Timing.BelowSourceResolutionJobs, report.Timing.BelowSourceResolutionSteps, report.Timing.RuntimeRejectionCount)
	fmt.Fprintf(&builder, "operations manifest=%d observed=%d skipped=%d missing=%d rejected=%d\n", report.OperationCounts.Manifest, report.OperationCounts.Observed, report.OperationCounts.Skipped, report.OperationCounts.Missing, report.OperationCounts.Rejected)
	for _, operation := range report.Operations {
		unknownReason := ""
		if operation.Unknown != nil {
			unknownReason = operation.Unknown.Reason
		}
		fmt.Fprintf(&builder, "operation %s state=%s wall_ms=%d below_source_resolution=%t unknown=%q rejection=%q\n", operation.ID, operation.State, operation.WallMS, operation.BelowSourceResolution, unknownReason, operation.RejectionReason)
	}
	fmt.Fprintf(&builder, "evidence_reuse_allowed=%t promotion_allowed=%t repository_writes=%d local_test_executions=%d improvement=%s\n", report.EvidenceReuseAllowed, report.PromotionAllowed, report.RepositoryWrites, report.LocalTestExecutions, report.Improvement)
	fmt.Fprintf(&builder, "read-only observation records source timing and operation state only; it does not establish savings or a comparable pair\n")
	return builder.String()
}
