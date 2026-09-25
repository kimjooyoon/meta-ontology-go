package selfimprovementtransport

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

func lifecycleInput() ArtifactLifecycleInput {
	return ArtifactLifecycleInput{Selection: ArtifactSelectionInput{Repository: "owner/repo", ExpectedRunID: 41, ExpectedRunAttempt: 1, ArtifactName: ArtifactName}}
}

func lifecycleArtifacts(digest string, expired bool, id, runID int64, head string) []byte {
	return []byte(fmt.Sprintf(`{"artifacts":[{"id":%d,"name":"%s","expired":%t,"digest":"%s","size_in_bytes":2953,"workflow_run":{"id":%d,"head_sha":"%s"}}]}`, id, ArtifactName, expired, digest, runID, head))
}

func observeLifecycle(runRaw, artifactsRaw []byte, input ArtifactLifecycleInput) (TransportMetadata, LifecycleReceipt) {
	repository := fstest.MapFS{"transport.gooo": {Data: []byte(contractFixture)}}
	return ObserveArtifactLifecycle(repository, "transport.gooo", runRaw, artifactsRaw, input)
}

func TestArtifactLifecycleExactReplay(t *testing.T) {
	archive := []byte("immutable archive")
	digest := digestBytes(archive)
	metadata, located := observeLifecycle(selectionRun(1), lifecycleArtifacts(digest, false, 99, 41, selectionHead), lifecycleInput())
	if metadata.ArtifactID != 99 || located.Metrics.VerifiedTotal != 3 {
		t.Fatalf("located = %+v metadata = %+v", located, metadata)
	}
	first := CompleteArtifactLifecycle(located, archive, 0)
	second := CompleteArtifactLifecycle(located, archive, 0)
	if err := ValidateArtifactLifecycleReceipt(first); err != nil {
		t.Fatal(err)
	}
	if first.Digest != second.Digest || first.Decision != DecisionPass || first.Metrics.VerifiedTotal != 5 || first.Metrics.DischargedTotal != 5 {
		t.Fatalf("replay mismatch: first=%+v second=%+v", first, second)
	}
}

func TestArtifactLifecycleLocateFailuresAreStageSpecific(t *testing.T) {
	validDigest := "sha256:" + strings.Repeat("b", 64)
	cases := []struct {
		name, reason   string
		run, artifacts []byte
		input          ArtifactLifecycleInput
	}{{"input", "ARTIFACT_LIFECYCLE_INPUT_INVALID", selectionRun(1), selectionArtifacts(99), ArtifactLifecycleInput{}}, {"lookup", "ARTIFACT_METADATA_LOOKUP_FAILED", selectionRun(1), selectionArtifacts(99), func() ArtifactLifecycleInput { value := lifecycleInput(); value.RunLookupExit = 1; return value }()}, {"invalid-response", "ARTIFACT_METADATA_RESPONSE_INVALID", selectionRun(1), []byte(`{}`), lifecycleInput()}, {"missing", "ARTIFACT_NOT_FOUND", selectionRun(1), []byte(`{"artifacts":[]}`), lifecycleInput()}, {"ambiguous", "ARTIFACT_SELECTION_AMBIGUOUS", selectionRun(1), selectionArtifacts(98, 99), lifecycleInput()}, {"expired", "ARTIFACT_EXPIRED", selectionRun(1), lifecycleArtifacts(validDigest, true, 99, 41, selectionHead), lifecycleInput()}, {"invalid-metadata", "ARTIFACT_METADATA_INVALID", selectionRun(1), lifecycleArtifacts("bad", false, 99, 41, selectionHead), lifecycleInput()}, {"run-attempt", "ARTIFACT_RUN_BINDING_MISMATCH", selectionRun(2), selectionArtifacts(99), lifecycleInput()}, {"artifact-run", "ARTIFACT_RUN_BINDING_MISMATCH", selectionRun(1), lifecycleArtifacts(validDigest, false, 99, 40, selectionHead), lifecycleInput()}}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, receipt := observeLifecycle(test.run, test.artifacts, test.input)
			if err := ValidateArtifactLifecycleReceipt(receipt); err != nil {
				t.Fatal(err)
			}
			if receipt.Decision != DecisionFailClosed || receipt.Reason != test.reason || receipt.Metrics.UnknownPathCount != 1 || receipt.Authority != (LifecycleAuthority{}) {
				t.Fatalf("receipt = %+v", receipt)
			}
		})
	}
}

func TestArtifactLifecycleTransportFailuresAreNoEffect(t *testing.T) {
	archive := []byte("archive")
	_, located := observeLifecycle(selectionRun(1), lifecycleArtifacts(digestBytes(archive), false, 99, 41, selectionHead), lifecycleInput())
	download := CompleteArtifactLifecycle(located, nil, 22)
	if download.Reason != "ARTIFACT_DOWNLOAD_FAILED" || download.EnforcementEffect != LifecycleEffectNoEffect {
		t.Fatalf("download = %+v", download)
	}
	mismatch := CompleteArtifactLifecycle(located, []byte("different"), 0)
	if mismatch.Reason != "ARCHIVE_DIGEST_MISMATCH" || mismatch.Indicators[4].ObservationClass != "CONTRADICTION" || mismatch.Metrics.VerifiedTotal != 4 {
		t.Fatalf("mismatch = %+v", mismatch)
	}
}
