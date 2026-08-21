package verify

import (
	"fmt"
	"sort"
	"strings"
)

func validateKernelPaths(paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("protected kernel paths are required")
	}
	for _, path := range paths {
		wildcard := strings.HasSuffix(path, "/**")
		base := strings.TrimSuffix(path, "/**")
		if path == "" || strings.ContainsAny(base, "*?[]") || strings.HasPrefix(path, "/") || (strings.Contains(path, "*") && !wildcard) {
			return fmt.Errorf("invalid protected kernel path %q", path)
		}
	}
	for _, required := range mandatoryKernelPaths() {
		if !contains(paths, required) {
			return fmt.Errorf("protected kernel paths cannot shrink below %q", required)
		}
	}
	return nil
}
func mandatoryKernelPaths() []string {
	return []string{
		".github/workflows/**",
		".github/ci-governance.json",
		".github/agent-scope-table.md",
		".github/branch-policy.md",
		".github/conformance-plan.md",
		"scripts/ci-proof/**",
		"scripts/ci-evidence/**",
		"scripts/verify/**",
		"internal/verify/**",
		"go.mod",
		"go.sum",
	}
}
func validatePromotion(promotion GovernancePromotion) error {
	if promotion.Source != "dev" || promotion.Target != "main" {
		return fmt.Errorf("promotion must be dev to main")
	}
	if !promotion.BranchProtectionRequired {
		return fmt.Errorf("promotion branch protection is required")
	}
	if !sameStrings(promotion.RequiredChecks, canonicalJobs()) {
		return fmt.Errorf("promotion checks do not match canonical CI jobs")
	}
	return nil
}
func samePaths(left, right []string) bool {
	return sameStrings(normalizeMatrixPaths(left), normalizeMatrixPaths(right))
}
func normalizeMatrixPaths(paths []string) []string {
	normalized := make([]string, len(paths))
	for index, path := range paths {
		normalized[index] = strings.TrimSuffix(path, "/**")
	}
	sort.Strings(normalized)
	return normalized
}
