package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	return RatioMetric{Numerator: numerator, Denominator: denominator, BasisPoints: basisPoints, Decision: decisionForRatio(numerator, denominator)}
}

func integrationMetrics(localCount, sharedCount, generatedChanged, adopted, manualEdits int) IntegrationMetrics {
	return IntegrationMetrics{
		ExistingSharedSourceTouchpoints: ratioMetric(sharedCount, sharedCount),
		GeneratorChangedSharedOutputs:   ratioMetric(generatedChanged, 8),
		IndependentConformanceConsumer:  ratioMetric(adopted, 1),
		ConceptLocalTouchpoints:         ratioMetric(localCount, localCount),
		ManualSourceRegistrationEdits:   ratioMetric(manualEdits, sharedCount),
	}
}

func contextualRatioMetric(numerator, denominator int, stage, step, reason string) RatioMetric {
	metric := ratioMetric(numerator, denominator)
	metric.Stage, metric.Step, metric.Reason = stage, step, reason
	return metric
}

func unknownRatioMetric(stage, step, reason string) RatioMetric {
	return RatioMetric{Numerator: 0, Denominator: 1, BasisPoints: 0, Decision: "UNKNOWN", Stage: stage, Step: step, Reason: reason}
}

func producerPackageImportMetric(root string) RatioMetric {
	metric := contextualRatioMetric(0, 1, "COHERENCE", "PRODUCER_PACKAGE_IMPORT", "INDEPENDENT_CONSUMER_IMPORTS_PRODUCER_PACKAGE")
	moduleBytes, err := os.ReadFile(joinRoot(root, "go.mod"))
	if err != nil {
		metric.Decision, metric.Reason = "UNKNOWN", "MODULE_PATH_UNAVAILABLE"
		return metric
	}
	module := ""
	for _, line := range strings.Split(string(moduleBytes), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			module = fields[1]
			break
		}
	}
	if module == "" {
		metric.Decision, metric.Reason = "UNKNOWN", "MODULE_PATH_UNAVAILABLE"
		return metric
	}
	producerPath := module + "/scripts/conflict-free-registry-projection"
	files, err := filepath.Glob(joinRoot(root, "scripts/conflict-free-registry-projection-consumer/*.go"))
	if err != nil {
		metric.Decision, metric.Reason = "UNKNOWN", "CONSUMER_SOURCE_UNAVAILABLE"
		return metric
	}
	fileSet := token.NewFileSet()
	for _, path := range files {
		file, parseErr := parser.ParseFile(fileSet, path, nil, 0)
		if parseErr != nil {
			metric.Decision, metric.Reason = "FAIL_CLOSED", "CONSUMER_SOURCE_PARSE_FAILED"
			return metric
		}
		for _, spec := range file.Imports {
			importPath, unquoteErr := strconv.Unquote(spec.Path.Value)
			if unquoteErr == nil && importPath == producerPath {
				metric.Numerator = 1
				metric.BasisPoints = 10000
				metric.Decision = "FAIL_CLOSED"
				metric.Reason = "PRODUCER_PACKAGE_IMPORT_OBSERVED"
				return metric
			}
		}
	}
	metric.Decision = "PASS"
	metric.Reason = "PRODUCER_PACKAGE_IMPORT_NOT_OBSERVED"
	return metric
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
