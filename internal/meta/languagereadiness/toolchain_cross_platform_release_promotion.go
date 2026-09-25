package languagereadiness

import (
	metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"
	metaconformance "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainconformance"
	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
	metalsp "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainlsp"
	release "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainrelease"
)

func EvaluateWithToolchainCrossPlatformRelease(raw []byte, bundle PromotionEvidence,
	cliReport metacli.Report, formatFixReport metaff.Report,
	conformanceReport metaconformance.Report, lspReport metalsp.Report,
	releaseReport release.Report,
	expectedRepository, expectedHeadSHA, expectedPredecessorSHA string) (Snapshot, error) {
	evidence, err := validatePromotionEvidence(bundle, expectedRepository, expectedHeadSHA, expectedPredecessorSHA)
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
	releaseDigest, err := validateToolchainCrossPlatformRelease([]release.Report{releaseReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	evidence.toolchainCLI, evidence.toolchainFormatFix = cliDigest, formatDigest
	evidence.toolchainConformance, evidence.toolchainLSP = conformanceDigest, lspDigest
	evidence.toolchainRelease = releaseDigest
	return evaluate(raw, evidence)
}
