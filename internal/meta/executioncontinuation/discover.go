package executioncontinuation

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
)

// DiscoverGrant binds the supplied logical subject to the current producer.
func (c Client) DiscoverGrant(ctx context.Context, e Envelope) (Envelope, error) {
	e.Repository = c.Repository
	e.PolicyDigest = policy.SourceDigest()
	e.Wait = "CREATED"
	if !validSHA(e.SubjectSHA) || !validSHA(e.WorkflowSHA) || e.RunID <= 0 || e.Attempt != 1 {
		return e, refuted("SOURCE_DECLARATION_INVALID")
	}
	run, err := c.observeRun(ctx, "grant", e.RunID, e.Attempt, false)
	if err != nil {
		return e, err
	}
	if run.Head != e.WorkflowSHA {
		return e, refuted("SOURCE_WORKFLOW_MISMATCH")
	}
	rows, err := c.artifacts(ctx, e.RunID)
	if err != nil {
		return e, err
	}
	matches := 0
	for _, a := range rows {
		if a.Name == policy.ArtifactLabel("grant", e.SubjectSHA) {
			e.ArtifactID = a.ID
			e.ArchiveDigest = a.Digest
			matches++
		}
	}
	if matches == 0 {
		return e, unknown("SOURCE_ARTIFACT_MISSING")
	}
	if matches != 1 {
		return e, refuted("SOURCE_ARTIFACT_DUPLICATED")
	}
	files, err := c.readArtifact(ctx, run, policy.ArtifactLabel("grant", e.SubjectSHA), e.ArtifactID, e.ArchiveDigest, []string{"live-grant.json"})
	if err != nil {
		return e, err
	}
	var g grant.LiveReport
	if decode(files["live-grant.json"], &g) != nil {
		return e, refuted("GRANT_REPORT_INVALID")
	}
	e.RequestDigest = g.Request.Digest
	if err := e.Validate(); err != nil {
		return e, refuted(err.Error())
	}
	return e, ValidateGrant(c.Root, g, e)
}
