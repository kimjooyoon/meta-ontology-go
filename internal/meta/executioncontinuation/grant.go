package executioncontinuation

import (
	"context"
	"io/fs"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
	grant "github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementexecutiongrant"
)

// ValidateGrant reclassifies the supplied request under the Gooo grant contract.
func ValidateGrant(root fs.FS, g grant.LiveReport, e Envelope) error {
	if root == nil {
		return unknown("GRANT_POLICY_UNAVAILABLE")
	}
	unsigned := g
	unsigned.Digest = ""
	if g.Schema != grant.Schema || g.Digest != digestValue(unsigned) || g.DecisionSource != grant.DecisionSourceSystem || !reflect.DeepEqual(g.Metrics, g.Resolution.Metrics) {
		return refuted("GRANT_REPORT_INVALID")
	}
	program, err := grant.CompilePolicy(root, grant.PolicyPath)
	if err != nil {
		return unknown("GRANT_POLICY_UNAVAILABLE")
	}
	verification := grant.Verify(program, grant.GrantInput{Request: g.Request, Live: true}, g.Resolution)
	if !verification.Verified || !reflect.DeepEqual(verification, g.Verification) {
		return refuted("GRANT_VERIFICATION_MISMATCH")
	}
	if g.Request.Digest != e.RequestDigest {
		return refuted("GRANT_REQUEST_MISMATCH")
	}
	reason := policy.BindSubjects(e.SubjectSHA, g.Request.V24.SubjectSHA, g.Request.V25.SubjectSHA)
	if reason == "SUBJECT_UNKNOWN" {
		return unknown(reason)
	}
	if reason != "SUBJECT_BOUND" {
		return refuted(reason)
	}
	r := g.Resolution
	if !policy.PermitContinuation(string(r.Decision), string(r.Resolution), int64(r.RemainingUses), int64(r.ConsumedUses), int64(r.ExecutionCount), verification.Verified) {
		if r.Decision == grant.DecisionUnknown {
			return unknown("GRANT_ELIGIBILITY_UNKNOWN")
		}
		return refuted("GRANT_NOT_CONSUMABLE")
	}
	if grant.ValidateResolution(r) != nil || r.Receipt == nil || grant.ValidateGrantReceipt(*r.Receipt) != nil {
		return refuted("GRANT_RECEIPT_INVALID")
	}
	return nil
}

func (c Client) loadGrant(ctx context.Context, e Envelope, complete bool) (grant.LiveReport, Run, error) {
	run, err := c.observeRun(ctx, "grant", e.RunID, e.Attempt, complete)
	if err != nil {
		return grant.LiveReport{}, run, err
	}
	if run.Head != e.WorkflowSHA {
		return grant.LiveReport{}, run, refuted("GRANT_WORKFLOW_ORIGIN_MISMATCH")
	}
	files, err := c.readArtifact(ctx, run, policy.ArtifactLabel("grant", e.SubjectSHA), e.ArtifactID, e.ArchiveDigest, []string{"live-grant.json"})
	if err != nil {
		return grant.LiveReport{}, run, err
	}
	var report grant.LiveReport
	if decode(files["live-grant.json"], &report) != nil {
		return report, run, refuted("GRANT_REPORT_INVALID")
	}
	return report, run, nil
}
