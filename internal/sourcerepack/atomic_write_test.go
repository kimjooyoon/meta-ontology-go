package sourcerepack

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFailedTemporaryWritesDoNotPublishEitherSource(t *testing.T) {
	for _, failedAt := range []int{0, 1, 2} {
		t.Run(fmt.Sprint(failedAt), func(t *testing.T) {
			plan := stagingFixture(t)
			calls := 0
			create := func(directory, pattern string) (*os.File, error) {
				calls++
				if calls == failedAt {
					return readOnlyTemporary(directory, pattern)
				}
				return os.CreateTemp(directory, pattern)
			}
			err := applyRepackWithTemp(plan, create)
			if (err == nil) != (failedAt == 0) {
				t.Fatal("temporary write failure lost its error", err)
			}
			for _, edit := range plan.Edits {
				want := edit.Before
				if failedAt == 0 {
					want = edit.After
				}
				got, err := os.ReadFile(edit.Path)
				if err != nil || !bytes.Equal(got, want) {
					t.Fatal("publication did not retain its expected source", edit.Subject, err)
				}
			}
			entries, err := os.ReadDir(filepath.Dir(plan.Edits[0].Path))
			if err != nil || len(entries) != 2 {
				t.Fatal("staged temporary file remained", entries, err)
			}
		})
	}
}
