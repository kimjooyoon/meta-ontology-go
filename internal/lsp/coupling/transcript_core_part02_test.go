package coupling

import (
	"strings"
	"testing"
)

func TestProtocolTranscriptDeltaAndNoDelta(t *testing.T) {
	for _, claim := range []ChangeClaim{ClaimDelta, ClaimNoDelta} {
		t.Run(string(claim), func(t *testing.T) {
			envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", claim)
			result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
			if result.Outcome != OutcomePass || len(result.Links) != 1 || result.Hover == nil {
				t.Fatalf("result = %#v, want one standard link and hover", result)
			}
			if got := result.Links[0].TargetURI; got != "file:///workspace/model.gooo" {
				t.Fatalf("target URI = %q", got)
			}
			if !strings.Contains(result.Hover.Contents.Value, string(claim)) {
				t.Fatalf("hover = %#v, missing claim %q", result.Hover, claim)
			}
			if len(result.Diagnostics) != 1 || len(result.Diagnostics[0].RelatedInformation) != 2 {
				t.Fatalf("diagnostics = %#v, want two causal spans", result.Diagnostics)
			}
			if got := result.Diagnostics[0].RelatedInformation[0].Message; got != "registered binding" {
				t.Fatalf("first causal message = %q, ordinal order was not preserved", got)
			}
		})
	}
}
func TestProtocolTranscriptAmbiguityHasNoNavigation(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	duplicate := envelope.Explanations[0]
	duplicate.CodeSymbolID = "stable://code/other"
	duplicate.SemanticOwnerID = "stable://activity/other"
	duplicate.Origin.StableID = "stable://span/other-origin"
	duplicate.Origin.SourceMapID = "stable://map/other-code"
	duplicate.Origin.SourceMapDigest = transcriptDigest("source-map-other-code")
	duplicate.Origin.Range.End.Character = 21
	envelope.Explanations = append(envelope.Explanations, duplicate)
	result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Hover != nil {
		t.Fatalf("ambiguous result = %#v, want no navigation", result)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != DiagnosticAmbiguous || len(result.Diagnostics[0].RelatedInformation) != 0 {
		t.Fatalf("ambiguous diagnostics = %#v", result.Diagnostics)
	}
}
func TestProtocolTranscriptStaleSnapshotAndWrongVersion(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimNoDelta)
	adapter := transcriptAdapter(t, envelope)
	stale := transcriptRequest(envelope)
	stale.SnapshotDigest = transcriptDigest("new-snapshot")
	result := adapter.Resolve(stale)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticStaleSnapshot {
		t.Fatalf("stale result = %#v", result)
	}
	wrongVersion := transcriptRequest(envelope)
	wrongVersion.DocumentVersion++
	result = adapter.Resolve(wrongVersion)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticWrongVersion {
		t.Fatalf("wrong-version result = %#v", result)
	}
}
