package main

import (
	"fmt"
	"os"
)

func validateFailureJob(job failureJob, binding failureBinding) error {
	if job.ID <= 0 || job.Name == "" || job.Name == "CI failure summary" || containsUnknown(job.Name) || job.Status != "completed" || job.Conclusion == "success" || job.Conclusion == "" || containsUnknown(job.Conclusion) || job.RunID != binding.RunID || job.RunAttempt != binding.RunAttempt || job.HeadSHA != binding.HeadSHA {
		return fmt.Errorf("failure job is not bound to the exact run tuple")
	}
	return nil
}
func isFailureJobName(name string) bool {
	if name == "CI proof bundle" {
		return true
	}
	for _, canonical := range proofJobs {
		if name == canonical {
			return true
		}
	}
	return false
}
func buildFailureProvenance(manifest failureManifest, binding failureBinding) failureProvenance {
	runRef := fmt.Sprintf("https://github.com/%s/actions/runs/%d", binding.Repository, binding.RunID)
	jobRef := fmt.Sprintf("%s/job/%d", runRef, manifest.Job.ID)
	sourceRef := fmt.Sprintf("https://github.com/%s/commit/%s", binding.Repository, binding.HeadSHA)
	return failureProvenance{
		WasGeneratedBy:    manifest.Activity,
		WasAssociatedWith: manifest.Agent,
		WasDerivedFrom:    []string{runRef, jobRef},
		HadPrimarySource:  append([]string{sourceRef, manifest.OwnerRef, manifest.CatalogRef, manifest.CatalogSHA256}, manifest.ArtifactURLs...),
	}
}
func readFailureBinding() (failureBinding, error) {
	values := map[string]string{
		"repository": os.Getenv("CI_REPOSITORY"), "event": os.Getenv("CI_EVENT"), "event_ref": os.Getenv("CI_EVENT_REF"), "checkout_ref": os.Getenv("CI_CHECKOUT_REF"),
		"base_ref": os.Getenv("CI_BASE_REF"), "base_sha": os.Getenv("CI_BASE_SHA"), "head_sha": os.Getenv("CI_HEAD_SHA"),
		"workflow_sha": os.Getenv("CI_WORKFLOW_SHA"), "actor": os.Getenv("CI_ACTOR"), "owner_branch": os.Getenv("CI_OWNER_BRANCH"),
	}
	for name, value := range values {
		if value == "" || containsUnknown(value) {
			return failureBinding{}, fmt.Errorf("missing or unknown failure binding field %s", name)
		}
	}
	if !validSHA(values["base_sha"]) || !validSHA(values["head_sha"]) || !validSHA(values["workflow_sha"]) || values["base_sha"] == values["head_sha"] || !validEventRef(values["event"], values["event_ref"]) || values["checkout_ref"] != values["head_sha"] {
		return failureBinding{}, fmt.Errorf("failure binding revisions or event ref are invalid")
	}
	runID, err := failurePositiveInt("CI_RUN_ID")
	if err != nil {
		return failureBinding{}, err
	}
	runAttempt, err := failurePositiveInt("CI_RUN_ATTEMPT")
	if err != nil {
		return failureBinding{}, err
	}
	prNumber, err := failureOptionalInt("CI_PR_NUMBER")
	if err != nil {
		return failureBinding{}, err
	}
	return failureBinding{Repository: values["repository"], Event: values["event"], EventRef: values["event_ref"], CheckoutRef: values["checkout_ref"], BaseRef: values["base_ref"], BaseSHA: values["base_sha"], HeadSHA: values["head_sha"], WorkflowSHA: values["workflow_sha"], PRNumber: prNumber, RunID: runID, RunAttempt: runAttempt, Actor: values["actor"], OwnerBranch: values["owner_branch"]}, nil
}
