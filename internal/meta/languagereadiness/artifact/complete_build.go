package artifact

import (
	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagepackageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
	metaconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

func BuildWithCompleteEvidence(input CompleteEvidenceInput) (Receipt, error) {
	promotion, err := decodeCompleteEvidence[proposalpromotion.Receipt](input.Promotion)
	if err != nil {
		return Receipt{}, err
	}
	capability, err := decodeCompleteEvidence[guardedcapability.Receipt](input.Capability)
	if err != nil {
		return Receipt{}, err
	}
	useCases, err := decodeCompleteEvidence[toolchainusecases.Report](input.UseCases)
	if err != nil {
		return Receipt{}, err
	}
	syntaxReport, err := decodeCompleteEvidence[languagesyntax.Report](input.Syntax)
	if err != nil {
		return Receipt{}, err
	}
	diagnostic, err := decodeCompleteEvidence[languagediagnosticprovenance.Report](input.Diagnostic)
	if err != nil {
		return Receipt{}, err
	}
	runtimeReport, err := decodeCompleteEvidence[languagepackageruntime.Report](input.PackageRuntime)
	if err != nil {
		return Receipt{}, err
	}
	cliReport, err := decodeCompleteEvidence[metacli.Report](input.ToolchainCLI)
	if err != nil {
		return Receipt{}, err
	}
	formatFixReport, err := decodeCompleteEvidence[metaff.Report](input.ToolchainFormatFix)
	if err != nil {
		return Receipt{}, err
	}
	bundle := readiness.PromotionEvidence{Promotion: promotion, Capability: capability,
		UseCases: useCases, Syntax: syntaxReport, Diagnostic: diagnostic,
		PackageRuntime: []languagepackageruntime.Report{runtimeReport}}
	var snapshot readiness.Snapshot
	if len(input.ToolchainConformance) == 0 {
		snapshot, err = readiness.EvaluateWithToolchainFormatFix(
			input.ConceptArtifact, bundle, cliReport, formatFixReport, input.HeadSHA)
	} else {
		conformance, decodeErr := decodeCompleteEvidence[metaconformance.Report](
			input.ToolchainConformance)
		if decodeErr != nil {
			return Receipt{}, decodeErr
		}
		snapshot, err = readiness.EvaluateWithToolchainConformance(
			input.ConceptArtifact, bundle, cliReport, formatFixReport, conformance, input.HeadSHA)
	}
	if err != nil {
		return Receipt{}, err
	}
	return build(snapshot, input.HeadSHA, promotion.ReportDigest, capability.ReportDigest)
}
