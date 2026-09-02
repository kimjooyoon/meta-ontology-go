package selfimprovementtransport

import (
	"fmt"
	"strings"
	"testing"
)

const selectionHead = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func selectionRun(attempt int) []byte {
	return []byte(fmt.Sprintf(`{"id":41,"run_attempt":%d,"head_sha":"%s","path":".github/workflows/observation.yml"}`,
		attempt, selectionHead))
}

func selectionArtifacts(ids ...int64) []byte {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, fmt.Sprintf(`{"id":%d,"name":"%s","expired":false,"digest":"sha256:%s","size_in_bytes":2953,"workflow_run":{"id":41,"head_sha":"%s"}}`,
			id, ArtifactName, strings.Repeat("b", 64), selectionHead))
	}
	return []byte(`{"artifacts":[` + strings.Join(values, ",") + `]}`)
}

func TestSelectTransportMetadataBindsTopLevelRunAttempt(t *testing.T) {
	metadata, err := SelectTransportMetadata(selectionRun(2), selectionArtifacts(99),
		ArtifactSelectionInput{Repository: "owner/repo", ExpectedRunID: 41,
			ExpectedRunAttempt: 2, ArtifactName: ArtifactName})
	if err != nil {
		t.Fatal(err)
	}
	if metadata.ProducerRunAttempt != 2 || metadata.ArtifactID != 99 ||
		metadata.OrchestrationHeadSHA != selectionHead {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestSelectTransportMetadataRejectsTopLevelAttemptMismatch(t *testing.T) {
	_, err := SelectTransportMetadata(selectionRun(3), selectionArtifacts(99),
		ArtifactSelectionInput{Repository: "owner/repo", ExpectedRunID: 41,
			ExpectedRunAttempt: 2, ArtifactName: ArtifactName})
	if err == nil || !strings.Contains(err.Error(), "LOCATE/verify-run-attempt/SOURCE_RUN_ATTEMPT_MISMATCH") {
		t.Fatalf("error = %v", err)
	}
}
