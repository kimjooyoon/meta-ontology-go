package packageexecution

import "testing"

func TestReplayAcceptsUnchangedSealedArtifact(t *testing.T) {
	request := replayPackageRequest()
	artifact := Execute(request)
	replayed := Replay(request, artifact)
	if err := ValidateReplay(replayed); err != nil {
		t.Fatalf("replay receipt was not sealed: %v", err)
	}
	if replayed.Decision != "PASS" || replayed.Reason != "PACKAGE_ARTIFACT_REPLAYED" {
		t.Fatalf("unexpected replay result: %#v", replayed)
	}
}

func TestReplayRejectsChangedSourceIdentity(t *testing.T) {
	request := replayPackageRequest()
	artifact := Execute(request)
	request.Sources[1].Content += "\n"
	replayed := Replay(request, artifact)
	if err := ValidateReplay(replayed); err != nil {
		t.Fatalf("rejection receipt was not sealed: %v", err)
	}
	if replayed.Decision != "FAIL_CLOSED" || replayed.Reason != "PACKAGE_ARTIFACT_SOURCE_DIGEST_MISMATCH" {
		t.Fatalf("changed source was accepted: %#v", replayed)
	}
}

func TestReplayRejectsTamperedArtifact(t *testing.T) {
	request := replayPackageRequest()
	artifact := Execute(request)
	artifact.Digest = "sha256:tampered"
	replayed := Replay(request, artifact)
	if err := ValidateReplay(replayed); err != nil {
		t.Fatalf("tamper rejection receipt was not sealed: %v", err)
	}
	if replayed.Decision != "FAIL_CLOSED" || replayed.Reason != "PACKAGE_ARTIFACT_INVALID" {
		t.Fatalf("tampered artifact was accepted: %#v", replayed)
	}
}

func TestReplayRejectsMissingArtifact(t *testing.T) {
	replayed := Replay(replayPackageRequest(), Receipt{})
	if err := ValidateReplay(replayed); err != nil {
		t.Fatalf("missing-artifact receipt was not sealed: %v", err)
	}
	if replayed.Decision != "FAIL_CLOSED" || replayed.Reason != "PACKAGE_ARTIFACT_INVALID" {
		t.Fatalf("missing artifact was accepted: %#v", replayed)
	}
}

func replayPackageRequest() Request {
	return Request{
		PackagePath: "billing",
		Entry:       "PayOrder",
		Sources: []Source{
			{
				Filename: "activity.gooo",
				Content:  "package billing\nnamespace billing\n\nactivity PayOrder(Order) -> Receipt\n",
			},
			{
				Filename: "entities.gooo",
				Content:  "package billing\nnamespace billing\n\nentity Order id \"urn:gooo:billing:order\"\nentity Receipt id \"urn:gooo:billing:receipt\"\n",
			},
		},
	}
}
