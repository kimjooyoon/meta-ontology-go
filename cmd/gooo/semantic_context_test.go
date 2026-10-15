package main

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func contextFixture(t *testing.T, source string, extra ...string) (semanticContextReport, []byte) {
	t.Helper()
	var out, diagnostics bytes.Buffer
	args := append([]string{"context.gooo", "billing://entity/order"}, extra...)
	code := runSemanticContext(args, fixtureReader{source: source}, SyntaxSourceParser{}, &out, &diagnostics)
	if code != exitOK {
		t.Fatalf("context exit %d: %s", code, &diagnostics)
	}
	var report semanticContextReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	return report, out.Bytes()
}

func TestSemanticContextFollowsSourceRelationshipsAndBindings(t *testing.T) {
	source := sourceWithRuntimeBinding + "\nentity Unrelated id \"billing://entity/unrelated\"\n"
	report, first := contextFixture(t, source, "--depth", "8")
	_, repeated := contextFixture(t, source, "--depth", "8")
	if !bytes.Equal(first, repeated) || report.Truncated || len(report.Nodes) != 4 || len(report.RuntimeBindings) != 1 {
		t.Fatalf("unexpected closed context: %s", first)
	}
	if report.SourceDigest != semantic.StableHash([]byte(source)) || report.SemanticDigest == "" {
		t.Fatal("missing exact source and semantic identities")
	}
	selected := map[string]bool{}
	for _, node := range report.Nodes {
		selected[node.ID] = true
	}
	for _, relation := range report.Relations {
		if !selected[relation.Subject] || !selected[relation.Object] || relation.Status != "deterministic" {
			t.Fatalf("relation escaped the selected declarations: %#v", relation)
		}
	}
	digest := report.ContextDigest
	report.ContextDigest = ""
	canonical, err := json.Marshal(report)
	if err != nil || semantic.StableHash(canonical) != digest {
		t.Fatal("context identity does not bind the emitted view", err)
	}
}

func TestSemanticContextBudgetsRetainInspectableBoundaries(t *testing.T) {
	for _, tc := range []struct {
		option, value, reason string
	}{{"--depth", "0", "depth"}, {"--max-nodes", "1", "max_nodes"}, {"--max-bytes", "1024", "max_bytes"}} {
		t.Run(tc.option, func(t *testing.T) {
			report, raw := contextFixture(t, sourceWithRuntimeBinding, tc.option, tc.value)
			if !report.Truncated || !slices.Contains(report.Reasons, tc.reason) || report.NextRoot == "" {
				t.Fatalf("missing frontier: %s", raw)
			}
			if len(raw) > report.Limits.MaxBytes || len(report.Nodes) > report.Limits.MaxNodes ||
				report.OmittedAdjacentNodes == 0 {
				t.Fatalf("budget was exceeded or omissions hidden: %s", raw)
			}
			var out, diagnostics bytes.Buffer
			args := []string{"context.gooo", report.NextRoot, "--expect-source", report.SourceDigest}
			code := runSemanticContext(args, fixtureReader{source: sourceWithRuntimeBinding},
				SyntaxSourceParser{}, &out, &diagnostics)
			if code != exitOK {
				t.Fatalf("next root cannot be inspected: %s", &diagnostics)
			}
		})
	}
}

func TestSemanticContextRejectsStaleMissingAndOversizedRootsBeforeOutput(t *testing.T) {
	for _, tc := range []struct {
		source string
		args   []string
	}{
		{sourceWithRuntimeBinding, []string{"context.gooo", "billing://entity/order", "--expect-source", "old"}},
		{sourceWithRuntimeBinding, []string{"context.gooo", "billing://entity/absent"}},
		{
			"package billing\nnamespace billing\nentity " + strings.Repeat("A", 2000) +
				" id \"billing://entity/order\"\n",
			[]string{"context.gooo", "billing://entity/order", "--max-bytes", "1024"},
		},
	} {
		var out, diagnostics bytes.Buffer
		code := runSemanticContext(tc.args, fixtureReader{source: tc.source}, SyntaxSourceParser{}, &out, &diagnostics)
		if code != exitFailure || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatalf("invalid view leaked output: code=%d stdout=%s stderr=%s", code, &out, &diagnostics)
		}
	}
}

func TestSemanticContextCLIAndHelp(t *testing.T) {
	file := t.TempDir() + "/source.gooo"
	if err := os.WriteFile(file, []byte(sourceWithRuntimeBinding), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"context", file, "billing://entity/order"}, {"help", "context"}, {"context", "--help"}} {
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK || out.Len() == 0 {
			t.Fatalf("CLI %v failed: %d %s", args, code, &diagnostics)
		}
	}
}
