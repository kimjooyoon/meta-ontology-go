package artifact

import (
	"encoding/json"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

func BuildWithPromotionEvidence(
	conceptArtifact, promotionRaw, capabilityRaw, useCaseRaw, syntaxRaw []byte, headSHA string,
) (Receipt, error) {
	promotion := proposalpromotion.Receipt{}
	if err := json.Unmarshal(promotionRaw, &promotion); err != nil {
		return Receipt{}, err
	}
	capability := guardedcapability.Receipt{}
	if err := json.Unmarshal(capabilityRaw, &capability); err != nil {
		return Receipt{}, err
	}
	useCases := toolchainusecases.Report{}
	if err := json.Unmarshal(useCaseRaw, &useCases); err != nil {
		return Receipt{}, err
	}
	syntaxReport := languagesyntax.Report{}
	if err := json.Unmarshal(syntaxRaw, &syntaxReport); err != nil {
		return Receipt{}, err
	}
	snapshot, err := readiness.EvaluateWithPromotionEvidence(
		conceptArtifact, promotion, capability, useCases, syntaxReport, headSHA,
	)
	if err != nil {
		return Receipt{}, err
	}
	return build(snapshot, headSHA, promotion.ReportDigest, capability.ReportDigest)
}
