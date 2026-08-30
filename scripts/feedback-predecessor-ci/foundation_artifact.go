package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
)

func verifyFoundationArtifact(ctx context.Context, client *githubClient, repository string, runID int64, expected feedbackpredecessor.FoundationEvidence) error {
	var artifacts artifactList
	endpoint := fmt.Sprintf("/repos/%s/actions/runs/%d/artifacts?per_page=100", repository, runID)
	if err := client.getJSON(ctx, endpoint, &artifacts); err != nil {
		return err
	}
	if artifacts.TotalCount != len(artifacts.Artifacts) {
		return fmt.Errorf("foundation artifact pagination is incomplete")
	}
	selected, ok := exactFoundationArtifact(artifacts.Artifacts, expected)
	if !ok {
		return fmt.Errorf("last-known-good artifact selection is not unique")
	}
	archive, err := client.get(ctx, fmt.Sprintf("/repos/%s/actions/artifacts/%d/zip", repository, selected.ID))
	if err != nil {
		return err
	}
	if archiveDigest(archive) != selected.Digest {
		return fmt.Errorf("downloaded last-known-good artifact digest is not exact")
	}
	decoded := decodeReceipt(archive)
	if decoded.Receipt.ReceiptDigest != expected.LastKnownGoodReceiptDigest ||
		decoded.Receipt.RepositoryWrites != 0 || decoded.Receipt.Decision != "FIXED_POINT" ||
		decoded.Receipt.Report.Decision != "FIXED_POINT" ||
		decoded.Receipt.Report.Feedback.CommitSHA != expected.LastKnownGoodSHA ||
		decoded.Receipt.Report.Feedback.Decision != "FIXED_POINT" {
		return fmt.Errorf("last-known-good receipt is not bound to a fixed point")
	}
	return nil
}

func exactFoundationArtifact(artifacts []artifact, expected feedbackpredecessor.FoundationEvidence) (artifact, bool) {
	var selected artifact
	matches := 0
	for _, candidate := range artifacts {
		if candidate.Name != expected.LastKnownGoodArtifactName {
			continue
		}
		selected, matches = candidate, matches+1
	}
	if matches != 1 || selected.ID != expected.LastKnownGoodArtifactID || selected.Expired ||
		selected.SizeInBytes <= 0 || selected.Digest != expected.LastKnownGoodArtifactDigest {
		return artifact{}, false
	}
	return selected, true
}

func archiveDigest(archive []byte) string {
	sum := sha256.Sum256(archive)
	return "sha256:" + hex.EncodeToString(sum[:])
}
