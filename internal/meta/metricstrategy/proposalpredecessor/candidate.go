package proposalpredecessor

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func collectRun(ctx context.Context, client *http.Client, apiURL, token, repository, predecessorSHA string, run githubRun, collection *Collection) error {
	job, ready, err := collectSynthesisJob(
		ctx, client, apiURL, token, repository, run.ID, collection,
	)
	if err != nil {
		return err
	}
	if !ready {
		collection.Unresolved++
		return nil
	}
	artifactsURL := fmt.Sprintf("%s/repos/%s/actions/runs/%d/artifacts?per_page=100", strings.TrimRight(apiURL, "/"), repository, run.ID)
	var artifacts artifactsEnvelope
	if err := getJSON(ctx, client, artifactsURL, token, &artifacts); err != nil {
		return err
	}
	collection.ObservedArtifacts += len(artifacts.Artifacts)
	if artifacts.TotalCount != len(artifacts.Artifacts) {
		return &Failure{Reason: ReasonArtifactPaginationIncomplete}
	}
	for _, artifact := range artifacts.Artifacts {
		if artifact.Name != "metric-strategy-"+predecessorSHA {
			continue
		}
		collection.ExactArtifacts++
		candidate, err := collectArtifact(ctx, client, token, predecessorSHA, run, job, artifact)
		if err != nil {
			collection.Unresolved++
			if collection.FailureReason == "" {
				collection.FailureReason = FailureReason(err)
			}
			continue
		}
		collection.Candidates = append(collection.Candidates, candidate)
	}
	return nil
}

func collectArtifact(ctx context.Context, client *http.Client, token, predecessorSHA string, run githubRun, job githubJob, artifact githubArtifact) (Candidate, error) {
	if artifact.Expired || artifact.ID < 1 || artifact.ArchiveDownloadURL == "" {
		return Candidate{}, &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: fmt.Errorf("proposal predecessor artifact is unavailable")}
	}
	archive, err := getBytes(ctx, client, artifact.ArchiveDownloadURL, token)
	if err != nil {
		if FailureReason(err) != "" {
			return Candidate{}, err
		}
		return Candidate{}, &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: err}
	}
	payload, report, fileSHA, err := decodeProposalArchive(archive)
	if err != nil {
		return Candidate{}, err
	}
	if report.SubjectSHA != predecessorSHA || report.Decision != "PASS" || report.Reason != "CHANGE_PROPOSAL_CONTRACT_READY" {
		return Candidate{}, &Failure{Reason: ReasonArtifactPayloadUnavailable, Err: fmt.Errorf("proposal predecessor contract identity diverged")}
	}
	selected := Selected{
		RunID: run.ID, RunAttempt: run.RunAttempt, HeadSHA: run.HeadSHA,
		Event: run.Event, Status: run.Status, Conclusion: run.Conclusion, WorkflowName: run.Name,
		SynthesisJobID: job.ID, SynthesisJobName: job.Name,
		SynthesisJobStatus: job.Status, SynthesisJobConclusion: job.Conclusion,
		ArtifactID: artifact.ID, ArtifactName: artifact.Name, ProposalFileSHA256: fileSHA,
		ProposalReportDigest: report.ReportDigest, ContractSatisfied: report.Summary.Satisfied,
		ContractTotal: report.Summary.Total, ContractBPS: report.Summary.ReadinessBPS,
		ContractUnresolved: report.Summary.Unresolved, RepositoryWrites: report.RepositoryWrites,
		PromotionAuthorized: report.PromotionAuthorized,
	}
	return Candidate{Selected: selected, ProposalPayload: payload}, nil
}
