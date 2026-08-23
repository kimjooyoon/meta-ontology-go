package toolchaincli

import (
	"strings"

	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

const rootUsage = "usage: gooo <check|generate|roundtrip|query|inspect|graph|analyze|format|fix|provenance|selective-ci|lsp|version> [args]\n"

func inspectGuardrail(operation string, observed cliruntime.Observation) outputEvidence {
	expected := map[string]string{
		"ROOT_USAGE":      rootUsage,
		"UNKNOWN_COMMAND": "gooo: command \"unknown-contract\" is not implemented yet\n",
		"VERSION_USAGE":   "usage: gooo version [--json]\n",
		"CHECK_USAGE":     "usage: gooo check [--semantic] [--provenance-store <ledger.jsonl>] [--json] <file.gooo>\n",
		"ROUNDTRIP_USAGE": "usage: gooo roundtrip [--json] <file.gooo>\n",
		"LSP_USAGE":       "usage: gooo lsp\n",
	}
	result := outputEvidence{stdoutOK: observed.Stdout == "", stderrOK: observed.Stderr == expected[operation]}
	if operation == "ROOT_USAGE" && result.stderrOK {
		result.declaredCommands = declaredCommandCount(observed.Stderr)
	}
	return result
}

func declaredCommandCount(value string) int {
	start, end := strings.Index(value, "<"), strings.Index(value, ">")
	if start < 0 || end <= start {
		return 0
	}
	return len(strings.Split(value[start+1:end], "|"))
}
