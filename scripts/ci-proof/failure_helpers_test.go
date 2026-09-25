package main

import "strings"

func validFailureBinding() failureBinding {
	return failureBinding{
		Repository: "owner/repo", Event: "pull_request", EventRef: "refs/pull/7/merge", CheckoutRef: strings.Repeat("a", 40), BaseRef: "dev",
		BaseSHA: strings.Repeat("b", 40), HeadSHA: strings.Repeat("a", 40), WorkflowSHA: strings.Repeat("c", 40),
		PRNumber: 7, RunID: 9, RunAttempt: 2, Actor: "builder", OwnerBranch: "agent/ci-workflow",
	}
}

func validFailureInput() failureInput {
	head := strings.Repeat("a", 40)
	job := failureJob{ID: 11, Name: "go test", Status: "completed", Conclusion: "failure", HeadSHA: head, RunID: 9, RunAttempt: 2}
	return failureInput{
		Code: "CI-TEST-001", FailureCodes: []string{"CI-TEST-001"}, Message: "go test failed in the exact PR run", Remediation: "reproduce and fix the failing test",
		OwnerBranch: "agent/ci-workflow", ArtifactStatus: "not_applicable", ArtifactReason: "canonical_job_failure",
		TerminalFailures: []failureJob{job}, TerminalFailureCodes: []string{"CI-TEST-001"}, Job: job,
	}
}
