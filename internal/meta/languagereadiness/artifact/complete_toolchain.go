package artifact

import (
	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
	metaconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
	metalsp "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainlsp"
)

func buildToolchainSnapshot(input CompleteEvidenceInput, bundle readiness.PromotionEvidence,
	cliReport metacli.Report, formatFixReport metaff.Report) (readiness.Snapshot, error) {
	if len(input.ToolchainConformance) == 0 {
		return readiness.EvaluateWithToolchainFormatFix(input.ConceptArtifact, bundle,
			cliReport, formatFixReport, input.HeadSHA)
	}
	conformance, err := decodeCompleteEvidence[metaconformance.Report](input.ToolchainConformance)
	if err != nil { return readiness.Snapshot{}, err }
	if len(input.ToolchainLSP) == 0 {
		return readiness.EvaluateWithToolchainConformance(input.ConceptArtifact, bundle,
			cliReport, formatFixReport, conformance, input.HeadSHA)
	}
	lspReport, err := decodeCompleteEvidence[metalsp.Report](input.ToolchainLSP)
	if err != nil { return readiness.Snapshot{}, err }
	return readiness.EvaluateWithToolchainLSP(input.ConceptArtifact, bundle, cliReport,
		formatFixReport, conformance, lspReport, input.HeadSHA)
}
