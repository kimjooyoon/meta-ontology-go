package languagereadiness

import (
	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
	metaconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
)

func EvaluateWithToolchainConformance(raw []byte, bundle PromotionEvidence,
	cliReport metacli.Report, formatFixReport metaff.Report,
	conformanceReport metaconformance.Report, expectedHeadSHA string) (Snapshot, error) {
	evidence, err := validatePromotionEvidence(bundle, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	cliDigest, err := validateToolchainCLI([]metacli.Report{cliReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	formatFixDigest, err := validateToolchainFormatFix(
		[]metaff.Report{formatFixReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	conformanceDigest, err := validateToolchainConformance(
		[]metaconformance.Report{conformanceReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	evidence.toolchainCLI = cliDigest
	evidence.toolchainFormatFix = formatFixDigest
	evidence.toolchainConformance = conformanceDigest
	return evaluate(raw, evidence)
}
