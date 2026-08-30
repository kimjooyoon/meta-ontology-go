package verify

const devMainSyncRerunBranch = "agent/dev-main-sync-20260831-rerun"

func init() {
	branchScopeAllowlist[devMainSyncRerunBranch] = []string{
		".github",
		"bootstrap",
		"cmd",
		"examples",
		"internal",
		"scripts",
	}
}
