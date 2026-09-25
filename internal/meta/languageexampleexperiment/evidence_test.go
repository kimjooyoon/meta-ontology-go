package languageexampleexperiment

import "testing"

func TestEvaluateRejectsCopiedDigestOnChangedArtifact(t *testing.T) {
	input := validInput()
	input.Artifact.Operation.Activity = "Other"
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "EXACT" ||
		report.Reason != "ARTIFACT_DIGEST_INVALID" {
		t.Fatalf("report=%#v", report)
	}
}

func TestEvaluateRejectsValidDivergentReplay(t *testing.T) {
	input := validInput()
	input.Replay = fixtureArtifact("CancelOrder")
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Reason != "ARTIFACT_REPLAY_MISMATCH" {
		t.Fatalf("report=%#v", report)
	}
}

func TestEvaluateRejectsImpossibleResourceSample(t *testing.T) {
	input := validInput()
	input.Profile.Samples[0].RSSKiB = -1
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Reason != "PROFILE_SAMPLE_INVALID" {
		t.Fatalf("report=%#v", report)
	}
}

func TestEvaluateDistinguishesKnownFailureFromUnknown(t *testing.T) {
	input := validInput()
	input.Artifact = input.UnknownEmitter
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "EXACT" ||
		report.Reason != "ARTIFACT_DECISION_REJECTED" {
		t.Fatalf("report=%#v", report)
	}
}
