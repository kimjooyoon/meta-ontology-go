package main

import "fmt"

func validateSchedulerInputs(results []schedulerInput, head string, runID, runAttempt int64) (map[string]schedulerInput, error) {
	if len(results) != len(proofJobs) {
		return nil, fmt.Errorf("same-run scheduler evidence must contain all six canonical jobs")
	}
	sources := map[string]string{
		"gofmt":                "needs.format.result",
		"go vet":               "needs.vet.result",
		"go test":              "needs.test.result",
		"go test -race":        "needs.race.result",
		"Semantic conformance": "needs.semantic.result",
		"CI policy":            "needs.policy.result",
	}
	byName := make(map[string]schedulerInput, len(results))
	for _, result := range results {
		if !isProofJob(result.Name) || byName[result.Name].Name != "" {
			return nil, fmt.Errorf("duplicate or unknown same-run scheduler result %q", result.Name)
		}
		if result.Source != sources[result.Name] || result.Result != "success" || result.HeadSHA != head || result.RunID != runID || result.RunAttempt != runAttempt {
			return nil, fmt.Errorf("same-run scheduler result %q is missing, unsuccessful, or mismatched", result.Name)
		}
		byName[result.Name] = result
	}
	for _, name := range proofJobs {
		if _, ok := byName[name]; !ok {
			return nil, fmt.Errorf("same-run scheduler result %q is missing", name)
		}
	}
	return byName, nil
}

func validateSchedulerAgreement(evidence, final []schedulerInput, head string, runID, runAttempt int64) error {
	if _, err := validateSchedulerInputs(evidence, head, runID, runAttempt); err != nil {
		return fmt.Errorf("downloaded evidence scheduler is invalid: %w", err)
	}
	if _, err := validateSchedulerInputs(final, head, runID, runAttempt); err != nil {
		return fmt.Errorf("proof-side scheduler is invalid: %w", err)
	}
	if len(evidence) != len(final) {
		return fmt.Errorf("proof-side scheduler does not exactly match downloaded evidence scheduler")
	}
	for index := range evidence {
		if evidence[index] != final[index] {
			return fmt.Errorf("proof-side scheduler does not exactly match downloaded evidence scheduler at index %d", index)
		}
	}
	return nil
}

func jobObservationState(job jobInput, scheduler schedulerInput) (string, error) {
	if job.Status == nil || *job.Status == "" {
		return "", fmt.Errorf("raw API status is missing")
	}
	if !validRawJobStatus(*job.Status) {
		return "", fmt.Errorf("raw API status %q is malformed", *job.Status)
	}
	if job.Conclusion != nil && !validRawJobConclusion(*job.Conclusion) {
		return "", fmt.Errorf("raw API conclusion %q is malformed", *job.Conclusion)
	}
	if scheduler.Name == "" {
		if *job.Status == "completed" && job.Conclusion != nil && *job.Conclusion == "success" {
			return apiTerminalSuccess, nil
		}
		return "", fmt.Errorf("raw API job is not terminally successful without scheduler evidence")
	}
	if scheduler.Result != "success" {
		return "", fmt.Errorf("same-run scheduler result is not success")
	}
	if job.Conclusion != nil && *job.Status != "completed" {
		return "", fmt.Errorf("raw API has a nonterminal status with a terminal conclusion")
	}
	if job.Conclusion != nil && *job.Conclusion != "success" {
		return "", fmt.Errorf("raw API conclusion %q contradicts same-run scheduler success", *job.Conclusion)
	}
	if *job.Status == "completed" && job.Conclusion != nil && *job.Conclusion == "success" {
		return apiTerminalSuccess, nil
	}
	return observerLag, nil
}

func validRawJobStatus(status string) bool {
	switch status {
	case "queued", "in_progress", "completed", "waiting", "requested", "pending":
		return true
	default:
		return false
	}
}

func validRawJobConclusion(conclusion string) bool {
	switch conclusion {
	case "success", "failure", "neutral", "cancelled", "skipped", "timed_out", "action_required", "stale", "startup_failure":
		return true
	default:
		return false
	}
}

func sameOptionalString(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
