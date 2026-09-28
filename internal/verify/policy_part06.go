package verify

import (
	"fmt"
	"sort"
	"strings"
)

// CheckPullRequestPolicy enforces the steady-state branch policy used by CI.
func CheckPullRequestPolicy(head, base string) error {
	if base == "main" {
		if head != "dev" && !isPromotionSnapshotHead(head) {
			return fmt.Errorf(
				"main promotion head must be dev or an exact dev-tree snapshot, got %q",
				head,
			)
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

func isPromotionSnapshotHead(head string) bool {
	const prefix = "agent/main-promotion-snapshot-"
	if !strings.HasPrefix(head, prefix) {
		return false
	}
	sha := strings.TrimPrefix(head, prefix)
	if len(sha) != 40 || sha == strings.Repeat("0", 40) {
		return false
	}
	for _, character := range sha {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
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
