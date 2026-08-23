package languagereadiness

import (
	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
	metaconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
	metalsp "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainlsp"
)

func EvaluateWithToolchainLSP(raw []byte, bundle PromotionEvidence, cliReport metacli.Report,
	formatFixReport metaff.Report, conformanceReport metaconformance.Report,
	lspReport metalsp.Report, expectedHeadSHA string) (Snapshot, error) {
	evidence, err := validatePromotionEvidence(bundle, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	cliDigest, err := validateToolchainCLI([]metacli.Report{cliReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	formatDigest, err := validateToolchainFormatFix([]metaff.Report{formatFixReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	conformanceDigest, err := validateToolchainConformance([]metaconformance.Report{conformanceReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	lspDigest, err := validateToolchainLSP([]metalsp.Report{lspReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	evidence.toolchainCLI, evidence.toolchainFormatFix = cliDigest, formatDigest
	evidence.toolchainConformance, evidence.toolchainLSP = conformanceDigest, lspDigest
	return evaluate(raw, evidence)
}
