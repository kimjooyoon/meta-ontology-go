package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSemanticContextRejectsInvalidOptions(t *testing.T) {
	for _, extra := range [][]string{
		{"--depth", "-1"}, {"--depth", "9"}, {"--max-nodes", "0"}, {"--max-nodes", "257"},
		{"--max-bytes", "1023"}, {"--max-bytes", "1048577"}, {"--depth", "1", "--depth", "2"},
		{"--expect-source", "x", "--expect-source", "y"}, {"--unknown", "1"}, {"--depth"},
	} {
		args := append([]string{"context.gooo", "billing://entity/order"}, extra...)
		var out, diagnostics bytes.Buffer
		code := runSemanticContext(args, fixtureReader{}, SyntaxSourceParser{}, &out, &diagnostics)
		if code != exitUsage || out.Len() != 0 {
			t.Fatalf("accepted invalid options %v: %d %s", extra, code, &diagnostics)
		}
	}
}

func TestSemanticContextNodeBudgetOnLargerGraph(t *testing.T) {
	source := sourceWithRuntimeBinding
	for i := range 80 {
		source += fmt.Sprintf("\nactivity Step%d(Order) -> Payment\n", i)
	}
	report, raw := contextFixture(t, source, "--depth", "8", "--max-nodes", "16")
	if len(report.Nodes) != 16 || !report.Truncated || report.OmittedAdjacentNodes == 0 || len(raw) > report.Limits.MaxBytes {
		t.Fatalf("larger graph failed to retain an honest bounded view: %s", raw)
	}
}
