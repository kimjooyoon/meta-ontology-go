package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestCandidateOnlyObservationCannotCloseReceipt(t *testing.T) {
	fixture := newFixture(t, ChangeClaimDelta)
	authority := fixture.input.InferencePath.Edges[0]
	authority.Kind = semantic.InferenceObservationCandidate
	authority.Authority.Layer = semantic.AuthorityCandidate
	authority.Authority.Effect = semantic.AuthorityObserve
	authority.SourceRoots = nil
	authority.Phase.Phase = semantic.PhaseObservation
	authority.After.Semantic = authority.Before.Semantic
	authority.Controls.CatalogDigest = fixtureDigest("catalog")
	fixture.input.InferencePath.Edges[0] = authority
	fixture.input.InferencePath.Evidence[0].After.Semantic = authority.After.Semantic
	fixture.input.InferencePath.Evidence[0].Controls = authority.Controls
	result := Evaluate(fixture.input, fixture.authorityContext)
	if result.Status != StatusFailClosed || !reflect.DeepEqual(result.Reasons[0].Code, ReasonCandidateOnlyPath) {
		t.Fatalf("candidate-only result = %#v", result)
	}
}
func TestAcceptedLiftClosesAChangedSurface(t *testing.T) {
	fixture := newFixture(t, ChangeClaimDelta)
	authority := fixture.input.InferencePath.Edges[0]
	authority.Kind = semantic.InferenceAcceptedLift
	authority.Phase.Phase = semantic.PhaseLift
	authority.Authority.Layer = semantic.AuthoritySemantic
	authority.Authority.Effect = semantic.AuthorityLift
	authority.Controls.PolicyDigest = fixture.input.Config.ProfileDigest
	authority.AcceptanceReceipt = fixture.input.InferencePath.Evidence[0].ID
	fixture.input.InferencePath.Edges[0] = authority
	fixture.input.InferencePath.Evidence[0].Controls = authority.Controls
	result := Evaluate(fixture.input, fixture.authorityContext)
	if result.Status != StatusPass {
		t.Fatalf("accepted-lift result = %#v", result)
	}
}
