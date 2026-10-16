package executioncontinuation

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/executioncontinuation/policy"
)

type Prepared struct {
	Pair        Pair     `json:"pair"`
	Metadata    Metadata `json:"metadata"`
	SourceBytes string   `json:"subject_source"`
}

func (c Client) Prepare(ctx context.Context, e Envelope, consumerRun int64, attempt int) (Prepared, error) {
	if err := e.Validate(); err != nil {
		return Prepared{}, refuted(err.Error())
	}
	if e.Repository != c.Repository {
		return Prepared{}, refuted("SOURCE_REPOSITORY_MISMATCH")
	}
	own, err := c.observeRun(ctx, "consumer", consumerRun, attempt, false)
	if err != nil {
		return Prepared{}, err
	}
	if policy.BindTargetRun(e.RunID, consumerRun, own.ID) != "TARGET_RUN_BOUND" {
		return Prepared{}, refuted("TARGET_RUN_REFUTED")
	}
	g, _, err := c.loadGrant(ctx, e, true)
	if err != nil {
		return Prepared{}, err
	}
	if err := ValidateGrant(c.Root, g, e); err != nil {
		return Prepared{}, err
	}
	source := g.Request.Source
	run, err := c.observeRun(ctx, "contract", source.WorkflowRunID, source.WorkflowRunAttempt, true)
	if err != nil {
		return Prepared{}, err
	}
	files, err := c.readArtifact(ctx, run, policy.ArtifactLabel("contract", e.SubjectSHA), source.ArtifactID, source.ArtifactDigest,
		[]string{"live-contract.json", "source-authorization-request.json", "source-authorization-resolution.json"})
	if err != nil {
		return Prepared{}, err
	}
	p := Prepared{Pair: Pair{Grant: g}, Metadata: Metadata{Source: e, ConsumerRunID: consumerRun, ConsumerAttempt: attempt, ConsumerWorkflowSHA: own.Head,
		ContractRunID: run.ID, ContractAttempt: run.Attempt, ContractArtifactID: source.ArtifactID, ContractDigest: source.ArtifactDigest, ContractWorkflowSHA: run.Head, PolicyDigest: policy.SourceDigest()}}
	if decode(files["live-contract.json"], &p.Pair.Contract) != nil || decode(files["source-authorization-request.json"], &p.Pair.AuthorizationRequest) != nil || decode(files["source-authorization-resolution.json"], &p.Pair.AuthorizationResolution) != nil {
		return p, refuted("CONTRACT_BUNDLE_INVALID")
	}
	input := p.Pair.Contract.ExecutionInput
	if input == nil {
		return p, refuted("EXECUTION_INPUT_MISSING")
	}
	content, blob, err := c.sourceBytes(ctx, e.SubjectSHA, input.Source.Path)
	if err != nil {
		return p, err
	}
	p.SourceBytes = string(content)
	p.Metadata.SubjectBlobSHA = blob
	p.Metadata.SubjectFileDigest = Digest(content)
	p.Metadata.Digest = digestValue(p.Metadata)
	if _, err := ValidatePair(c.Root, p.Pair, p.Metadata); err != nil {
		return p, err
	}
	if input.Source.Bytes != p.SourceBytes || input.Source.Digest != p.Metadata.SubjectFileDigest {
		return p, refuted("SUBJECT_SOURCE_MISMATCH")
	}
	return p, nil
}

// HasReservation searches complete retained history without a subject/head filter.
func (c Client) HasReservation(ctx context.Context, e Envelope) (bool, error) {
	name := ReservationName(e.RequestDigest)
	total, observed := -1, 0
	seen := map[int64]bool{}
	for page := 1; page <= 8; page++ {
		var part struct {
			Total *int  `json:"total_count"`
			Runs  []Run `json:"workflow_runs"`
		}
		if err := c.get(ctx, fmt.Sprintf("actions/workflows/self-improvement-execution-consumer.yml/runs?branch=dev&per_page=100&page=%d", page), &part); err != nil {
			return false, unknown("RESERVATION_HISTORY_UNKNOWN")
		}
		if part.Total == nil || part.Runs == nil || *part.Total < 0 || len(part.Runs) > 100 ||
			(total >= 0 && *part.Total != total) {
			return false, unknown("RESERVATION_HISTORY_UNKNOWN")
		}
		total = *part.Total
		observed += len(part.Runs)
		for _, run := range part.Runs {
			if seen[run.ID] {
				return false, unknown("RESERVATION_HISTORY_AMBIGUOUS")
			}
			seen[run.ID] = true
			as, err := c.artifacts(ctx, run.ID)
			if err != nil {
				return false, err
			}
			for _, a := range as {
				if a.Name == name {
					return c.validReservation(ctx, run, a, e)
				}
			}
		}
		if observed == total {
			return false, nil
		}
		if observed > total {
			return false, unknown("RESERVATION_HISTORY_AMBIGUOUS")
		}
	}
	return false, unknown("RESERVATION_HISTORY_INCOMPLETE")
}
