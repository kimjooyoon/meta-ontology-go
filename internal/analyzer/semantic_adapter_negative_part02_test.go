package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestSemanticAdapterRejectsReversedTypedEndpoint(t *testing.T) {
	analysis := Result{Registrations: []Registration{
		registration(KindEntity, "billing://entity/order", "Order"),
		registration(KindActivity, "billing://activity/pay-order", "PayOrder"),
	}, Delta: SemanticDelta{Added: []Fact{{
		Subject: NewIdentity("billing", "billing://entity/order"), Relation: RelationUses,
		Object: NewIdentity("billing", "billing://activity/pay-order"), Span: testSpan(),
	}}}}
	_, err := AdaptSemantic(SemanticAdapterInput{
		Base: semantic.NewIR("billing", semantic.Namespace("billing")), Analysis: analysis,
		Policy: billingPolicy(t, RelationUses), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, SourceDigest: semantic.StableHash([]byte("reverse")),
	})
	assertAdapterCode(t, err, AdapterEndpointKind)
}
