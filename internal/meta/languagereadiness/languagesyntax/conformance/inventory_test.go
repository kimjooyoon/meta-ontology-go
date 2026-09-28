package languagesyntax_test

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax"
)

// Inventory is an observation of this source snapshot, not a frozen measure
// of language completeness. Scan each input independently of the producer.
func assertSourceInventory(t *testing.T, repository fs.FS, source languagesyntax.Source, reportedLines int) {
	t.Helper()
	if err := independentSourceInventoryError(repository, source, reportedLines); err != nil {
		t.Fatal(err)
	}
}

func independentSourceInventoryError(repository fs.FS, source languagesyntax.Source, reportedLines int) error {
	inputs := map[string][]byte{}
	err := fs.WalkDir(repository, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".gooo") {
			return nil
		}
		raw, err := fs.ReadFile(repository, path)
		if err != nil {
			return err
		}
		inputs[path] = raw
		return nil
	})
	if err != nil {
		return err
	}
	total := 0
	for _, observed := range source.GoooFiles {
		raw, exists := inputs[observed.Path]
		if !exists {
			return fmt.Errorf("source inventory repeats or reports an unwalked path %q", observed.Path)
		}
		delete(inputs, observed.Path)
		scanner := bufio.NewScanner(bytes.NewReader(raw))
		scanner.Buffer(make([]byte, 1024), len(raw)+1)
		lines := 0
		for scanner.Scan() {
			lines++
		}
		if err := scanner.Err(); err != nil {
			return err
		}
		digest := fmt.Sprintf("sha256:%x", sha256.Sum256(raw))
		if observed.GoooLines != lines || observed.SourceDigest != digest {
			return fmt.Errorf("source inventory does not match %q: lines=%d want=%d digest=%q want=%q",
				observed.Path, observed.GoooLines, lines, observed.SourceDigest, digest)
		}
		total += lines
	}
	if len(inputs) != 0 {
		return fmt.Errorf("source inventory omits %d independently walked .gooo files", len(inputs))
	}
	if reportedLines != total {
		return fmt.Errorf("reported source lines=%d, observed input lines=%d", reportedLines, total)
	}
	return nil
}

func TestIndependentSourceInventoryRejectsUnreportedFile(t *testing.T) {
	repository := fstest.MapFS{
		"nested/unregistered.gooo": {Data: []byte("package unregistered\nnamespace unregistered\n")},
		"README.md":                {Data: []byte("not a Gooo source\n")},
	}
	err := independentSourceInventoryError(repository, languagesyntax.Source{}, 0)
	if err == nil || !strings.Contains(err.Error(), "omits 1 independently walked .gooo files") {
		t.Fatalf("omitted source was not independently rejected: %v", err)
	}
}
