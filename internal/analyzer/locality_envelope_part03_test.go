package analyzer

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func localityBase(t *testing.T, includeUnrelated bool) semantic.IR {
	t.Helper()
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	activity := mustBillingNode(t, semantic.Activity, "billing://activity/pay-order", "PayOrder")
	order := mustBillingNode(t, semantic.Entity, "billing://entity/order", "Order")
	for _, node := range []semantic.Node{activity, order} {
		if err := base.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := base.AddFact(semantic.NewUsedFact(activity.ID, order.ID)); err != nil {
		t.Fatal(err)
	}
	if !includeUnrelated {
		return base
	}
	unrelatedActivity := mustBillingNode(t, semantic.Activity, "billing://activity/unrelated", "Unrelated")
	unrelatedEntity := mustBillingNode(t, semantic.Entity, "billing://entity/unrelated", "UnrelatedEntity")
	for _, node := range []semantic.Node{unrelatedActivity, unrelatedEntity} {
		if err := base.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := base.AddFact(semantic.NewUsedFact(unrelatedActivity.ID, unrelatedEntity.ID)); err != nil {
		t.Fatal(err)
	}
	return base
}
func adaptGeneratedBillingWithBase(t *testing.T, base semantic.IR, registry *Registry) SemanticAdapterResult {
	t.Helper()
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{generatedBillingSource(t)}, Registry: registry,
		Policy: generatedBillingPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: billingToolchain(),
	})
	if err != nil {
		t.Fatalf("adapt generated billing locality fixture: %v", err)
	}
	return result
}
func containsLocalityID(ids []semantic.ID, want semantic.ID) bool {
	return slices.Contains(ids, want)
}
