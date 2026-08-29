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
	observedJobs, window, err := observeJobs(jobs.Jobs)
	if err != nil {
		return Report{}, err
	}
	operations, accounting := observeOperations(manifest.Operations, jobs.Jobs)
	openTofu, err := observeOpenTofu(config.OpenTofuPath, config.OpenTofuMetaPath, source.HeadSHA)
	if err != nil {
		return Report{}, err
	}
	key, err := buildReuseKey(config, source, manifestBytes, contractBytes, sourceBytes, jobsBytes, openTofu)
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
		SourceRunURL: source.HTMLURL, Window: window, Jobs: observedJobs, Operations: operations,
		Accounting: accounting, Reuse: reuse, OpenTofu: openTofu,
		OperationManifestDigest: digestBytes(manifestBytes),
		Graph: graph, RepositoryStatus: repositoryStatus, RepositoryWrites: repositoryStatus.Writes, LocalTestExecutions: 0,
		CrossProjectRequiredGates: 0, Improvement: "UNKNOWN",
		Counterexamples: fixedCounterexamples(),
	}
	report.Cells = buildCells(contract, report)
	report.Decision, report.Resolution, report.Reason = classifyReport(report)
	report.HumanReport = humanReport(report)
	report.ReportDigest = sealReport(report)
	return report, nil
}

func observeJobs(input []APIJob) ([]JobObservation, WorkflowWindow, error) {
	if len(input) == 0 {
		return nil, WorkflowWindow{}, fmt.Errorf("source run has no jobs")
	}
	sorted := append([]APIJob(nil), input...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	result := make([]JobObservation, 0, len(sorted))
	var window WorkflowWindow
	for index, job := range sorted {
		wall, err := durationMS(job.StartedAt, job.CompletedAt)
		if err != nil {
			return nil, WorkflowWindow{}, fmt.Errorf("job %d: %w", job.ID, err)
		}
		steps, stepWall, err := observeSteps(job.Steps)
		if err != nil {
			return nil, WorkflowWindow{}, fmt.Errorf("job %d: %w", job.ID, err)
		}
		result = append(result, JobObservation{ID: job.ID, Name: job.Name, Status: job.Status,
			Conclusion: job.Conclusion, HeadSHA: job.HeadSHA, StartedAt: job.StartedAt,
			CompletedAt: job.CompletedAt, WallMS: wall, Steps: steps})
		if index == 0 || job.StartedAt < window.StartAt {
			window.StartAt = job.StartedAt
		}
		if index == 0 || job.CompletedAt > window.EndAt {
			window.EndAt = job.CompletedAt
		}
		window.JobWallMSSum += wall
		window.StepWallMSSum += stepWall
	}
	window.WallMS, _ = durationMS(window.StartAt, window.EndAt)
	return result, window, nil
}

func observeSteps(input []APIStep) ([]StepObservation, int64, error) {
	result := make([]StepObservation, 0, len(input))
	var total int64
	for _, step := range input {
		wall := int64(0)
		if step.StartedAt != "" || step.CompletedAt != "" {
			var err error
			wall, err = durationMS(step.StartedAt, step.CompletedAt)
			if err != nil {
				return nil, 0, fmt.Errorf("step %q: %w", step.Name, err)
			}
			total += wall
		}
		result = append(result, StepObservation{Name: step.Name, Status: step.Status,
			Conclusion: step.Conclusion, StartedAt: step.StartedAt, CompletedAt: step.CompletedAt,
			WallMS: wall})
	}
	return result, total, nil
}

func durationMS(start, end string) (int64, error) {
	if start == "" || end == "" {
		return 0, fmt.Errorf("missing timestamp")
	}
	begin, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		return 0, fmt.Errorf("invalid start timestamp: %w", err)
	}
	finish, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		return 0, fmt.Errorf("invalid end timestamp: %w", err)
	}
	delta := finish.Sub(begin)
	if delta <= 0 {
		return 0, fmt.Errorf("non-positive duration")
	}
	value := int64(delta / time.Millisecond)
	if delta%time.Millisecond != 0 {
		value++
	}
	if value == 0 {
		value = 1
	}
	return value, nil
}

func observeOperations(specs []OperationSpec, jobs []APIJob) ([]OperationObservation, Accounting) {
	result := make([]OperationObservation, 0, len(specs))
	accounting := Accounting{ManifestOperations: len(specs)}
	for _, spec := range specs {
		result = append(result, observeOperation(spec, jobs))
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

func observeOperation(spec OperationSpec, jobs []APIJob) OperationObservation {
	base := OperationObservation{ID: spec.ID, Kind: spec.Kind, JobName: spec.JobName,
		StepName: spec.StepName, Command: append([]string(nil), spec.Command...), State: "UNKNOWN"}
	matchingJobs := make([]APIJob, 0, 1)
	for _, job := range jobs {
		if job.Name == spec.JobName {
			matchingJobs = append(matchingJobs, job)
		}
	}
	if len(matchingJobs) != 1 {
		if len(matchingJobs) > 1 {
			base.State = "REJECTED"
		}
		base.EvidenceDigest = operationDigest(base)
		return base
	}
	job := matchingJobs[0]
	base.JobID, base.JobConclusion = job.ID, job.Conclusion
	matches := make([]APIStep, 0, 1)
	for _, step := range job.Steps {
		if step.Name == spec.StepName {
			matches = append(matches, step)
		}
	}
	if len(matches) != 1 {
		if len(matches) > 1 {
			base.State = "REJECTED"
		}
		base.EvidenceDigest = operationDigest(base)
		return base
	}
	step := matches[0]
	base.StepStatus, base.StepConclusion = step.Status, step.Conclusion
	base.StartedAt, base.CompletedAt = step.StartedAt, step.CompletedAt
	if step.Conclusion == "skipped" {
		base.State = "SKIPPED"
	} else if step.Status == "completed" && step.Conclusion != "" && step.StartedAt != "" && step.CompletedAt != "" {
		base.State = "EXECUTED"
		base.WallMS, _ = durationMS(step.StartedAt, step.CompletedAt)
	}
	base.EvidenceDigest = operationDigest(base)
	return base
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
