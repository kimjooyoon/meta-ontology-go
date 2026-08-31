package languagereadiness

import metacli "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchaincli"

func EvaluateWithToolchainCLI(raw []byte, bundle PromotionEvidence,
	cliReport metacli.Report,
	expectedRepository, expectedHeadSHA, expectedPredecessorSHA string) (Snapshot, error) {
	evidence, err := validatePromotionEvidence(bundle, expectedRepository, expectedHeadSHA, expectedPredecessorSHA)
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
