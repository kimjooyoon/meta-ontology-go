package generation

import (
	"strings"
	"testing"
)

func authorityBoundaryIR(resultEffects []string, resultSource string) SemanticOperationIR {
	source := SourceRevision{ID: "source-r1", Digest: envelopeDigestString("authority-boundary-source")}
	request := &EffectRequest{RequestID: "request-1", OperationID: "operation-1", SourceRevision: source.ID, Effects: []string{"read:source"}}
	return SemanticOperationIR{
		Schema:          SemanticOperationEnvelopeSchema,
		Intent:          OperationIntent{OperationID: "operation-1"},
		Source:          source,
		Grant:           EffectGrant{GrantID: "grant-1", ParentID: "root-grant", Effects: []string{"read:source"}},
		Request:         request,
		Result:          &EffectResult{ResultID: "result-1", RequestID: request.RequestID, SourceRevision: resultSource, Effects: resultEffects},
		AuthorityDigest: envelopeDigestString("authority-boundary-authority"),
	}
}

func TestObserveAuthorityBoundaryBindsRequestGrantResultAndSourceRevision(t *testing.T) {
	ir := authorityBoundaryIR([]string{"read:source"}, "source-r1")
	observation, err := ObserveAuthorityBoundary(ir)
	if err != nil {
		t.Fatalf("observe authority boundary: %v", err)
	}
	if observation.BoundaryStatus != "WITHIN_DECLARED_GRANT" || !observation.RequestWithinGrant || !observation.ResultWithinGrant || !observation.ResultSourceRevisionOK {
		t.Fatalf("unexpected boundary observation: %#v", observation)
	}
	if observation.RequestDigest == "" || observation.GrantDigest == "" || observation.ResultDigest == "" || observation.StableHash() == "" {
		t.Fatal("boundary observation did not preserve exact digests")
	}
	if err := observation.Validate(ir); err != nil {
		t.Fatalf("validate authority boundary observation: %v", err)
	}
}

func TestObserveAuthorityBoundaryDistinguishesEscalationAndStaleResult(t *testing.T) {
	escalated, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"write:repository"}, "source-r1"))
	if err != nil {
		t.Fatal(err)
	}
	if escalated.BoundaryStatus != "EFFECT_ESCALATION" || escalated.ResultWithinGrant {
		t.Fatalf("effect escalation was not observed: %#v", escalated)
	}
	stale, err := ObserveAuthorityBoundary(authorityBoundaryIR([]string{"read:source"}, "source-r2"))
	if err != nil {
		t.Fatal(err)
	}
	if stale.BoundaryStatus != "STALE_RESULT" || !stale.ResultWithinGrant || stale.ResultSourceRevisionOK {
		t.Fatalf("stale result was not observed: %#v", stale)
	}
	tampered := stale
	tampered.GrantDigest = strings.Repeat("0", 64)
	if err := tampered.Validate(authorityBoundaryIR([]string{"read:source"}, "source-r2")); err == nil {
		t.Fatal("tampered grant digest was accepted")
	}
}
