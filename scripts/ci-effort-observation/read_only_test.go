package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/publicworkflowlineage"
)

func readOnlyTestPolicy(t *testing.T) publicworkflowlineage.Policy {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate read-only projection test source")
	}
	path := filepath.Join(filepath.Dir(filename), "../../examples/ci-effort-observation/main.gooo")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := publicworkflowlineage.Load(path, source)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func readOnlyTestLineageInput(policy publicworkflowlineage.Policy) (sourceRunInput, readOnlyLineageObservation, publicworkflowlineage.ReadOnlyObservationEvaluation) {
	head := "0123456789abcdef0123456789abcdef01234567"
	source := sourceRunInput{ID: 42, Name: "CI", WorkflowName: ".github/workflows/ci.yml", Event: "pull_request", Ref: "refs/pull/1/merge", HeadBranch: "feature", HeadSHA: head, HeadRepository: struct {
		FullName string `json:"full_name"`
	}{FullName: "kimjooyoon/meta-ontology-go"}, Status: "completed", Conclusion: "failure", RunAttempt: 1}
	lineageSource := publicworkflowlineage.SourceRun{ID: source.ID, Name: source.Name, Workflow: source.WorkflowName, WorkflowPath: source.WorkflowName, WorkflowID: 7, Event: source.Event, Ref: source.Ref, RefState: publicworkflowlineage.RefStateValue, HeadBranch: source.HeadBranch, HeadSHA: source.HeadSHA, Repository: source.HeadRepository.FullName, APIRepositoryName: source.HeadRepository.FullName, APIQueryRunID: source.ID, ResolvedBy: "actions-run-api:workflow_run.id", Status: source.Status, Conclusion: source.Conclusion, RunAttempt: source.RunAttempt}
	trigger := publicworkflowlineage.Trigger{SourceWorkflow: source.WorkflowName, SourceRunID: source.ID, SourceRunAttempt: source.RunAttempt, SourceSubjectSHA: source.HeadSHA, SourceRef: source.Ref, SourceRefState: publicworkflowlineage.RefStateValue, SourceHeadBranch: source.HeadBranch, SourceEvent: source.Event, SourceRepository: source.HeadRepository.FullName, ConsumerWorkflow: "CI effort observation", ConsumerRunID: 100, ConsumerRunAttempt: 1, ConsumerSubjectSHA: source.HeadSHA, ConsumerRef: source.Ref}
	input := publicworkflowlineage.Input{Trigger: trigger, Source: lineageSource, ExpectedArtifactName: "ci-evidence-42-1", ExpectedRepository: policy.Repository, ExpectedWorkflow: policy.SourceWorkflow, ExpectedSourceAPIKey: policy.SourceAPIKey, ExpectedArtifactSubjectBinding: policy.ArtifactSubjectBinding}
	evaluation := publicworkflowlineage.Evaluate(input)
	observation := policy.EvaluateReadOnlyObservation(input)
	lineage := readOnlyLineageObservation{Schema: publicworkflowlineage.ReportSchema, Decision: evaluation.Decision, LineageState: evaluation.LineageState, Reason: evaluation.Reason, Trigger: trigger, Source: lineageSource, Evaluation: evaluation, PolicyDigest: policy.SourceDigest}
	return source, lineage, observation
}

func TestReadOnlyTimingPreservesMissingRuntimeAsUnknown(t *testing.T) {
	source := sourceRunInput{ID: 42, RunStartedAt: "2026-08-30T00:00:00Z", UpdatedAt: "2026-08-30T00:00:02Z"}
	jobs, window, err := observeJobsWithSource([]APIJob{{ID: 7, RunID: 42, Name: "check", Status: "completed", Conclusion: "failure", Steps: []APIStep{{Name: "Verify", Status: "completed", Conclusion: "failure"}}}}, source)
	if err != nil || len(jobs) != 1 || jobs[0].Unknown == nil || jobs[0].WallMS != 0 || jobs[0].Steps[0].Unknown == nil || jobs[0].Steps[0].WallMS != 0 {
		t.Fatalf("missing source timing was not kept unknown: jobs=%+v window=%+v err=%v", jobs, window, err)
	}
	timing := summarizeReadOnlyTiming(window, jobs, false)
	if timing.ObservedJobIntervals != 0 || timing.ObservedStepIntervals != 0 || timing.MissingJobIntervals != 1 || timing.MissingStepIntervals != 1 || timing.WindowWallMS != 2000 {
		t.Fatalf("read-only timing summary lost missing intervals: %+v", timing)
	}
}

func TestReadOnlyOperationCountsKeepUnknownOperationsSeparate(t *testing.T) {
	specs := []OperationSpec{{ID: "check", JobName: "check", StepName: "Verify", Kind: "VERIFICATION", Command: []string{"go", "test"}, ProofObligationID: "ci-effort/check"}}
	operations, accounting := observeOperations(specs, nil, ".github/workflows/ci.yml", nil, nil, "push")
	if len(operations) != 1 || operations[0].State != "UNKNOWN" || operations[0].WallMS != 0 || accounting.Unknown != 1 || accounting.Executed != 0 || accounting.Skipped != 0 || accounting.Rejected != 0 {
		t.Fatalf("unknown operation was not accounted separately: operations=%+v accounting=%+v", operations, accounting)
	}
	counts := readOnlyCounts(accounting)
	if counts.Manifest != 1 || counts.Missing != 1 || counts.Observed != 0 || counts.Skipped != 0 || counts.Rejected != 0 {
		t.Fatalf("read-only operation counts lost missing operation: %+v", counts)
	}
}

func TestReadOnlyLineageInputsRejectForgedEligibleReceipt(t *testing.T) {
	policy := readOnlyTestPolicy(t)
	source, lineage, observation := readOnlyTestLineageInput(policy)
	if err := validateReadOnlyLineageInputs(source, lineage, observation, policy); err != nil {
		t.Fatalf("canonical failed-source observation was rejected: %v", err)
	}
	lineage.Trigger.SourceSubjectSHA = "fedcba9876543210fedcba9876543210fedcba98"
	if err := validateReadOnlyLineageInputs(source, lineage, observation, policy); err == nil {
		t.Fatal("forged read-only eligibility was accepted")
	}
}
