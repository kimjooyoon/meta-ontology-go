package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	continuation "github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation"
)

func dispatch(ctx context.Context, c continuation.Client, o options) error {
	seed := continuation.Envelope{SubjectSHA: o.subject, WorkflowSHA: o.workflow, RunID: o.run, Attempt: o.attempt}
	envelope, err := c.DiscoverGrant(ctx, seed)
	receipt := continuation.Diagnostic(envelope, err)
	if err == nil {
		receipt, err = c.Dispatch(ctx, envelope)
	}
	if writeErr := writeJSON(o.output, receipt); writeErr != nil {
		return writeErr
	}
	if receipt.State == "CONTINUATION_REFUTED" {
		return fmt.Errorf("%s", receipt.Reason)
	}
	return nil
}

func prepare(ctx context.Context, c continuation.Client, o options) error {
	if err := os.MkdirAll(filepath.Dir(o.output), 0700); err != nil {
		return err
	}
	if err := os.Mkdir(o.output, 0700); err != nil {
		return err
	}
	e := envelopeFromEnvironment()
	if o.input != "" {
		if err := readJSON(o.input, &e); err != nil {
			return err
		}
	}
	p, err := c.Prepare(ctx, e, o.run, o.attempt)
	if err == nil {
		var exists bool
		exists, err = c.HasReservation(ctx, e)
		if exists {
			err = &continuation.Failure{State: "REFUTED", Reason: "GRANT_ALREADY_RESERVED"}
		}
	}
	receipt := continuation.Diagnostic(e, err)
	if err == nil {
		if err := writeJSON(filepath.Join(o.output, "prepared.json"), p); err != nil {
			return err
		}
		receipt = continuation.BoundSource(e)
	}
	if err := writeJSON(filepath.Join(o.output, "binding.json"), receipt); err != nil {
		return err
	}
	if receipt.State == "CONTINUATION_REFUTED" {
		return fmt.Errorf("%s", receipt.Reason)
	}
	return nil
}

func envelopeFromEnvironment() continuation.Envelope {
	return continuation.Envelope{Repository: os.Getenv("SOURCE_REPOSITORY"), WorkflowSHA: os.Getenv("SOURCE_HEAD_SHA"),
		SubjectSHA: os.Getenv("SUBJECT_SHA"), RunID: envInt("SOURCE_RUN_ID"), Attempt: envAttempt("SOURCE_RUN_ATTEMPT"),
		ArtifactID: envInt("SOURCE_ARTIFACT_ID"), ArchiveDigest: os.Getenv("SOURCE_ARTIFACT_DIGEST"),
		RequestDigest: os.Getenv("GRANT_REQUEST_DIGEST"), Wait: os.Getenv("SOURCE_WAIT"), PolicyDigest: os.Getenv("SOURCE_POLICY_DIGEST")}
}
