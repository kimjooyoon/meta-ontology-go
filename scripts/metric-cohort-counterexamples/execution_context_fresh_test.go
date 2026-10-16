package main

import (
	"strings"
	"testing"
)

func TestCountercaseExecutionOriginsFresh1691(t *testing.T) {
	subject, workflow := strings.Repeat("a", 40), strings.Repeat("b", 40)
	for _, name := range []string{"distinct-workflow-source", "wrong-run", "rerun", "wrong-checkout",
		"wrong-job", "missing-workflow-source", "dispatch", "dirty-checkout"} {
		t.Run(name, func(t *testing.T) {
			expected := origin{Repository: "example/gooo", SubjectSHA: subject, RunID: 77, Attempt: 1}
			env := map[string]string{"GITHUB_REPOSITORY": expected.Repository, "GITHUB_RUN_ID": "77",
				"GITHUB_RUN_ATTEMPT": "1", "GITHUB_JOB": "concept-operation-observation",
				"GITHUB_WORKFLOW_SHA": workflow, "GITHUB_WORKFLOW_REF": "example/gooo/.github/workflows/metric.yml@refs/pull/1/merge",
				"GITHUB_EVENT_NAME": "pull_request"}
			checkout, clean := subject, true
			switch name {
			case "wrong-run":
				env["GITHUB_RUN_ID"] = "78"
			case "rerun":
				env["GITHUB_RUN_ATTEMPT"] = "2"
			case "wrong-checkout":
				checkout = workflow
			case "wrong-job":
				env["GITHUB_JOB"] = "program"
			case "missing-workflow-source":
				env["GITHUB_WORKFLOW_SHA"] = ""
			case "dispatch":
				env["GITHUB_EVENT_NAME"] = "workflow_dispatch"
			case "dirty-checkout":
				clean = false
			}
			observed, err := ciExecution(expected, env, checkout, clean)
			if (err == nil) != (name == "distinct-workflow-source") {
				t.Fatalf("source/execution origin binding: %+v, %v", observed, err)
			}
			if err == nil && (observed.CheckoutSHA != subject || observed.WorkflowSHA != workflow ||
				observed.CheckoutSHA == observed.WorkflowSHA || observed.RunID != 77) {
				t.Fatal("subject and workflow source roles were conflated")
			}
		})
	}
}
