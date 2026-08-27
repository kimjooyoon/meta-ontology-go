package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type baselineTouchpoint struct {
	Path   string
	Kind   string
	Tokens []string
}

var baselineTouchpoints = []baselineTouchpoint{
	{Path: "internal/meta/languageconcept/catalog.go", Kind: "CATALOG", Tokens: []string{"language-syntax-roundtrip"}},
	{Path: "examples/language-syntax-roundtrip/corpus.json", Kind: "SYNTAX_CORPUS", Tokens: []string{"language-syntax-roundtrip-corpus"}},
	{Path: "examples/language-syntax-roundtrip/README.md", Kind: "SYNTAX_README", Tokens: []string{"corpus.json"}},
	{Path: "docs/language/language-syntax-roundtrip.md", Kind: "SYNTAX_DOCUMENTATION", Tokens: []string{"syntax round-trip"}},
	{Path: "internal/meta/languagereadiness/languagesyntax/registry.go", Kind: "SYNTAX_REGISTRY", Tokens: []string{"examples/language-syntax-roundtrip"}},
	{Path: "examples/language-semantic-model/corpus.json", Kind: "SEMANTIC_CORPUS", Tokens: []string{"language-semantic-model-corpus"}},
	{Path: "examples/language-semantic-model/README.md", Kind: "SEMANTIC_README", Tokens: []string{"fixed denominator"}},
	{Path: "docs/language/language-semantic-model.md", Kind: "SEMANTIC_DOCUMENTATION", Tokens: []string{"language semantic model"}},
	{Path: "internal/meta/languagereadiness/languagesemantic/registry_definition.go", Kind: "SEMANTIC_REGISTRY", Tokens: []string{"expectedSources"}},
	{Path: "internal/meta/languagereadiness/languagesemanticbinding/denominator.go", Kind: "SEMANTIC_DENOMINATOR", Tokens: []string{"semanticCaseDenominator"}},
	{Path: "examples/toolchain-conformance/corpus.json", Kind: "CONFORMANCE_CORPUS", Tokens: []string{"toolchain-conformance-corpus"}},
	{Path: "docs/language/toolchain-conformance.md", Kind: "CONFORMANCE_DOCUMENTATION", Tokens: []string{"Toolchain conformance"}},
}

type baselineObservation struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Observed bool   `json:"observed"`
}

type baselineReport struct {
	Touchpoints []baselineObservation `json:"touchpoints"`
	Observed    int                   `json:"observed"`
	Expected    int                   `json:"expected"`
}

func observeBaseline(root string) (baselineReport, *Diagnostic) {
	report := baselineReport{Expected: len(baselineTouchpoints), Touchpoints: make([]baselineObservation, 0, len(baselineTouchpoints))}
	for _, touchpoint := range baselineTouchpoints {
		data, err := os.ReadFile(joinRoot(root, touchpoint.Path))
		if err != nil {
			return report, &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BASELINE_TOUCHPOINT", Reason: "MISSING_BASELINE_TOUCHPOINT"}
		}
		observed := false
		for _, token := range touchpoint.Tokens {
			if strings.Contains(string(data), token) {
				observed = true
				break
			}
		}
		report.Touchpoints = append(report.Touchpoints, baselineObservation{Path: touchpoint.Path, Kind: touchpoint.Kind, Observed: observed})
		if observed {
			report.Observed++
		}
	}
	if report.Observed != report.Expected {
		return report, &Diagnostic{Decision: "FAIL_CLOSED", Stage: "FOUNDATION", Step: "BASELINE_TOUCHPOINT", Reason: "BASELINE_TOUCHPOINT_NOT_OBSERVED"}
	}
	return report, nil
}

func ratioMetric(numerator, denominator int) RatioMetric {
	basisPoints := 0
	if denominator > 0 {
		basisPoints = numerator * 10000 / denominator
	}
	return RatioMetric{Numerator: numerator, Denominator: denominator, BasisPoints: basisPoints}
}

func integrationMetrics(localCount, sharedCount, generatedChanged, adopted, manualEdits int) IntegrationMetrics {
	return IntegrationMetrics{
		ExistingSharedSourceTouchpoints: ratioMetric(sharedCount, sharedCount),
		GeneratorChangedSharedOutputs:   ratioMetric(generatedChanged, 8),
		ProductionConsumerAdoption:      ratioMetric(adopted, 1),
		ConceptLocalTouchpoints:         ratioMetric(localCount, localCount),
		ManualSourceRegistrationEdits:   ratioMetric(manualEdits, sharedCount),
	}
}

func metricDeltas(sharedCount, generatedChanged, manualEdits int) map[string]MetricDelta {
	full := ratioMetric(sharedCount, sharedCount)
	generated := ratioMetric(generatedChanged, 8)
	manual := ratioMetric(manualEdits, sharedCount)
	return map[string]MetricDelta{
		"existing_shared_source_touchpoints": {Before: full, After: manual},
		"generator_changed_shared_outputs":   {Before: ratioMetric(0, 8), After: generated},
		"manual_source_registration_edits":   {Before: full, After: manual},
	}
}

func joinRoot(root, path string) string {
	return root + "/" + path
}

func sortedStringsCopy(values []string) []string {
	copyOf := append([]string(nil), values...)
	sort.Strings(copyOf)
	return copyOf
}

func formatBaseline(report baselineReport) string {
	return fmt.Sprintf("%d/%d observed manual shared registration touchpoints", report.Observed, report.Expected)
}
