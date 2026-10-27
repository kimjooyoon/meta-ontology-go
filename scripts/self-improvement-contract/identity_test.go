package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func contractIdentitySource(t *testing.T) (string, string) {
	t.Helper()
	path := filepath.Join("..", "..", "examples", "self-improvement", "main.gooo")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return path, string(raw)
}

func renameContractLabels(t *testing.T, source string, names map[string]string) string {
	t.Helper()
	tokens, diagnostics := syntax.Lex(source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	for _, token := range slices.Backward(tokens) {
		if name := names[token.Text]; token.Kind == syntax.TokenIdentifier && name != "" {
			source = source[:token.Span.Start.Offset] + name + source[token.Span.End.Offset:]
		}
	}
	return source
}

func TestExplicitContractActivityIDsPreserveExistingMeaning(t *testing.T) {
	path, source := contractIdentitySource(t)
	implicit := regexp.MustCompile(` id "meta://activity/[a-z-]+"`).ReplaceAllString(source, "")
	if implicit == source {
		t.Fatal("expected authored activity IDs")
	}
	before, after := buildReport(path, []byte(implicit), strings.Repeat("a", 40)), buildReport(path, []byte(source), strings.Repeat("a", 40))
	if before.Status != "PASS" || after.Status != "PASS" || before.SemanticHash != after.SemanticHash || before.SourceSHA256 == after.SourceSHA256 {
		t.Fatal("explicit IDs changed the existing source meaning", before.Status, after.Status)
	}
}

func TestContractNamesCanChangeWithoutChangingMeaning(t *testing.T) {
	path, source := contractIdentitySource(t)
	baseline := buildReport(path, []byte(source), strings.Repeat("a", 40))
	file, diagnostics := syntax.ParseFile(path, source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	names := map[string]string{}
	for index, declaration := range file.Decls {
		name := ""
		switch declaration := declaration.(type) {
		case *syntax.EntityDecl:
			name = declaration.Name
		case *syntax.ActivityDecl:
			name = declaration.Name
		}
		if name != "" {
			names[name] = fmt.Sprintf("의미_%d", index)
		}
	}
	for name, replacement := range names {
		t.Run(name, func(t *testing.T) {
			renamed := renameContractLabels(t, source, map[string]string{name: replacement})
			checkRenamedContract(t, path, renamed, baseline)
		})
	}
	t.Run("all-labels", func(t *testing.T) {
		checkRenamedContract(t, path, renameContractLabels(t, source, names), baseline)
	})
}

func checkRenamedContract(t *testing.T, path, source string, baseline Report) {
	t.Helper()
	result := buildReport(path, []byte(source), baseline.CommitSHA)
	if result.Status != "PASS" || result.SemanticHash != baseline.SemanticHash || result.SourceSHA256 == baseline.SourceSHA256 {
		t.Fatalf("display rename changed semantic admission: status=%s errors=%v", result.Status, result.Errors)
	}
	for _, coverage := range result.ExecutorCoverage {
		origin := coverage.InputOrigin
		if !coverage.Covered || coverage.ActivityID != "meta://activity/execute" || coverage.ActivityName == "" || origin == nil ||
			origin.File != path || origin.Start.Offset < 0 || origin.End.Offset > len(source) || origin.End.Offset <= origin.Start.Offset ||
			source[origin.Start.Offset:origin.End.Offset] != coverage.EntityName {
			t.Fatalf("executor lost its authored input origin: %+v", coverage)
		}
	}
}
