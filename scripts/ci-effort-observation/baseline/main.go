package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"
)

const (
	inputSchema  = "gooo/ci-required-check-baseline-input/v1"
	reportSchema = "gooo/ci-required-check-baseline/v1"
	minSamples   = 5
	maxSamples   = 20
)

var requiredChecks = []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}

type repositoryRef struct {
	FullName string `json:"full_name"`
}

type WorkflowRun struct {
	ID             int64         `json:"id"`
	WorkflowID     int64         `json:"workflow_id"`
	Event          string        `json:"event"`
	Ref            string        `json:"ref"`
	HeadBranch     string        `json:"head_branch"`
	HeadSHA        string        `json:"head_sha"`
	HeadRepository repositoryRef `json:"head_repository"`
	Status         string        `json:"status"`
	Conclusion     string        `json:"conclusion"`
	RunAttempt     int64         `json:"run_attempt"`
	CreatedAt      string        `json:"created_at"`
	RunStartedAt   string        `json:"run_started_at"`
	HTMLURL        string        `json:"html_url"`
}

type Job struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	HeadSHA     string `json:"head_sha"`
	RunID       int64  `json:"run_id"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
}

type Jobs struct {
	Jobs []Job `json:"jobs"`
}

type HistorySample struct {
	WorkflowRun
	JobLookupStatus string `json:"job_lookup_status"`
	Jobs            []Job  `json:"jobs"`
}

type History struct {
	Schema       string          `json:"schema"`
	LookupStatus string          `json:"lookup_status"`
	Samples      []HistorySample `json:"samples"`
}

type SampleObservation struct {
	RunID        int64            `json:"run_id"`
	RunAttempt   int64            `json:"run_attempt"`
	HeadSHA      string           `json:"head_sha"`
	CreatedAt    string           `json:"created_at"`
	RunStartedAt string           `json:"run_started_at"`
	HTMLURL      string           `json:"html_url"`
	Durations    map[string]int64 `json:"durations_wall_ms"`
}

type CheckBaseline struct {
	Name              string `json:"name"`
	MedianWallMS      int64  `json:"median_wall_ms"`
	CurrentWallMS     *int64 `json:"current_wall_ms,omitempty"`
	DeltaFromMedianMS *int64 `json:"delta_from_median_wall_ms,omitempty"`
}

type ExcludedSample struct {
	RunID  int64  `json:"run_id"`
	Reason string `json:"reason"`
}

type Report struct {
	Schema               string              `json:"schema"`
	MetricID             string              `json:"metric_id"`
	Repository           string              `json:"repository"`
	WorkflowID           int64               `json:"workflow_id"`
	Event                string              `json:"event"`
	Ref                  string              `json:"ref"`
	HeadBranch           string              `json:"head_branch"`
	HeadSHA              string              `json:"head_sha"`
	CurrentRunID         int64               `json:"current_run_id"`
	CurrentRunAttempt    int64               `json:"current_run_attempt"`
	CurrentRunConclusion string              `json:"current_run_conclusion"`
	HistoryInputDigest   string              `json:"history_input_digest"`
	LookupStatus         string              `json:"lookup_status"`
	BaselineState        string              `json:"baseline_state"`
	ComparisonState      string              `json:"comparison_state"`
	Reason               string              `json:"reason,omitempty"`
	RequiredChecks       []string            `json:"required_checks"`
	MinimumSamples       int                 `json:"minimum_samples"`
	MaximumSamples       int                 `json:"maximum_samples"`
	SampleCount          int                 `json:"sample_count"`
	Samples              []SampleObservation `json:"samples"`
	ExcludedSamples      []ExcludedSample    `json:"excluded_samples"`
	Checks               []CheckBaseline     `json:"checks"`
	Interpretation       string              `json:"interpretation"`
	ReportDigest         string              `json:"report_digest"`
}

func main() {
	var runPath, jobsPath, historyPath, outputPath, markdownPath string
	flag.StringVar(&runPath, "run", "", "exact current GitHub Actions run JSON")
	flag.StringVar(&jobsPath, "jobs", "", "exact current run jobs JSON")
	flag.StringVar(&historyPath, "history", "", "recent successful exact-tuple runs and jobs JSON")
	flag.StringVar(&outputPath, "output", "", "JSON report path; defaults to stdout")
	flag.StringVar(&markdownPath, "markdown", "", "Markdown summary path")
	flag.Parse()

	var current WorkflowRun
	if _, err := readJSON(runPath, &current); err != nil {
		fail(err)
	}
	var jobs Jobs
	if _, err := readJSON(jobsPath, &jobs); err != nil {
		fail(err)
	}
	var history History
	historyBytes, err := readJSON(historyPath, &history)
	if err != nil {
		fail(err)
	}
	report := buildReport(current, jobs.Jobs, history, historyBytes)
	report.ReportDigest = seal(report)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fail(err)
	}
	data = append(data, '\n')
	if outputPath == "" {
		if _, err := os.Stdout.Write(data); err != nil {
			fail(err)
		}
	} else if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		fail(err)
	}
	if markdownPath != "" {
		if err := os.WriteFile(markdownPath, []byte(markdown(report)), 0o644); err != nil {
			fail(err)
		}
	}
}

func readJSON(path string, destination any) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("input path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, destination); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return data, nil
}

func buildReport(current WorkflowRun, currentJobs []Job, history History, historyBytes []byte) Report {
	digest := sha256.Sum256(historyBytes)
	report := Report{
		Schema: reportSchema, MetricID: "gooo.metric.ci.required-check-baseline.v1",
		Repository: current.HeadRepository.FullName, WorkflowID: current.WorkflowID,
		Event: current.Event, Ref: current.Ref, HeadBranch: current.HeadBranch, HeadSHA: current.HeadSHA,
		CurrentRunID: current.ID, CurrentRunAttempt: current.RunAttempt, CurrentRunConclusion: current.Conclusion,
		HistoryInputDigest: "sha256:" + hex.EncodeToString(digest[:]), LookupStatus: history.LookupStatus,
		BaselineState: "UNKNOWN", ComparisonState: "UNKNOWN", RequiredChecks: append([]string(nil), requiredChecks...),
		MinimumSamples: minSamples, MaximumSamples: maxSamples, Samples: []SampleObservation{},
		ExcludedSamples: []ExcludedSample{}, Checks: []CheckBaseline{},
		Interpretation: "Descriptive same-tuple baseline only; job durations exclude queue time and do not establish causal savings.",
	}
	if history.Schema != inputSchema {
		report.Reason = "HISTORY_SCHEMA_INVALID"
		return report
	}
	if history.LookupStatus != "OK" {
		report.Reason = "HISTORY_API_UNAVAILABLE"
		return report
	}
	if !validIdentity(current) {
		report.Reason = "CURRENT_RUN_IDENTITY_INCOMPLETE"
		return report
	}
	if len(history.Samples) == 0 {
		report.Reason = "HISTORY_SAMPLE_COUNT_BELOW_MINIMUM"
		return report
	}

	ordered := append([]HistorySample(nil), history.Samples...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID > ordered[j].ID })
	counts := make(map[int64]int, len(ordered))
	for _, sample := range ordered {
		counts[sample.ID]++
	}
	durationsByCheck := make(map[string][]int64, len(requiredChecks))
	for _, name := range requiredChecks {
		durationsByCheck[name] = []int64{}
	}
	for _, sample := range ordered {
		if counts[sample.ID] > 1 {
			report.ExcludedSamples = append(report.ExcludedSamples, ExcludedSample{RunID: sample.ID, Reason: "HISTORY_RUN_ID_DUPLICATE"})
			continue
		}
		durations, reason := observeSample(current, sample)
		if reason != "" {
			report.ExcludedSamples = append(report.ExcludedSamples, ExcludedSample{RunID: sample.ID, Reason: reason})
			continue
		}
		if len(report.Samples) >= maxSamples {
			continue
		}
		report.Samples = append(report.Samples, SampleObservation{
			RunID: sample.ID, RunAttempt: sample.RunAttempt, HeadSHA: sample.HeadSHA,
			CreatedAt: sample.CreatedAt, RunStartedAt: sample.RunStartedAt, HTMLURL: sample.HTMLURL,
			Durations: durations,
		})
		for _, name := range requiredChecks {
			durationsByCheck[name] = append(durationsByCheck[name], durations[name])
		}
	}
	report.SampleCount = len(report.Samples)
	if report.SampleCount < minSamples {
		report.Reason = "HISTORY_SAMPLE_COUNT_BELOW_MINIMUM"
		return report
	}
	report.BaselineState = "OBSERVED"

	currentDurations, currentReason := observeCurrent(current, currentJobs)
	if currentReason == "" {
		report.ComparisonState = "OBSERVED"
	} else {
		report.Reason = currentReason
	}
	for _, name := range requiredChecks {
		median := median(durationsByCheck[name])
		baseline := CheckBaseline{Name: name, MedianWallMS: median}
		if currentReason == "" {
			value := currentDurations[name]
			delta := value - median
			baseline.CurrentWallMS = &value
			baseline.DeltaFromMedianMS = &delta
		}
		report.Checks = append(report.Checks, baseline)
	}
	return report
}

func validIdentity(run WorkflowRun) bool {
	return run.ID > 0 && run.WorkflowID > 0 && run.RunAttempt > 0 && run.Event != "" && run.Ref != "" && run.HeadBranch != "" && run.HeadRepository.FullName != "" && validSHA(run.HeadSHA)
}

func observeSample(current WorkflowRun, sample HistorySample) (map[string]int64, string) {
	if sample.JobLookupStatus != "OK" {
		return nil, "HISTORY_JOB_API_UNAVAILABLE"
	}
	if !validIdentity(sample.WorkflowRun) || sample.ID == current.ID || sample.Status != "completed" || sample.Conclusion != "success" {
		return nil, "HISTORY_RUN_NOT_SUCCESSFUL_OR_TERMINAL"
	}
	if sample.HeadRepository.FullName != current.HeadRepository.FullName || sample.WorkflowID != current.WorkflowID || sample.Event != current.Event || sample.Ref != current.Ref || sample.HeadBranch != current.HeadBranch {
		return nil, "HISTORY_RUN_TUPLE_MISMATCH"
	}
	return observeJobs(sample.WorkflowRun, sample.Jobs, true)
}

func observeCurrent(current WorkflowRun, jobs []Job) (map[string]int64, string) {
	if current.Status != "completed" || current.Conclusion != "success" {
		return nil, "CURRENT_RUN_NOT_SUCCESSFUL_OR_TERMINAL"
	}
	return observeJobs(current, jobs, true)
}

func observeJobs(run WorkflowRun, jobs []Job, requireSuccess bool) (map[string]int64, string) {
	byName := make(map[string]Job, len(requiredChecks))
	required := make(map[string]struct{}, len(requiredChecks))
	for _, name := range requiredChecks {
		required[name] = struct{}{}
	}
	for _, job := range jobs {
		if _, ok := required[job.Name]; !ok {
			continue
		}
		if _, duplicate := byName[job.Name]; duplicate {
			return nil, "REQUIRED_CHECK_AMBIGUOUS"
		}
		byName[job.Name] = job
	}
	if len(byName) != len(requiredChecks) {
		return nil, "REQUIRED_CHECK_MISSING"
	}
	durations := make(map[string]int64, len(requiredChecks))
	for _, name := range requiredChecks {
		job := byName[name]
		if job.ID <= 0 || job.RunID != run.ID || job.HeadSHA != run.HeadSHA || job.Status != "completed" || job.Conclusion == "skipped" {
			return nil, "REQUIRED_CHECK_NOT_TERMINAL_OR_EXACT_HEAD"
		}
		if requireSuccess && job.Conclusion != "success" {
			return nil, "REQUIRED_CHECK_NOT_SUCCESSFUL"
		}
		started, err := time.Parse(time.RFC3339, job.StartedAt)
		if err != nil {
			return nil, "REQUIRED_CHECK_START_TIMESTAMP_INVALID"
		}
		completed, err := time.Parse(time.RFC3339, job.CompletedAt)
		if err != nil {
			return nil, "REQUIRED_CHECK_END_TIMESTAMP_INVALID"
		}
		wall := completed.Sub(started).Milliseconds()
		if wall < 1000 {
			return nil, "REQUIRED_CHECK_BELOW_SOURCE_RESOLUTION"
		}
		durations[name] = wall
	}
	return durations, ""
}

func median(values []int64) int64 {
	sorted := append([]int64(nil), values...)
	slices.Sort(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func validSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func seal(report Report) string {
	report.ReportDigest = ""
	data, _ := json.Marshal(report)
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func markdown(report Report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "CI required-check baseline: %s/%s reason=%s\n", report.BaselineState, report.ComparisonState, report.Reason)
	fmt.Fprintf(&builder, "metric=%s run=%d attempt=%d workflow=%d event=%s ref=%s samples=%d/%d lookup=%s\n", report.MetricID, report.CurrentRunID, report.CurrentRunAttempt, report.WorkflowID, report.Event, report.Ref, report.SampleCount, report.MaximumSamples, report.LookupStatus)
	for _, check := range report.Checks {
		if check.CurrentWallMS == nil {
			fmt.Fprintf(&builder, "check=%q median_wall_ms=%d current=UNKNOWN\n", check.Name, check.MedianWallMS)
			continue
		}
		fmt.Fprintf(&builder, "check=%q median_wall_ms=%d current_wall_ms=%d delta_from_median_wall_ms=%d\n", check.Name, check.MedianWallMS, *check.CurrentWallMS, *check.DeltaFromMedianMS)
	}
	fmt.Fprintf(&builder, "input_digest=%s report_digest=%s; descriptive only, no causal savings or gate claim\n", report.HistoryInputDigest, report.ReportDigest)
	return builder.String()
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
