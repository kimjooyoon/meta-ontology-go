package verify

import (
	"fmt"
	"sort"
	"strings"
)

const reconciliationMainBranchPrefix = "agent/main-history-reconciliation-"

// CheckPullRequestPolicy enforces the steady-state branch policy used by CI.
func CheckPullRequestPolicy(head, base string) error {
	if base == "main" {
		if head != "dev" && (len(head) <= len(reconciliationMainBranchPrefix) || !strings.HasPrefix(head, reconciliationMainBranchPrefix)) {
			return fmt.Errorf("main promotion head must be dev, got %q", head)
		}
		return nil
	}
	if base != "dev" {
		return fmt.Errorf("feature pull request base must be dev, got %q", base)
	}
	if !strings.HasPrefix(head, "agent/") || len(strings.TrimPrefix(head, "agent/")) == 0 {
		return fmt.Errorf("feature pull request head must use agent/*, got %q", head)
	}
	return nil
}
func sortedUnique(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	if len(result) < 2 {
		return result
	}
	write := 1
	for _, value := range result[1:] {
		if value == result[write-1] {
			continue
		}
		result[write] = value
		write++
	}
	return result[:write]
}
