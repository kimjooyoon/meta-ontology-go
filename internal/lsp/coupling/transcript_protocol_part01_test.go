package coupling

import (
	"reflect"
	"testing"
)

func TestProtocolTranscriptUpstreamUnknownAndFailClosed(t *testing.T) {
	tests := []struct {
		name        string
		status      Outcome
		reason      Reason
		wantOutcome Outcome
		wantCode    string
	}{
		{name: "unknown", status: OutcomeUnknown, reason: ReasonUpstreamUnknown, wantOutcome: OutcomeUnknown, wantCode: DiagnosticUpstreamUnknown},
		{name: "fail-closed", status: OutcomeFailClosed, reason: ReasonUpstreamFail, wantOutcome: OutcomeFailClosed, wantCode: DiagnosticUpstreamFail},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
			envelope.Status, envelope.Reason = test.status, test.reason
			result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
			if result.Outcome != test.wantOutcome || len(result.Links) != 0 || result.Hover != nil {
				t.Fatalf("upstream result = %#v", result)
			}
			if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != test.wantCode {
				t.Fatalf("upstream diagnostics = %#v", result.Diagnostics)
			}
		})
	}
}
func TestProtocolTranscriptRootRelocationAndLabelChanges(t *testing.T) {
	oldEnvelope := transcriptEnvelope("file:///old-root/main.go", "Old label", ClaimNoDelta)
	oldEnvelope.Explanations[0].Target.URI = "file:///old-root/model.gooo"
	oldEvidence, err := ComputeEvidenceDigest(oldEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	old := transcriptAdapter(t, oldEnvelope).Resolve(transcriptRequest(oldEnvelope))

	newEnvelope := transcriptEnvelope("file:///new-root/main.go", "Renamed label", ClaimNoDelta)
	newEnvelope.Explanations[0].Target.URI = "file:///new-root/model.gooo"
	newEvidence, err := ComputeEvidenceDigest(newEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	if oldEvidence != newEvidence {
		t.Fatalf("presentation-only root relocation changed evidence digest: %q != %q", oldEvidence, newEvidence)
	}
	new := transcriptAdapter(t, newEnvelope).Resolve(transcriptRequest(newEnvelope))
	if old.Outcome != OutcomePass || new.Outcome != OutcomePass {
		t.Fatalf("relocation results = %#v %#v", old, new)
	}
	if old.Links[0].TargetURI == new.Links[0].TargetURI || new.Links[0].TargetURI != "file:///new-root/model.gooo" {
		t.Fatalf("root relocation target = %q from %q", new.Links[0].TargetURI, old.Links[0].TargetURI)
	}

	labelEnvelope := transcriptEnvelope("file:///old-root/main.go", "Renamed label", ClaimNoDelta)
	labelEnvelope.Explanations[0].Target.URI = oldEnvelope.Explanations[0].Target.URI
	labelResult := transcriptAdapter(t, labelEnvelope).Resolve(transcriptRequest(labelEnvelope))
	if !reflect.DeepEqual(old.Links, labelResult.Links) {
		t.Fatalf("label change altered exact navigation: old=%#v new=%#v", old.Links, labelResult.Links)
	}
	if old.Hover.Contents.Value == labelResult.Hover.Contents.Value {
		t.Fatal("label change did not alter presentation hover")
	}
}
