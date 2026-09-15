package selfimprovementtransport

import (
	"strings"
	"testing"
)

func TestSelectTransportMetadataRejectsAmbiguousArtifact(t *testing.T) {
	_, err := SelectTransportMetadata(selectionRun(1), selectionArtifacts(98, 99),
		ArtifactSelectionInput{Repository: "owner/repo", ExpectedRunID: 41,
			ExpectedRunAttempt: 1, ArtifactName: ArtifactName})
	if err == nil || !strings.Contains(err.Error(), "LOCATE/select-immutable-artifact/ARTIFACT_SELECTION_AMBIGUOUS") {
		t.Fatalf("error = %v", err)
	}
}

func TestSelectTransportMetadataPreservesMissingArtifactCoordinate(t *testing.T) {
	_, err := SelectTransportMetadata(selectionRun(1), selectionArtifacts(),
		ArtifactSelectionInput{Repository: "owner/repo", ExpectedRunID: 41,
			ExpectedRunAttempt: 1, ArtifactName: ArtifactName})
	if err == nil || !strings.Contains(err.Error(), "LOCATE/select-immutable-artifact/ARTIFACT_NOT_FOUND") {
		t.Fatalf("error = %v", err)
	}
}
