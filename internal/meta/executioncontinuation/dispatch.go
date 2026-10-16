package executioncontinuation

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
)

// Dispatch records creation of a target run. It never waits for target completion.
func (c Client) Dispatch(ctx context.Context, e Envelope) (DispatchReceipt, error) {
	r := DispatchReceipt{Schema: Schema, State: policy.ObserveStage(false, false, false), Source: e, PolicyDigest: policy.SourceDigest()}
	if err := e.Validate(); err != nil {
		r.State = "CONTINUATION_REFUTED"
		return sealDispatch(r, err.Error()), err
	}
	if e.Repository != c.Repository {
		r.State = "CONTINUATION_REFUTED"
		return sealDispatch(r, "SOURCE_REPOSITORY_MISMATCH"), fmt.Errorf("source repository mismatch")
	}
	g, _, err := c.loadGrant(ctx, e, false)
	if err != nil {
		return failedDispatch(r, err), err
	}
	if err := ValidateGrant(c.Root, g, e); err != nil {
		return failedDispatch(r, err), err
	}
	if err := c.completedGrantJob(ctx, e); err != nil {
		return sealDispatch(r, "SOURCE_GRANT_JOB_UNAVAILABLE"), err
	}
	body := struct {
		Ref    string            `json:"ref"`
		Inputs map[string]string `json:"inputs"`
	}{"dev", dispatchInputs(e)}
	raw, err := c.request(ctx, http.MethodPost, "actions/workflows/self-improvement-execution-consumer.yml/dispatches", body)
	if err != nil {
		return sealDispatch(r, "DISPATCH_RESPONSE_UNKNOWN"), err
	}
	r.Accepted = true
	r.State = policy.ObserveStage(true, false, false)
	var response struct {
		ID int64 `json:"workflow_run_id"`
	}
	if decode(raw, &response) != nil || response.ID <= 0 {
		return sealDispatch(r, "TARGET_REFERENCE_UNKNOWN"), nil
	}
	r.TargetRunID = response.ID
	if policy.BindTargetRun(e.RunID, response.ID, response.ID) != "TARGET_RUN_BOUND" {
		r.State = "CONTINUATION_REFUTED"
		return sealDispatch(r, "TARGET_RUN_REFUTED"), fmt.Errorf("target run is parent")
	}
	var run Run
	if err := c.get(ctx, fmt.Sprintf("actions/runs/%d", response.ID), &run); err != nil {
		return sealDispatch(r, "TARGET_RUN_UNAVAILABLE"), nil
	}
	if err := c.runIdentity(run, "consumer", response.ID, 1); err != nil {
		state, _ := errorState(err)
		if state == "UNKNOWN" {
			return sealDispatch(r, "TARGET_RUN_METADATA_UNKNOWN"), nil
		}
		r.State = "CONTINUATION_REFUTED"
		return sealDispatch(r, "TARGET_RUN_MISMATCH"), err
	}
	r.TargetWorkflowSHA = run.Head
	r.State = policy.ObserveStage(true, true, false)
	return sealDispatch(r, "EXACT_TARGET_RUN_CREATED"), nil
}

func dispatchInputs(e Envelope) map[string]string {
	return map[string]string{"source_repository": e.Repository, "source_head_sha": e.WorkflowSHA, "subject_sha": e.SubjectSHA,
		"source_run_id": strconv.FormatInt(e.RunID, 10), "source_run_attempt": strconv.Itoa(e.Attempt),
		"source_artifact_id": strconv.FormatInt(e.ArtifactID, 10), "source_artifact_digest": e.ArchiveDigest,
		"grant_request_digest": e.RequestDigest, "source_wait": e.Wait, "policy_digest": e.PolicyDigest}
}

func (c Client) completedGrantJob(ctx context.Context, e Envelope) error {
	var jobs struct {
		Total int                                         `json:"total_count"`
		Jobs  []struct{ Name, Status, Conclusion string } `json:"jobs"`
	}
	if err := c.get(ctx, fmt.Sprintf("actions/runs/%d/attempts/%d/jobs?per_page=100", e.RunID, e.Attempt), &jobs); err != nil {
		return err
	}
	if jobs.Total != len(jobs.Jobs) {
		return fmt.Errorf("source job inventory incomplete")
	}
	matches := 0
	for _, j := range jobs.Jobs {
		if j.Name == "execution-grant" && j.Status == "completed" && j.Conclusion == "success" {
			matches++
		}
	}
	if matches != 1 {
		return fmt.Errorf("source grant job not complete")
	}
	return nil
}
