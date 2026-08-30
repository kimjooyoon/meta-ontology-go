package verify

func init() {
	branchScopeAllowlist["agent/dev-main-sync-20260831"] = []string{
		".github",
		"bootstrap",
		"cmd",
		"examples",
		"internal",
		"scripts",
	}
}
