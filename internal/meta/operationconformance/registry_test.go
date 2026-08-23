package operationconformance_test

import (
	"slices"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/operationconformance"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestGenerationRegistryUsesSplitGoV1Denominator(t *testing.T) {
	var got []string
	for _, binding := range generation.DefaultRegistry() {
		if binding.Operation == sourcepolicy.OperationSplitGo {
			got = binding.RequiredIndicatorIDs
		}
	}
	want := operationconformance.SplitGoV1IndicatorIDs()
	if !slices.Equal(got, want) {
		t.Fatalf("registry indicators = %v, want %v", got, want)
	}
}
