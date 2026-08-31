package artifact

import (
	"encoding/json"
	"fmt"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedcapability"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagediagnosticprovenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagepackageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/proposalpromotion"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainusecases"
)

func BuildWithPromotionEvidence(
	conceptArtifact, promotionRaw, capabilityRaw, useCaseRaw, syntaxRaw,
	diagnosticRaw []byte, expectedRepository, headSHA, expectedPredecessorSHA string,
	packageRuntimeRaw ...[]byte,
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
	diagnosticReport := languagediagnosticprovenance.Report{}
	if err := json.Unmarshal(diagnosticRaw, &diagnosticReport); err != nil {
		return Receipt{}, err
	}
	packageRuntimeReports, err := decodePackageRuntime(packageRuntimeRaw)
	if err != nil {
		return Receipt{}, err
	}
	snapshot, err := readiness.EvaluateWithPromotionEvidence(
		conceptArtifact, promotion, capability, useCases, syntaxReport,
		diagnosticReport, expectedRepository, headSHA, expectedPredecessorSHA, packageRuntimeReports...,
	)
	if err != nil {
		return Receipt{}, err
	}
	return build(snapshot, headSHA, promotion.ReportDigest, capability.ReportDigest)
}

func decodePackageRuntime(rawValues [][]byte) ([]languagepackageruntime.Report, error) {
	if len(rawValues) == 0 {
		return nil, nil
	}
	if len(rawValues) != 1 {
		return nil, fmt.Errorf("package runtime evidence is not unique")
	}
	report := languagepackageruntime.Report{}
	if err := json.Unmarshal(rawValues[0], &report); err != nil {
		return nil, err
	}
	return []languagepackageruntime.Report{report}, nil
}
