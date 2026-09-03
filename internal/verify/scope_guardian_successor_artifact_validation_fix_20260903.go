package verify

func init() {
	branchScopeAllowlist["agent/guardian-successor-artifact-validation-fix-20260903"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_guardian_successor_artifact_validation_fix_20260903.go",
		"scripts/ci-proof/guardian.js",
	}
}
