package verify

import (
	"fmt"
	"sort"
	"strings"
)

// branchScopeAllowlist is the protected ownership map for agent branches.
// Paths are repository-relative directory prefixes unless they name a file.
var branchScopeAllowlist = map[string][]string{
	"agent/analyzer":                      {"internal/analyzer"},
	"agent/bidir":                         {"internal/bidir"},
	"agent/bidir-research":                {"docs/research/bidirectional.md"},
	"agent/bidirectional-research":        {"docs/research/bidirectional.md"},
	"agent/cache":                         {"internal/cache"},
	"agent/cache-research":                {"docs/research/cache.md"},
	"agent/ci-workflow":                   {".github", "scripts", "internal/verify"},
	"agent/cli":                           {"cmd/gooo"},
	"agent/codegen-research":              {"docs/research/codegen.md"},
	"agent/conformance-fuzz":              {"internal/conformance/fuzz"},
	"agent/dependency-cycle-detector":     {"internal/detection/cycles"},
	"agent/detection":                     {"internal/detection"},
	"agent/detection-cycles":              {"internal/detection/cycles"},
	"agent/docs":                          {"AGENTS.md", "CONTRIBUTING.md", "README.md", "docs", "examples"},
	"agent/formatter":                     {"internal/formatter"},
	"agent/freshness-detection":           {"internal/detection/freshness"},
	"agent/fuzz-conformance":              {"internal/conformance/fuzz"},
	"agent/generator":                     {"internal/generator"},
	"agent/go-version":                    {"go.mod"},
	"agent/grammar-research":              {"docs/research/grammar.md"},
	"agent/grammar-review":                {"docs/research/grammar.md"},
	"agent/lsp-research":                  {"docs/research/lsp.md"},
	"agent/lsp":                           {"internal/lsp"},
	"agent/line-cap-detector":             {"internal/detection/linecaps"},
	"agent/performance-regression":        {"internal/detection/performance"},
	"agent/protected-regions":             {"internal/detection/protectedregions"},
	"agent/prototype-conformance":         {"internal/conformance"},
	"agent/prototype-detection":           {"internal/detection"},
	"agent/prototype-formatter":           {"internal/formatter"},
	"agent/prototype-provenance":          {"internal/provenance"},
	"agent/prototype-query":               {"internal/query"},
	"agent/prov-o-research":               {"docs/research/prov-o.md"},
	"agent/provenance-evidence":           {"internal/provenance"},
	"agent/provenance-freshness-detector": {"internal/detection/freshness"},
	"agent/provenance-store":              {"internal/provenance"},
	"agent/query-engine":                  {"internal/query"},
	"agent/query-research":                {"docs/research/query.md"},
	"agent/roundtrip-detector":            {"internal/detection/roundtrip"},
	"agent/roundtrip-detection":           {"internal/detection/roundtrip"},
	"agent/security-research":             {"docs/research/security.md"},
	"agent/security":                      {"docs/research/security.md"},
	"agent/semantic":                      {"internal/semantic"},
	"agent/semantic-delta-detector":       {"internal/detection/semanticdelta"},
	"agent/semanticdelta":                 {"internal/detection/semanticdelta"},
	"agent/self-hosting-bootstrap":        {"docs/research/self-hosting.md", "internal/bootstrap", "internal/verify/scope.go", "internal/verify/policy_test.go"},
	"agent/syntax":                        {"internal/syntax"},
	"agent/testing-research":              {"docs/research/testing.md"},
	"agent/zerolang-research":             {"docs/research/zerolang.md"},
}

// BranchScope returns a defensive copy of the configured ownership paths.
func BranchScope(branch string) ([]string, bool) {
	paths, ok := branchScopeAllowlist[branch]
	return append([]string(nil), paths...), ok
}

// CheckPathScopeForBranch applies the explicit ownership map and fails closed
// for unknown agent branches. Shared CI files belong only to agent/ci-workflow.
func CheckPathScopeForBranch(paths []string, branch string) error {
	allowed, known := BranchScope(branch)
	if !known {
		return fmt.Errorf("unknown agent branch %q; no paths are allowed", branch)
	}
	return CheckPathScope(paths, allowed)
}

// CheckGoModToolchainDiff accepts only added or removed go/toolchain
// directives from the agent/go-version go.mod diff.
func CheckGoModToolchainDiff(diff string) error {
	violations := make([]string, 0)
	for _, line := range strings.Split(diff, "\n") {
		if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
			continue
		}
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		content := strings.TrimSpace(line[1:])
		if !isToolchainDirective(content) {
			violations = append(violations, content)
		}
	}
	if len(violations) > 0 {
		return fmt.Errorf("go.mod changes outside Go toolchain directives: %s", strings.Join(violations, "; "))
	}
	return nil
}

func isToolchainDirective(line string) bool {
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return false
	}
	if fields[0] == "go" {
		return validGoVersion(fields[1])
	}
	return fields[0] == "toolchain" && strings.HasPrefix(fields[1], "go") && validGoVersion(strings.TrimPrefix(fields[1], "go"))
}

func validGoVersion(value string) bool {
	if !strings.HasPrefix(value, "1.") || len(value) < 3 {
		return false
	}
	for _, character := range value[2:] {
		if character != '.' && character != '-' && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

// ConfiguredBranches returns branch names in deterministic order for policy
// diagnostics and tests.
func ConfiguredBranches() []string {
	branches := make([]string, 0, len(branchScopeAllowlist))
	for branch := range branchScopeAllowlist {
		branches = append(branches, branch)
	}
	sort.Strings(branches)
	return branches
}
