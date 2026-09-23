package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestCompletionUsesCanonicalSyntaxKeywords(t *testing.T) {
	items := serverCompletionLabels(NewServer().completion("missing://document"))
	for _, keyword := range syntax.CanonicalKeywordNames() {
		if !items[keyword] {
			t.Fatalf("completion omitted canonical keyword %q", keyword)
		}
	}
}

func serverCompletionLabels(list *CompletionList) map[string]bool {
	labels := make(map[string]bool, len(list.Items))
	for _, item := range list.Items {
		labels[item.Label] = true
	}
	return labels
}
