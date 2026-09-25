package verify

func init() {
	branchScopeAllowlist["agent/guardian-kernel-digest-alignment-20260903"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		"internal/verify/scope_guardian_kernel_digest_alignment_20260903.go",
		"scripts/ci-proof/guardian.js",
	}
}
