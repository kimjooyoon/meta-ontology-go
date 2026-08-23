package languagereadiness

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

func validatePromotionEvidence(bundle PromotionEvidence, expectedHeadSHA string) (evidenceDigests, error) {
	promotionDigest, err := validateProposalPromotion(bundle.Promotion, expectedHeadSHA)
	if err != nil {
		return evidenceDigests{}, err
	}
	if err := guardedcapability.ValidateForHead(bundle.Capability, expectedHeadSHA); err != nil {
		return evidenceDigests{}, fmt.Errorf("verify guarded promotion capability: %w", err)
	}
	if bundle.Capability.Decision != guardedcapability.DecisionPass {
		return evidenceDigests{}, fmt.Errorf("FAIL_CLOSED: guarded capability decision %q", bundle.Capability.Decision)
	}
	if err := toolchainusecases.Validate(bundle.UseCases, expectedHeadSHA); err != nil {
		return evidenceDigests{}, fmt.Errorf("verify executable use cases: %w", err)
	}
	if bundle.UseCases.Decision != toolchainusecases.DecisionPass {
		return evidenceDigests{}, fmt.Errorf("FAIL_CLOSED: executable use case decision %q", bundle.UseCases.Decision)
	}
	if err := validateLanguageEvidence(bundle, expectedHeadSHA); err != nil {
		return evidenceDigests{}, err
	}
	runtimeDigest, err := validatePackageRuntime(bundle.PackageRuntime, expectedHeadSHA)
	if err != nil {
		return evidenceDigests{}, err
	}
	return evidenceDigests{proposal: promotionDigest, guarded: bundle.Capability.ReportDigest,
		useCases: bundle.UseCases.ReportDigest, syntax: bundle.Syntax.ReportDigest,
		diagnostic: bundle.Diagnostic.ReportDigest, packageRuntime: runtimeDigest}, nil
}
