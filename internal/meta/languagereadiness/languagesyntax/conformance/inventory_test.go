package languagesyntax_test

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

// Inventory is an observation of this source snapshot, not a frozen measure
// of language completeness. Scan each input independently of the producer.
func assertSourceInventory(t *testing.T, repository fs.FS, source languagesyntax.Source, reportedLines int) {
	t.Helper()
	seen := map[string]bool{}
	total := 0
	for _, observed := range source.GoooFiles {
		if seen[observed.Path] {
			t.Fatalf("source inventory repeats %q", observed.Path)
		}
		seen[observed.Path] = true
		raw, err := fs.ReadFile(repository, observed.Path)
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(raw))
		scanner.Buffer(make([]byte, 1024), len(raw)+1)
		lines := 0
		for scanner.Scan() {
			lines++
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
		digest := fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
		if observed.GoooLines != lines || observed.SourceDigest != digest {
			t.Fatalf("source inventory does not match %q: lines=%d want=%d digest=%q want=%q",
				observed.Path, observed.GoooLines, lines, observed.SourceDigest, digest)
		}
		total += lines
	}
	if reportedLines != total {
		t.Fatalf("reported source lines=%d, observed input lines=%d", reportedLines, total)
	}
}
