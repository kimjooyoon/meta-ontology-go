package languagereadiness

import (
	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
)

func EvaluateWithToolchainFormatFix(raw []byte, bundle PromotionEvidence,
	cliReport metacli.Report, formatFixReport metaff.Report, expectedHeadSHA string) (Snapshot, error) {
	evidence, err := validatePromotionEvidence(bundle, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	cliDigest, err := validateToolchainCLI([]metacli.Report{cliReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	formatFixDigest, err := validateToolchainFormatFix([]metaff.Report{formatFixReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	evidence.toolchainCLI, evidence.toolchainFormatFix = cliDigest, formatFixDigest
	return evaluate(raw, evidence)
}
