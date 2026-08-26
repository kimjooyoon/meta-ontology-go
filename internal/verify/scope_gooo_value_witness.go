package verify

func init() {
	branchScopeAllowlist["agent/gooo-value-witness"] = []string{
		".github/agent-scope-table.md",
		".github/ci-governance.json",
		".github/workflows/language-value-witness.yml",
		"examples/language-value-witness",
		"internal/bidir/lower_part05.go",
		"internal/bidir/value_program.go",
		"internal/bidir/value_program_test.go",
		"internal/syntax/activity_parser.go",
		"internal/syntax/ast_part03.go",
		"internal/syntax/format_part04.go",
		"internal/syntax/value_program_test.go",
		"internal/valueexecution",
		"internal/verify/scope_gooo_value_witness.go",
		"scripts/language-value-witness",
	}
}
