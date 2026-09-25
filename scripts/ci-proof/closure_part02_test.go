package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNoFailureClosureDoesNotWriteOnInvalidInput(t *testing.T) {
	root := t.TempDir()
	inputPath := filepath.Join(root, "closure-input.json")
	outputPath := filepath.Join(root, "closure.json")
	data, err := json.Marshal(closureInput{CanonicalJobs: validClosureInput().CanonicalJobs[:1]})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(inputPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	setFailureBindingEnvironment(t, validFailureBinding())
	if err := writeClosureManifest(inputPath, outputPath); err == nil {
		t.Fatal("invalid closure input was accepted")
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("invalid closure wrote output: %v", err)
	}
}
func validClosureInput() closureInput {
	binding := validFailureBinding()
	return validClosureInputFor(binding)
}
func validClosureInputFor(binding failureBinding) closureInput {
	jobs := make([]failureJob, len(proofJobs))
	for index, name := range proofJobs {
		jobs[index] = failureJob{ID: int64(index + 1), Name: name, Status: "completed", Conclusion: "success", HeadSHA: binding.HeadSHA, RunID: binding.RunID, RunAttempt: binding.RunAttempt}
	}
	return closureInput{CanonicalJobs: jobs, TerminalFailures: []failureJob{}, TerminalFailureCodes: []string{}}
}
func setFailureBindingEnvironment(t *testing.T, binding failureBinding) {
	t.Helper()
	t.Setenv("CI_REPOSITORY", binding.Repository)
	t.Setenv("CI_EVENT", binding.Event)
	t.Setenv("CI_EVENT_REF", binding.EventRef)
	t.Setenv("CI_CHECKOUT_REF", binding.CheckoutRef)
	t.Setenv("CI_BASE_REF", binding.BaseRef)
	t.Setenv("CI_BASE_SHA", binding.BaseSHA)
	t.Setenv("CI_HEAD_SHA", binding.HeadSHA)
	t.Setenv("CI_WORKFLOW_SHA", binding.WorkflowSHA)
	t.Setenv("CI_PR_NUMBER", "7")
	t.Setenv("CI_RUN_ID", "9")
	t.Setenv("CI_RUN_ATTEMPT", "2")
	t.Setenv("CI_ACTOR", binding.Actor)
	t.Setenv("CI_OWNER_BRANCH", binding.OwnerBranch)
}
