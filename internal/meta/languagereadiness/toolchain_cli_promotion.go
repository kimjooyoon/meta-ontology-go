package languagereadiness

import metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"

func EvaluateWithToolchainCLI(raw []byte, bundle PromotionEvidence,
	cliReport metacli.Report, expectedHeadSHA string) (Snapshot, error) {
	evidence, err := validatePromotionEvidence(bundle, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	cliDigest, err := validateToolchainCLI([]metacli.Report{cliReport}, expectedHeadSHA)
	if err != nil {
		return Snapshot{}, err
	}
	evidence.toolchainCLI = cliDigest
	return evaluate(raw, evidence)
}
