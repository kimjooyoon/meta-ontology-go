package extractor

import "testing"

func TestOpenFGAComparisonNormalizerRenderedCapacity(t *testing.T) {
	requireNamedCapacities(t, "../../../..", []string{
		"examples/openfga-model-change/normalize#digest",
		"examples/openfga-model-change/normalize#encoded",
		"examples/openfga-model-change/normalize#normalizeObservation",
		"examples/openfga-model-change/normalize#readSnapshot",
		"examples/openfga-model-change/normalize#main",
	})
}
