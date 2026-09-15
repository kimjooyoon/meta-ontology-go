package coupling

import (
	"context"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
	"testing"
)

func TestLiveQueryDecisionAndMutationMatrixNeverLinks(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*couplingexplain.VerifiedEnvelope, *LiveRequest)
	}{
		{name: "decision", mutate: func(envelope *couplingexplain.VerifiedEnvelope, _ *LiveRequest) {
			envelope.Verdict = couplingexplain.VerdictUnknown
			envelope.NoLinkReason = couplingexplain.ReasonMissing
			envelope.Diagnostics = []couplingexplain.Diagnostic{{Code: "missing-live-term-path-verifier"}}
		}},
		{name: "reason", mutate: func(envelope *couplingexplain.VerifiedEnvelope, _ *LiveRequest) {
			envelope.NoLinkReason = couplingexplain.ReasonAmbiguous
		}},
		{name: "digest", mutate: func(envelope *couplingexplain.VerifiedEnvelope, _ *LiveRequest) {
			envelope.Binding.SnapshotDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}},
		{name: "path", mutate: func(envelope *couplingexplain.VerifiedEnvelope, _ *LiveRequest) {
			envelope.OriginPath.Steps[1].ToID = "term://other"
		}},
		{name: "term", mutate: func(envelope *couplingexplain.VerifiedEnvelope, _ *LiveRequest) {
			envelope.Term.TermID = "term://other"
		}},
		{name: "evidence", mutate: func(envelope *couplingexplain.VerifiedEnvelope, _ *LiveRequest) {
			envelope.Verifier.EvidenceDigest = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		}},
		{name: "version", mutate: func(_ *couplingexplain.VerifiedEnvelope, request *LiveRequest) {
			request.DocumentVersion = 8
			request.Locations.DocumentVersion = 8
		}},
		{name: "cancellation", mutate: func(_ *couplingexplain.VerifiedEnvelope, request *LiveRequest) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			request.Context = ctx
		}},
		{name: "out-of-order", mutate: func(_ *couplingexplain.VerifiedEnvelope, request *LiveRequest) {
			request.Query.Control.ObservedVersion = 6
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			envelope, err := couplingexplain.DecodeVerifiedEnvelope([]byte(literalVerifiedQueryEnvelope))
			if err != nil {
				t.Fatal(err)
			}
			request := liveRequest()
			test.mutate(&envelope, &request)
			data, err := json.Marshal(envelope)
			if err != nil {
				t.Fatal(err)
			}
			if test.name != "version" && test.name != "cancellation" && test.name != "out-of-order" {
				envelope.EnvelopeDigest = envelope.Digest()
				data, err = json.Marshal(envelope)
				if err != nil {
					t.Fatal(err)
				}
			}
			result := ResolveLive(request, data)
			if len(result.Links) != 0 || result.Hover != nil || len(result.Diagnostics) != 1 {
				t.Fatalf("mutation produced navigation: %#v", result)
			}
		})
	}
}
