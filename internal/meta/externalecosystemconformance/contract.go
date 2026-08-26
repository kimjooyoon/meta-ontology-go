package externalecosystemconformance

const (
	ReferenceSchema = "gooo/external-ecosystem-reference/v1"
	ReportSchema    = "gooo/external-ecosystem-reference-report/v1"
	SuiteSchema     = "gooo/external-ecosystem-reference-conformance/v1"

	ExpectedReferenceID = "cosmos72/gomacro@cf0d4bf32da393dbda97e3572f216731013ffa55"
	ExpectedRepository  = "https://github.com/cosmos72/gomacro"
	ExpectedCommit      = "cf0d4bf32da393dbda97e3572f216731013ffa55"
	ExpectedTree        = "8cc240a53dd29432ad83620b20fd8a0a05674c6d"
	ExpectedLicense     = "MPL-2.0"
	ExpectedModule      = "github.com/cosmos72/gomacro"
	ExpectedGoVersion   = "1.23.0"
	ExpectedReadmeHash  = "sha256:8b453f2cfb1808b0de96a530525889f4bcdcc6b5aa866d8fc93b10d1473e2bc7"
	ExpectedGoModHash   = "sha256:95da44a97a57bc25ef6eac57d519281a3ccb3662ecb352e9fb6abdb22f804250"
	ExpectedReadmeBytes = 23127
	ExpectedGoModBytes  = 328

	DecisionReferenceBound = "REFERENCE_BOUND"
	DecisionFailClosed     = "FAIL_CLOSED"
	ResolutionExact        = "EXACT"
	ResolutionUnknown      = "UNKNOWN"
	ResolutionInvariant    = "INVARIANT_ONLY"
	EffectNoEffect         = "NO_EFFECT"
	EffectBlock            = "BLOCK"

	ReasonReferenceBound  = "EXTERNAL_ECOSYSTEM_REFERENCE_BOUND"
	ReasonUnavailable     = "EXTERNAL_ECOSYSTEM_EVIDENCE_UNAVAILABLE"
	ReasonDigestMismatch  = "EXTERNAL_ECOSYSTEM_DOCUMENT_DIGEST_MISMATCH"
	ReasonCapsuleMismatch = "EXTERNAL_ECOSYSTEM_CAPSULE_MISMATCH"
	ReasonCommitMismatch  = "EXTERNAL_ECOSYSTEM_COMMIT_MISMATCH"
	ReasonLicenseMismatch = "EXTERNAL_ECOSYSTEM_LICENSE_MISMATCH"
	ReasonRelationUnknown = "EXTERNAL_ECOSYSTEM_RELATION_UNKNOWN"
	ReasonModuleMismatch  = "EXTERNAL_ECOSYSTEM_MODULE_CONTRACT_MISMATCH"
	ReasonExecution       = "EXTERNAL_ECOSYSTEM_EXECUTION_OBSERVED"
	ReasonWrite           = "EXTERNAL_ECOSYSTEM_WRITE_OBSERVED"
	ReasonCaseUnknown     = "EXTERNAL_ECOSYSTEM_CASE_UNKNOWN"
)

type capabilityRule struct {
	Relation      string
	MetaOperation string
}

var capabilityRules = map[string]capabilityRule{
	"REPL":                       {"STRUCTURAL_HINT", "compare-interactive-loop"},
	"EMBEDDED_EVAL":              {"STRUCTURAL_HINT", "compare-embedded-evaluator"},
	"SOURCE_SCRIPTING":           {"STRUCTURAL_HINT", "compare-source-execution"},
	"AST_MACRO_EXPANSION":        {"STRUCTURAL_HINT", "compare-ast-transformation"},
	"THIRD_PARTY_IMPORT":         {"STRUCTURAL_HINT", "compare-import-boundary"},
	"GENERICS":                   {"DOCUMENTED_LIMITATION", "preserve-upstream-limitation"},
	"DEBUGGER":                   {"STRUCTURAL_HINT", "compare-debug-observability"},
	"UNRESTRICTED_MACRO_EFFECTS": {"GUARDRAIL_CONTRAST", "deny-unsealed-meta-effects"},
}

var caseIDs = []string{
	"exact", "readme-unavailable", "gomod-unavailable", "unknown-relation",
	"readme-digest-mismatch", "gomod-digest-mismatch", "commit-mismatch",
	"license-mismatch", "external-execution", "observed-write",
}
