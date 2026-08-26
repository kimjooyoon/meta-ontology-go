package languagepackageruntime

import (
	"runtime"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func Evaluate(input Input) Report {
	source := Source{ExpectedHeadSHA: input.ExpectedHeadSHA, GoVersion: runtime.Version(),
		ConceptDigest: input.ConceptArtifact.ArtifactDigest, RegistryDigest: digestBytes(input.RegistryRaw),
		ManifestSchema: packageruntime.ManifestSchema}
	registry, err := DecodeRegistry(input.RegistryRaw)
	source.ObservationKnown = err == nil && input.ExpectedHeadSHA != "" &&
		strings.HasPrefix(runtime.Version(), "go1.27") && input.ConceptArtifact.Ready() &&
		input.ConceptArtifact.Report.Summary.Concepts >= 19
	if !source.ObservationKnown {
		return lowerResolution(source, "PACKAGE_RUNTIME_EVIDENCE_UNKNOWN")
	}
	baseline, err := packageruntime.Run(baseManifest())
	if err != nil {
		return lowerResolution(source, "PACKAGE_RUNTIME_BASELINE_UNKNOWN")
	}
	results := make([]CaseResult, 0, FixedTotal)
	for _, definition := range registry.Cases {
		if definition.Kind == "POSITIVE" {
			results = append(results, executePositive(definition, baseline))
		} else {
			results = append(results, executeGuardrail(definition))
		}
	}
	summary := summarize(registry.Cases, results)
	report := Report{Schema: ReportSchema, Decision: DecisionClosed, Resolution: ResolutionExact,
		ReasonCode: "PACKAGE_RUNTIME_CASE_NOT_SATISFIED", Source: source, Summary: summary, Cases: results}
	report.Indicators = indicators(summary, report.Resolution)
	report.Proofs, report.Stages = proofs(report), stages(source, summary)
	if summary.Satisfied == FixedTotal && allIndicators(report.Indicators) && allProofs(report.Proofs) {
		report.Decision, report.ReasonCode = DecisionPass, "ALL_PACKAGE_RUNTIME_CASES_SATISFIED"
	}
	report.ReportDigest = reportDigest(report)
	return report
}
