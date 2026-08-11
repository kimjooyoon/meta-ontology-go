package verify

import (
	"fmt"
	"sort"
	"strings"
)

// branchScopeAllowlist is the protected ownership map for agent branches.
// Paths are repository-relative directory prefixes unless they name a file.
var branchScopeAllowlist = map[string][]string{
	"agent/analyzer":                          {"internal/analyzer"},
	"agent/analyzer-contract":                 {"internal/analyzer"},
	"agent/bidir":                             {"internal/bidir"},
	"agent/bidir-followup":                    {"internal/bidir"},
	"agent/bidirectional-experiment-contract": {"docs/research/bidirectional.md"},
	"agent/bidirectional-property-matrix":     {"docs/research/bidirectional.md"},
	"agent/bidir-research":                    {"docs/research/bidirectional.md"},
	"agent/bidirectional-research":            {"docs/research/bidirectional.md"},
	"agent/cache":                             {"internal/cache"},
	"agent/cache-experiment-followup":         {"docs/research/cache.md"},
	"agent/cache-research":                    {"docs/research/cache.md"},
	"agent/ci-ownership-audit-current":        {".github", "scripts", "internal/verify"},
	"agent/ci-ownership-audit-current2":       {".github", "scripts", "internal/verify"},
	"agent/ci-alias-refresh":                  {".github", "scripts", "internal/verify"},
	"agent/ci-alias-generator-current7":       {".github", "scripts", "internal/verify"},
	"agent/ci-ownership-audit":                {".github", "scripts", "internal/verify"},
	"agent/ci-scope-triage":                   {".github", "scripts", "internal/verify"},
	"agent/ci-workflow":                       {".github", "scripts", "internal/verify"},
	"agent/ci-evidence-contract":              {".github", "scripts", "internal/verify"},
	"agent/ci-workflow-stage":                 {".github", "scripts", "internal/verify"},
	"agent/cli":                               {"cmd/gooo"},
	"agent/cli-check":                         {"cmd/gooo"},
	"agent/cli-check-current":                 {"cmd/gooo"},
	"agent/cli-check-current2":                {"cmd/gooo"},
	"agent/cli-bootstrap-contract":            {"cmd/gooo"},
	"agent/codegen-followup":                  {"docs/research/codegen-reproducibility.md"},
	"agent/codegen-fixture-adapter":           {"docs/research/codegen-fixture-adapter.md"},
	"agent/codegen-research":                  {"docs/research/codegen.md"},
	"agent/codegen-hypotheses":                {"docs/research/codegen-experiments.md"},
	"agent/conformance-fuzz":                  {"internal/conformance/fuzz", "internal/syntax"},
	"agent/dependency-cycle-detector":         {"internal/detection/cycles"},
	"agent/detection":                         {"internal/detection"},
	"agent/detection-cycles":                  {"internal/detection/cycles"},
	"agent/docs":                              {"AGENTS.md", "CONTRIBUTING.md", "README.md", "docs", "examples"},
	"agent/formatter":                         {"internal/formatter"},
	"agent/freshness-detection":               {"internal/detection/freshness"},
	"agent/freshness-research":                {"internal/research/freshness", "docs/research"},
	"agent/fuzz-conformance":                  {"internal/conformance/fuzz"},
	"agent/generator":                         {"internal/generator"},
	"agent/generator-fixtures-current":        {"internal/generator"},
	"agent/generator-fixtures-current2":       {"internal/generator"},
	"agent/generator-fixtures-current5":       {"internal/generator"},
	"agent/generator-fixtures-current7":       {"internal/generator"},
	"agent/go-version":                        {"go.mod"},
	"agent/grammar-research":                  {"docs/research/grammar.md"},
	"agent/grammar-review":                    {"docs/research/grammar.md"},
	"agent/grammar-followup":                  {"docs/research/grammar.md"},
	"agent/integration-governance":            {"docs/governance"},
	"agent/integration-governance-followup":   {"docs/governance/integration-promotion.md"},
	"agent/lsp-research":                      {"docs/research/lsp.md"},
	"agent/lsp":                               {"internal/lsp"},
	"agent/lsp-contracts":                     {"docs/research/lsp.md"},
	"agent/lsp-experiments":                   {"docs/research/lsp.md"},
	"agent/linecaps":                          {"internal/detection/linecaps"},
	"agent/line-cap-detector":                 {"internal/detection/linecaps"},
	"agent/performance":                       {"internal/detection/performance"},
	"agent/performance-regression":            {"internal/detection/performance"},
	"agent/protected-regions":                 {"internal/detection/protectedregions"},
	"agent/prototype-conformance":             {"internal/conformance"},
	"agent/prototype-detection":               {"internal/detection"},
	"agent/prototype-formatter":               {"internal/formatter"},
	"agent/prototype-provenance":              {"internal/provenance"},
	"agent/prototype-query":                   {"internal/query"},
	"agent/prov-o-research":                   {"docs/research/prov-o.md"},
	"agent/provenance-evidence":               {"internal/provenance"},
	"agent/provenance-freshness-detector":     {"internal/detection/freshness"},
	"agent/provenance-store":                  {"internal/provenance"},
	"agent/query-engine":                      {"internal/query"},
	"agent/query-research":                    {"docs/research/query.md"},
	"agent/roundtrip-detector":                {"internal/detection/roundtrip"},
	"agent/roundtrip-detection":               {"internal/detection/roundtrip"},
	"agent/security-research":                 {"docs/research/security.md"},
	"agent/security":                          {"docs/research/security.md"},
	"agent/semantic":                          {"internal/semantic"},
	"agent/semantic-delta-detector":           {"internal/detection/semanticdelta"},
	"agent/semanticdelta":                     {"internal/detection/semanticdelta"},
	"agent/syntax":                            {"internal/syntax"},
	"agent/self-hosting-bootstrap":            {"docs/research/self-hosting.md", "internal/bootstrap"},
	"agent/testing-research":                  {"docs/research/testing.md"},
	"agent/testing-research-contracts":        {"docs/research/testing.md"},
	"agent/testing-research-followup":         {"docs/research/testing.md"},
	"agent/zerolang-research":                 {"docs/research/zerolang.md"},
	"agent/zerolang-experiments":              {"docs/research/zerolang.md"},
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
