package languagereadiness

const (
	autonomousProposalConcept = "autonomous-change-proposal"
	guardedPromotionConcept   = "guarded-exact-promotion"
	toolchainUseCasesConcept  = "toolchain-executable-use-cases"
	languageSyntaxConcept     = "language-syntax-roundtrip"
	diagnosticConcept         = "language-diagnostic-provenance"
	packageRuntimeConcept     = "language-package-runtime"
	toolchainCLIConcept       = "toolchain-cli"
	toolchainFormatFixConcept = "toolchain-format-fix"
	toolchainConformanceConcept = "toolchain-conformance"
)

type evidenceDigests struct {
	proposal           string
	guarded            string
	useCases           string
	syntax             string
	diagnostic         string
	packageRuntime     string
	toolchainCLI       string
	toolchainFormatFix string
	toolchainConformance string
}

type externalEvidence struct {
	Concept conceptEvidence `json:"concept"`
	Digest  string          `json:"digest"`
}

func requiredEvidence(conceptID string, evidence evidenceDigests) (string, string, bool) {
	switch conceptID {
	case autonomousProposalConcept:
		return evidence.proposal, "VERIFIED_PROMOTION_RECEIPT_REQUIRED", true
	case guardedPromotionConcept:
		return evidence.guarded, "GUARDED_CAPABILITY_RECEIPT_REQUIRED", true
	case toolchainUseCasesConcept:
		return evidence.useCases, "EXECUTABLE_USE_CASE_RECEIPT_REQUIRED", true
	case languageSyntaxConcept:
		return evidence.syntax, "LANGUAGE_SYNTAX_ROUNDTRIP_RECEIPT_REQUIRED", true
	case diagnosticConcept:
		return evidence.diagnostic, "LANGUAGE_DIAGNOSTIC_PROVENANCE_RECEIPT_REQUIRED", true
	case packageRuntimeConcept:
		return evidence.packageRuntime, "LANGUAGE_PACKAGE_RUNTIME_RECEIPT_REQUIRED", true
	case toolchainCLIConcept:
		return evidence.toolchainCLI, "TOOLCHAIN_CLI_RECEIPT_REQUIRED", true
	case toolchainFormatFixConcept:
		return evidence.toolchainFormatFix, "TOOLCHAIN_FORMAT_FIX_RECEIPT_REQUIRED", true
	case toolchainConformanceConcept:
		return evidence.toolchainConformance, "TOOLCHAIN_CONFORMANCE_RECEIPT_REQUIRED", true
	default:
		return "", "", false
	}
}
