package guardedpromotion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func (collector *Collector) runArtifacts(
	ctx context.Context, repositoryPath string, runID int64,
) ([]artifactResponse, error) {
	path := fmt.Sprintf("%s/actions/runs/%d/artifacts?per_page=100", repositoryPath, runID)
	var response artifactsResponse
	if err := collector.getJSON(ctx, path, &response); err != nil {
		return nil, err
	}
	return response.Artifacts, nil
}

func (collector *Collector) bindCandidate(
	ctx context.Context, repositoryPath string, source *Source, candidate promotionCandidate,
) error {
	path := fmt.Sprintf("%s/actions/artifacts/%d/zip", repositoryPath, candidate.artifact.ID)
	archive, err := collector.get(ctx, path)
	if err != nil {
		return err
	}
	data, err := promotionJSON(archive)
	if err != nil {
		return err
	}
	envelope, err := decodePromotion(data, source.PredecessorSHA)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	source.Artifact = ArtifactEvidence{
		RunID: candidate.run.ID, RunAttempt: candidate.run.RunAttempt,
		RunEvent: candidate.run.Event, ArtifactID: candidate.artifact.ID,
		ArtifactName: candidate.artifact.Name, ArtifactDigest: candidate.artifact.Digest,
		FileSHA256: "sha256:" + hex.EncodeToString(sum[:]), ReportSchema: envelope.Schema,
		ReportDigest: envelope.ReportDigest, ReportCurrentHeadSHA: envelope.CurrentHeadSHA,
		ReportDecision: envelope.Decision, ReportSatisfied: envelope.Summary.Satisfied,
		ReportTotal: envelope.Summary.Total, ReportUnresolved: envelope.Summary.Unresolved,
		ReportRepositoryWrites: envelope.Summary.RepositoryWrites,
	}
	return nil
}
