package languagereadiness

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagepackageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

func EvaluateWithProposalPromotion(
	raw []byte, promotion proposalpromotion.Receipt, expectedHeadSHA string,
) (Snapshot, error) {
	digest, err := validateProposalPromotion(promotion, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	return evaluate(raw, evidenceDigests{proposal: digest})
}

func EvaluateWithPromotionEvidence(raw []byte, promotion proposalpromotion.Receipt,
	capability guardedcapability.Receipt, useCases toolchainusecases.Report, syntaxReport languagesyntax.Report,
	diagnosticReport languagediagnosticprovenance.Report,
	expectedHeadSHA string, packageRuntime ...languagepackageruntime.Report) (Snapshot, error) {
	bundle := PromotionEvidence{Promotion: promotion, Capability: capability, UseCases: useCases,
		Syntax: syntaxReport, Diagnostic: diagnosticReport, PackageRuntime: packageRuntime}
	evidence, err := validatePromotionEvidence(bundle, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	return evaluate(raw, evidence)
}
