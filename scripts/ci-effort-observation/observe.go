package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const runtimeIntervalModel = "github-rest-iso8601-second-value-v1"

const runtimeIntervalModelDefinition = "timestamp values=UTC ISO-8601 second values; wall_ms=reported endpoint delta; physical elapsed bounds=not established; parallel sums are not critical-path claims"

type sourceRunInput struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Event          string `json:"event"`
	Ref            string `json:"ref"`
	HeadBranch     string `json:"head_branch"`
	HeadSHA        string `json:"head_sha"`
	HeadRepository struct {
		FullName string `json:"full_name"`
	} `json:"head_repository"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	RunAttempt   int64  `json:"run_attempt"`
	CreatedAt    string `json:"created_at"`
	RunStartedAt string `json:"run_started_at"`
	UpdatedAt    string `json:"updated_at"`
	HTMLURL      string `json:"html_url"`
}

func readJSON(path string, value any) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("required input path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, value); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return data, nil
}

func buildReport(config Config) (Report, error) {
	var source sourceRunInput
	sourceBytes, err := readJSON(config.RunPath, &source)
	if err != nil {
		return Report{}, err
	}
	var jobs JobsInput
	jobsBytes, err := readJSON(config.JobsPath, &jobs)
	if err != nil {
		return Report{}, err
	}
	var manifest Manifest
	manifestBytes, err := readJSON(config.ManifestPath, &manifest)
	if err != nil {
		return Report{}, err
	}
	var contract Contract
	contractBytes, err := readJSON(config.ContractPath, &contract)
	if err != nil {
		return Report{}, err
	}
	programBytes, err := os.ReadFile(config.ProgramPath)
	if err != nil {
		return Report{}, err
	}
	workflowBytes, workflowErr := os.ReadFile(manifest.WorkflowSource)
	if err := validateStaticInputs(manifest, contract, programBytes); err != nil {
		return Report{}, err
	}
	if source.HeadSHA == "" || source.HeadRepository.FullName == "" {
		return Report{}, fmt.Errorf("source run identity is incomplete")
	}
	var repositoryStatus RepositoryStatus
	if _, err := readJSON(config.RepositoryStatusPath, &repositoryStatus); err != nil {
		return Report{}, err
	}
	if repositoryStatus.Writes < 0 {
		return Report{}, fmt.Errorf("repository status writes is negative")
	}
	timeCausality, err := loadTimeCausality(config.TimeCausalityRoot)
	if err != nil {
		return Report{}, err
	}
	observedJobs, window, err := observeJobsWithSource(jobs.Jobs, source)
	if err != nil {
		return Report{}, err
	}
	operations, accounting := observeOperations(manifest.Operations, jobs.Jobs, manifest.WorkflowSource, workflowBytes, workflowErr, source.Event)
	openTofu, err := observeOpenTofu(config.OpenTofuPath, config.OpenTofuMetaPath, source.HeadSHA)
	if err != nil {
		return Report{}, err
	}
	key, err := buildReuseKey(config, source, manifestBytes, contractBytes, sourceBytes, jobsBytes, openTofu, timeCausality, workflowBytes, operations)
	if err != nil {
		return Report{}, err
	}
	reuse, err := buildReuse(config.PriorPath, key)
	if err != nil {
		return Report{}, err
	}
	graph := graphObservation(config.ProgramPath, programBytes, contract)
	report := Report{
		Schema: reportSchema, ContractID: contract.ID, Repository: source.HeadRepository.FullName,
		SourceWorkflow: manifest.Workflow, SourceEvent: source.Event, SourceRef: source.Ref,
		HeadSHA: source.HeadSHA, SourceRunConclusion: source.Conclusion, SourceRunID: source.ID, SourceRunAttempt: source.RunAttempt,
		SourceRunURL: source.HTMLURL, WorkflowSourcePath: manifest.WorkflowSource, WorkflowSourceDigest: digestIfPresent(workflowBytes),
		Window: window, RuntimeResolution: runtimeResolution(window), Jobs: observedJobs, Operations: operations,
		Accounting: accounting, Reuse: reuse, OpenTofu: openTofu,
		OperationManifestDigest: digestBytes(manifestBytes),
		Graph:                   graph, TimeCausality: timeCausality, RepositoryStatus: repositoryStatus, RepositoryWrites: repositoryStatus.Writes, LocalTestExecutions: 0,
		CrossProjectRequiredGates: 0, Improvement: "UNKNOWN",
		Counterexamples: fixedCounterexamples(),
		RuntimeCases:    runtimeCases(),
		UnknownEvidence: firstUnknownEvidence(observedJobs, operations, openTofu)}
	report.Cells = buildCells(contract, report)
	report.Decision, report.Resolution, report.Reason = classifyReport(report)
	report.HumanReport = humanReport(report)
	report.ReportDigest = sealReport(report)
	return report, nil
}

func observeJobs(input []APIJob) ([]JobObservation, WorkflowWindow, error) {
	return observeJobsWithSource(input, sourceRunInput{})
}

func observeJobsWithSource(input []APIJob, source sourceRunInput) ([]JobObservation, WorkflowWindow, error) {
	if len(input) == 0 {
		return nil, WorkflowWindow{}, fmt.Errorf("source run has no jobs")
	}
	sorted := append([]APIJob(nil), input...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	result := make([]JobObservation, 0, len(sorted))
	window := WorkflowWindow{OperationID: sourceRunOperationID(source), RunID: source.ID, Provider: githubActionsProvider, ClockDomain: githubActionsRunClockDomain}
	window.StartAt = firstNonEmpty(source.RunStartedAt, source.CreatedAt)
	window.EndAt = source.UpdatedAt
	for _, job := range sorted {
		skipped := job.Status == "skipped" || job.Conclusion == "skipped"
		duration := timestampObservation{}
		if !skipped {
			duration = observeOperationInterval(operationIntervalForJob(job))
		}
		var unknown *Unknown
		if !skipped && duration.missing {
			unknown = jobRuntimeUnknown()
		}
		steps, stepWall, belowSteps, stepRejections, stepReasons, err := observeStepsWithResolutionForJob(job.Steps, job)
		if err != nil {
			return nil, WorkflowWindow{}, fmt.Errorf("job %d: %w", job.ID, err)
		}
		window.StepIntervalCount += observedStepIntervalCount(steps)
		result = append(result, JobObservation{ID: job.ID, OperationID: jobOperationID(job), RunID: job.RunID, Provider: githubActionsProvider, ClockDomain: githubActionsJobClockDomain,
			Name: job.Name, Status: job.Status,
			Conclusion: job.Conclusion, HeadSHA: job.HeadSHA, StartedAt: job.StartedAt,
			CompletedAt: job.CompletedAt, WallMS: duration.wall, BelowSourceResolution: duration.below,
			RejectionReason: duration.rejection, Skipped: skipped, Steps: steps, Unknown: unknown})
		if source.ID == 0 && !skipped && job.StartedAt != "" && (window.StartAt == "" || job.StartedAt < window.StartAt) {
			window.StartAt = job.StartedAt
		}
		if source.ID == 0 && !skipped && job.CompletedAt != "" && (window.EndAt == "" || job.CompletedAt > window.EndAt) {
			window.EndAt = job.CompletedAt
		}
		window.JobWallMSSum += duration.wall
		window.StepWallMSSum += stepWall
		if !skipped && !duration.missing && duration.rejection == "" {
			window.JobIntervalCount++
		}
		if duration.below {
			window.BelowSourceResolutionJobs++
		}
		window.BelowSourceResolutionSteps += belowSteps
		if duration.rejection != "" {
			window.RuntimeRejectionCount++
			window.RuntimeRejectionReasons = append(window.RuntimeRejectionReasons, duration.rejection)
		}
		window.RuntimeRejectionCount += stepRejections
		window.RuntimeRejectionReasons = append(window.RuntimeRejectionReasons, stepReasons...)
	}
	if source.ID > 0 && window.StartAt != "" && window.EndAt != "" {
		windowDuration := observeOperationInterval(operationInterval{OperationID: window.OperationID, RunID: source.ID, Provider: githubActionsProvider, ClockDomain: githubActionsRunClockDomain, StartedAt: window.StartAt, CompletedAt: window.EndAt})
		window.WallMS = windowDuration.wall
		if windowDuration.rejection != "" {
			window.RuntimeRejectionCount++
			window.RuntimeRejectionReasons = append(window.RuntimeRejectionReasons, windowDuration.rejection)
		}
	} else if len(result) == 1 && !result[0].Skipped {
		window.OperationID = result[0].OperationID
		window.RunID = result[0].RunID
		window.Provider = result[0].Provider
		window.ClockDomain = result[0].ClockDomain
		window.WallMS = observeOperationInterval(operationInterval{OperationID: result[0].OperationID, RunID: result[0].RunID, JobID: result[0].ID, Provider: result[0].Provider, ClockDomain: result[0].ClockDomain, StartedAt: result[0].StartedAt, CompletedAt: result[0].CompletedAt}).wall
	}
	window.TimestampResolutionMS = 1000
	window.IntervalModel = runtimeIntervalModel
	window.IntervalModelDigest = runtimeIntervalModelDigest()
	window.JobWallMSNominal, window.StepWallMSNominal = runtimeNominalForJobs(result)
	sort.Strings(window.RuntimeRejectionReasons)
	return result, window, nil
}

func observeSteps(input []APIStep) ([]StepObservation, int64, error) {
	steps, total, _, _, _, err := observeStepsWithResolution(input)
	return steps, total, err
}

func observedStepIntervalCount(steps []StepObservation) int {
	count := 0
	for _, step := range steps {
		if !step.Skipped && step.Status != "skipped" && step.Conclusion != "skipped" && step.Unknown == nil && step.RejectionReason == "" && step.StartedAt != "" && step.CompletedAt != "" {
			count++
		}
	}
	return count
}

func runtimeIntervalModelDigest() string {
	return digestString(runtimeIntervalModel + ":" + runtimeIntervalModelDefinition)
}

func runtimeNominalForJobs(jobs []JobObservation) (int64, int64) {
	var jobNominal, stepNominal int64
	for _, job := range jobs {
		if job.Skipped {
			continue
		}
		duration := observeOperationInterval(operationInterval{OperationID: job.OperationID, RunID: job.RunID, Provider: job.Provider, ClockDomain: job.ClockDomain, StartedAt: job.StartedAt, CompletedAt: job.CompletedAt})
		if !duration.missing && duration.rejection == "" {
			jobNominal += duration.wall
		}
		for _, step := range job.Steps {
			if step.Skipped || step.Conclusion == "skipped" || step.Unknown != nil || step.RejectionReason != "" {
				continue
			}
			duration := observeOperationInterval(operationInterval{OperationID: step.OperationID, RunID: step.RunID, Provider: step.Provider, ClockDomain: step.ClockDomain, StartedAt: step.StartedAt, CompletedAt: step.CompletedAt})
			if !duration.missing && duration.rejection == "" {
				stepNominal += duration.wall
			}
		}
	}
	return jobNominal, stepNominal
}

func observeStepsWithResolution(input []APIStep) ([]StepObservation, int64, int, int, []string, error) {
	return observeStepsWithResolutionForJob(input, APIJob{ID: 1, RunID: 1})
}

func observeStepsWithResolutionForJob(input []APIStep, job APIJob) ([]StepObservation, int64, int, int, []string, error) {
	result := make([]StepObservation, 0, len(input))
	var total int64
	belowCount := 0
	rejectionCount := 0
	var rejectionReasons []string
	for _, step := range input {
		if isCleanupStep(step.Name) {
			continue
		}
		skipped := step.Status == "skipped" || step.Conclusion == "skipped"
		duration := timestampObservation{}
		if !skipped {
			duration = observeOperationInterval(operationIntervalForStep(job, step))
		}
		var unknown *Unknown
		if skipped {
			// A skipped step has no execution interval to observe.
		} else if duration.missing {
			unknown = stepRuntimeUnknown()
		} else {
			total += duration.wall
			if duration.below {
				belowCount++
			}
			if duration.rejection != "" {
				rejectionCount++
				rejectionReasons = append(rejectionReasons, duration.rejection)
			}
		}
		result = append(result, StepObservation{OperationID: stepOperationID(job, step.Name), RunID: job.RunID, Provider: githubActionsProvider, ClockDomain: githubActionsJobClockDomain,
			Name: step.Name, Status: step.Status,
			Conclusion: step.Conclusion, StartedAt: step.StartedAt, CompletedAt: step.CompletedAt,
			WallMS: duration.wall, BelowSourceResolution: duration.below,
			RejectionReason: duration.rejection, Skipped: skipped, Unknown: unknown})
	}
	return result, total, belowCount, rejectionCount, rejectionReasons, nil
}

func isCleanupStep(name string) bool {
	return strings.HasPrefix(strings.TrimSpace(name), "Post Run ")
}

func durationMS(start, end string) (int64, error) {
	duration := observeTimestamp(start, end)
	if duration.missing {
		return 0, fmt.Errorf("missing timestamp")
	}
	if duration.rejection != "" {
		return 0, fmt.Errorf("%s", duration.rejection)
	}
	if duration.below {
		return 0, nil
	}
	return duration.wall, nil
}

type timestampObservation struct {
	wall      int64
	below     bool
	missing   bool
	rejection string
}

func observeTimestamp(start, end string) timestampObservation {
	if start == "" || end == "" {
		return timestampObservation{missing: true}
	}
	begin, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		return timestampObservation{rejection: "OPERATION_TIMESTAMP_MALFORMED"}
	}
	finish, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		return timestampObservation{rejection: "OPERATION_TIMESTAMP_MALFORMED"}
	}
	if finish.Before(begin) {
		return timestampObservation{rejection: "OPERATION_DURATION_NEGATIVE"}
	}
	if finish.Equal(begin) {
		return timestampObservation{below: true}
	}
	delta := finish.Sub(begin)
	wall := int64(delta / time.Millisecond)
	if delta%time.Millisecond != 0 {
		wall++
	}
	if wall == 0 {
		wall = 1
	}
	return timestampObservation{wall: wall}
}

func runtimeResolution(window WorkflowWindow) string {
	if window.TimestampResolutionMS > 0 {
		return "LOWER/SOURCE_SECOND"
	}
	return "EXACT"
}

func observeOperations(specs []OperationSpec, jobs []APIJob, workflowPath string, workflow []byte, workflowErr error, sourceEvent string) ([]OperationObservation, Accounting) {
	result := make([]OperationObservation, 0, len(specs))
	accounting := Accounting{ManifestOperations: len(specs)}
	for _, spec := range specs {
		result = append(result, observeOperation(spec, jobs, workflowPath, workflow, workflowErr, sourceEvent))
		switch result[len(result)-1].State {
		case "EXECUTED":
			accounting.Executed++
			if result[len(result)-1].Kind == "TEST" {
				accounting.ExecutedTests++
			} else {
				accounting.ExecutedCommands++
			}
		case "SKIPPED":
			accounting.Skipped++
			if result[len(result)-1].Kind == "TEST" {
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
	return result, accounting
}

func observeOperation(spec OperationSpec, jobs []APIJob, workflowPath string, workflow []byte, workflowErr error, sourceEvents ...string) OperationObservation {
	sourceEvent := ""
	if len(sourceEvents) > 0 {
		sourceEvent = sourceEvents[0]
	}
	evidenceStep := operationEvidenceStep(spec, sourceEvent)
	base := OperationObservation{ID: spec.ID, Kind: spec.Kind, ProofObligationID: spec.ProofObligationID,
		SourceEvent: sourceEvent, JobName: spec.JobName, DeclaredStepCandidates: declaredStepCandidates(spec),
		GuardStepName: spec.GuardStepName, Command: append([]string(nil), spec.Command...), State: "UNKNOWN"}
	if evidenceStep == "" {
		base.Unknown = operationEventUnknown()
		base.EvidenceDigest = operationDigest(base)
		return base
	}
	if workflowErr != nil || len(workflow) == 0 {
		base.Unknown = commandContextUnknown()
		base.EvidenceDigest = operationDigest(base)
		return base
	}
	contextDigest, err := bindWorkflowCommandForEvent(workflow, workflowPath, spec, sourceEvent)
	if err != nil {
		base.Unknown = commandContextUnknown()
		base.EvidenceDigest = operationDigest(base)
		return base
	}
	base.StepName, base.BoundStepName, base.EvidenceStepName = evidenceStep, evidenceStep, evidenceStep
	base.WorkflowSourcePath, base.WorkflowSourceDigest = workflowPath, digestBytes(workflow)
	base.CommandContextDigest, base.CommandBound = contextDigest, true
	matchingJobs := make([]APIJob, 0, 1)
	for _, job := range jobs {
		if job.Name == spec.JobName {
			matchingJobs = append(matchingJobs, job)
		}
	}
	if len(matchingJobs) != 1 {
		if len(matchingJobs) > 1 {
			base.State, base.RejectionReason = "REJECTED", "DUPLICATE_JOB_OBSERVATION"
		} else {
			base.Unknown = operationEvidenceUnknown("JOB_OBSERVATION_MISSING")
		}
		base.EvidenceDigest = operationDigest(base)
		return base
	}
	job := matchingJobs[0]
	base.JobID, base.JobConclusion = job.ID, job.Conclusion
	base.RunID, base.Provider, base.ClockDomain = job.RunID, githubActionsProvider, githubActionsJobClockDomain
	matches := make([]APIStep, 0, 1)
	for _, step := range job.Steps {
		if step.Name == evidenceStep {
			matches = append(matches, step)
		}
	}
	if len(matches) != 1 {
		if len(matches) > 1 {
			base.State, base.RejectionReason = "REJECTED", "DUPLICATE_STEP_OBSERVATION"
		} else {
			base.Unknown = operationEvidenceUnknown("STEP_OBSERVATION_MISSING")
		}
		base.EvidenceDigest = operationDigest(base)
		return base
	}
	step := matches[0]
	base.OperationID = stepOperationID(job, step.Name)
	base.StepStatus, base.StepConclusion = step.Status, step.Conclusion
	base.StartedAt, base.CompletedAt = step.StartedAt, step.CompletedAt
	if spec.GuardStepName != "" {
		guards := namedSteps(job.Steps, spec.GuardStepName)
		if len(guards) != 1 {
			if len(guards) > 1 {
				base.State, base.RejectionReason = "REJECTED", "DUPLICATE_GUARD_STEP_OBSERVATION"
			} else {
				base.Unknown = operationEvidenceUnknown("GUARD_STEP_OBSERVATION_MISSING")
			}
			base.EvidenceDigest = operationDigest(base)
			return base
		}
		guard := guards[0]
		base.GuardStepStatus, base.GuardStepConclusion = guard.Status, guard.Conclusion
		if guard.Status != "completed" || guard.Conclusion != "success" {
			base.State, base.RejectionReason = "REJECTED", "GUARD_STEP_NOT_SUCCESSFUL"
			base.EvidenceDigest = operationDigest(base)
			return base
		}
		base.GuardBound = true
	}
	if step.Status == "skipped" || step.Conclusion == "skipped" {
		base.State, base.Skipped = "SKIPPED", true
	} else if step.Status == "completed" && step.Conclusion != "" && step.StartedAt != "" && step.CompletedAt != "" {
		duration := observeOperationInterval(operationIntervalForStep(job, step))
		switch {
		case duration.rejection != "":
			base.State, base.RejectionReason = "REJECTED", duration.rejection
		case duration.below:
			base.State, base.BelowSourceResolution = "EXECUTED", true
		default:
			base.State, base.WallMS = "EXECUTED", duration.wall
		}
	} else {
		base.Unknown = operationTimestampMissingUnknown()
	}
	base.EvidenceDigest = operationDigest(base)
	return base
}

func namedSteps(steps []APIStep, name string) []APIStep {
	result := make([]APIStep, 0, 1)
	for _, step := range steps {
		if step.Name == name {
			result = append(result, step)
		}
	}
	return result
}

func commandContextUnknown() *Unknown {
	return &Unknown{Stage: "OBSERVE", Step: "BIND_WORKFLOW_COMMAND", Reason: "COMMAND_CONTEXT_MISSING",
		UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_WORKFLOW_COMMAND_EVIDENCE", BlockedBy: []string{}}
}

func jobRuntimeUnknown() *Unknown {
	return &Unknown{Stage: "OBSERVE", Step: "READ_JOB_RUNTIME", Reason: "JOB_TIMESTAMP_MISSING",
		UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_JOB_TIMESTAMPS", BlockedBy: []string{}}
}

func stepRuntimeUnknown() *Unknown {
	return &Unknown{Stage: "OBSERVE", Step: "READ_STEP_RUNTIME", Reason: "STEP_TIMESTAMP_MISSING",
		UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_STEP_TIMESTAMPS", BlockedBy: []string{}}
}

func operationEvidenceUnknown(reason string) *Unknown {
	return &Unknown{Stage: "OBSERVE", Step: "READ_OPERATION_RECEIPT", Reason: reason,
		UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_OPERATION_RECEIPT", BlockedBy: []string{}}
}

func operationEventUnknown() *Unknown {
	return &Unknown{Stage: "OBSERVE", Step: "BIND_EVENT_OPERATION", Reason: "EVENT_OPERATION_STEP_MISSING",
		UnknownClass: "AMBIGUOUS", NextOperation: "RESTORE_EVENT_OPERATION_BINDING", BlockedBy: []string{}}
}

func operationTimestampMissingUnknown() *Unknown {
	return &Unknown{Stage: "OBSERVE", Step: "READ_OPERATION_RUNTIME", Reason: "OPERATION_TIMESTAMP_MISSING",
		UnknownClass: "DIRECT_MISSING", NextOperation: "RESTORE_OPERATION_TIMESTAMPS", BlockedBy: []string{}}
}

func digestIfPresent(value []byte) string {
	if len(value) == 0 {
		return ""
	}
	return digestBytes(value)
}

func firstUnknownEvidence(jobs []JobObservation, operations []OperationObservation, openTofu ExternalOpenTofu) *Unknown {
	for _, job := range jobs {
		if job.Unknown != nil {
			return job.Unknown
		}
		for _, step := range job.Steps {
			if step.Unknown != nil {
				return step.Unknown
			}
		}
	}
	for _, operation := range operations {
		if operation.Unknown != nil {
			return operation.Unknown
		}
	}
	return openTofu.Unknown
}

func graphObservation(path string, program []byte, contract Contract) GraphObservation {
	activities := make([]string, 0, len(contract.Cells))
	for _, cell := range contract.Cells {
		activities = append(activities, cell.Activity)
	}
	return GraphObservation{ProgramPath: path, ProgramDigest: digestBytes(program),
		ActivityCount: len(activities), BindingCount: len(activities), Activities: activities}
}

func digestBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(hash[:])
}

func digestString(value string) string { return digestBytes([]byte(value)) }

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return digestBytes(data)
}

func digestNamed(parts map[string][]byte) string {
	names := make([]string, 0, len(parts))
	for name := range parts {
		names = append(names, name)
	}
	sort.Strings(names)
	var builder strings.Builder
	for _, name := range names {
		builder.WriteString(name)
		builder.WriteByte(0)
		builder.WriteString(digestBytes(parts[name]))
		builder.WriteByte('\n')
	}
	return digestString(builder.String())
}
