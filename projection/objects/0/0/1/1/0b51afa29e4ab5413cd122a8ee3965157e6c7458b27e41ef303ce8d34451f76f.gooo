package main

import (
	"fmt"
	"strings"
)

func validEventRef(event, ref string) bool {
	if ref == "" || strings.ContainsAny(ref, "\r\n") {
		return false
	}
	if event == "pull_request" {
		return strings.HasPrefix(ref, "refs/pull/") && strings.HasSuffix(ref, "/merge")
	}
	if event == "push" {
		return strings.HasPrefix(ref, "refs/heads/")
	}
	return false
}
func normalizeJobs(apiJobs []apiJob, headSHA string, runID, runAttempt int64) ([]jobEvidence, error) {
	byName := make(map[string]apiJob)
	seenIDs := make(map[int64]bool)
	for _, job := range apiJobs {
		if !jobSet()[job.Name] {
			continue
		}
		if _, duplicate := byName[job.Name]; duplicate {
			return nil, fmt.Errorf("duplicate canonical CI job %q", job.Name)
		}
		if job.ID <= 0 || seenIDs[job.ID] {
			return nil, fmt.Errorf("duplicate or invalid canonical CI job id %d", job.ID)
		}
		seenIDs[job.ID] = true
		byName[job.Name] = job
	}
	result := make([]jobEvidence, 0, len(canonicalJobs))
	for _, name := range canonicalJobs {
		job, ok := byName[name]
		if !ok || job.ID <= 0 || job.Status != "completed" || job.Conclusion != "success" || !validSHA(job.HeadSHA) || job.HeadSHA != headSHA || job.RunID != runID {
			return nil, fmt.Errorf("canonical CI job %q is missing or mismatched", name)
		}
		result = append(result, jobEvidence{ID: job.ID, Name: job.Name, Status: job.Status, Conclusion: job.Conclusion, HeadSHA: job.HeadSHA, RunID: job.RunID, RunAttempt: runAttempt})
	}
	return result, nil
}
func jobSet() map[string]bool {
	result := make(map[string]bool, len(canonicalJobs))
	for _, name := range canonicalJobs {
		result[name] = true
	}
	return result
}
