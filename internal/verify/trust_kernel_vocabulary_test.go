package verify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProtectedTrustKernelRejectsReviewAuthorityPhrases(t *testing.T) {
	phrasesByPath := map[string][]string{
		".github/branch-policy.md": {
			"reviewed CI-owned change",
			"review boundary",
			"review table",
		},
		".github/conformance-plan.md": {
			"reviewed evidence samples",
			"approved promotion decision",
			"reviewed governance change",
			"gates are reviewed",
		},
		".github/agent-scope-table.md": {
			"review table",
		},
		".github/workflows/ci.yml": {
			"reviewed promotion",
		},
	}
	for relativePath, phrases := range phrasesByPath {
		path := filepath.Join("..", "..", relativePath)
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read protected file %s: %v", relativePath, err)
		}
		for _, phrase := range phrases {
			if strings.Contains(string(contents), phrase) {
				t.Errorf("protected file %s contains forbidden authority phrase %q", relativePath, phrase)
			}
		}
	}
}
