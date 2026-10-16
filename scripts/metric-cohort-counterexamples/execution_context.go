package main

import (
	"fmt"
	"strconv"
	"strings"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

type executionOrigin struct {
	Kind                 string `json:"kind"`
	Event                string `json:"event,omitempty"`
	Repository           string `json:"repository,omitempty"`
	CheckoutSHA          string `json:"checkout_sha,omitempty"`
	WorkflowSHA          string `json:"workflow_sha,omitempty"`
	WorkflowRef          string `json:"workflow_ref,omitempty"`
	RunID                int64  `json:"run_id,omitempty"`
	Attempt              int    `json:"run_attempt,omitempty"`
	Job                  string `json:"job,omitempty"`
	GoVersion            string `json:"go_version"`
	GOOS                 string `json:"goos"`
	GOARCH               string `json:"goarch"`
	Clean                bool   `json:"clean_worktree"`
	GitObservationBudget int64  `json:"git_observation_budget_nanoseconds"`
}

func ciExecution(expected origin, env map[string]string, checkout string, clean bool) (executionOrigin, error) {
	value := executionOrigin{Kind: "ORIGINAL_GITHUB_ACTIONS", Event: env["GITHUB_EVENT_NAME"], Repository: env["GITHUB_REPOSITORY"],
		CheckoutSHA: checkout, WorkflowSHA: env["GITHUB_WORKFLOW_SHA"], WorkflowRef: env["GITHUB_WORKFLOW_REF"],
		Job: env["GITHUB_JOB"], Clean: clean}
	run, runErr := strconv.ParseInt(env["GITHUB_RUN_ID"], 10, 64)
	attempt, attemptErr := strconv.Atoi(env["GITHUB_RUN_ATTEMPT"])
	if runErr != nil || attemptErr != nil || run != expected.RunID || attempt != 1 ||
		attempt != expected.Attempt || value.Repository != expected.Repository || checkout != expected.SubjectSHA || !clean ||
		!artifact.ValidSubject(expected.Repository, checkout) {
		return value, fmt.Errorf("countercase execution differs from the exact original source/run/attempt")
	}
	if value.Job != "concept-operation-observation" || !artifact.ValidSubject(value.Repository, value.WorkflowSHA) ||
		!strings.HasPrefix(value.WorkflowRef, value.Repository+"/.github/workflows/") {
		return value, fmt.Errorf("countercase execution workflow/job origin is unavailable")
	}
	if value.Event != "pull_request" && value.Event != "push" {
		return value, fmt.Errorf("countercase execution requires an original pull_request or push event")
	}
	value.RunID, value.Attempt = run, attempt
	return value, nil
}
