package languagereadiness

import (
	"fmt"

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
	promotionDigest, err := validateProposalPromotion(promotion, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	if err := guardedcapability.ValidateForHead(capability, expectedHeadSHA); err != nil {
		return Snapshot{}, fmt.Errorf("verify guarded promotion capability: %w", err)
	}
	if capability.Decision != guardedcapability.DecisionPass {
		return Snapshot{}, fmt.Errorf("FAIL_CLOSED: guarded capability decision %q", capability.Decision)
	}
	if err := toolchainusecases.Validate(useCases, expectedHeadSHA); err != nil {
		return Snapshot{}, fmt.Errorf("verify executable use cases: %w", err)
	}
	if useCases.Decision != toolchainusecases.DecisionPass {
		return Snapshot{}, fmt.Errorf("FAIL_CLOSED: executable use case decision %q", useCases.Decision)
	}
	if err := languagesyntax.Validate(syntaxReport, expectedHeadSHA); err != nil {
		return Snapshot{}, fmt.Errorf("verify language syntax roundtrip: %w", err)
	}
	if syntaxReport.Decision != languagesyntax.DecisionPass {
		return Snapshot{}, fmt.Errorf("FAIL_CLOSED: language syntax decision %q", syntaxReport.Decision)
	}
	if err := languagediagnosticprovenance.Validate(diagnosticReport, expectedHeadSHA); err != nil {
		return Snapshot{}, fmt.Errorf("verify diagnostic provenance: %w", err)
	}
	runtimeDigest, err := validatePackageRuntime(packageRuntime, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	return evaluate(raw, evidenceDigests{
		proposal: promotionDigest, guarded: capability.ReportDigest, useCases: useCases.ReportDigest,
		syntax: syntaxReport.ReportDigest, diagnostic: diagnosticReport.ReportDigest,
		packageRuntime: runtimeDigest,
	})
}
