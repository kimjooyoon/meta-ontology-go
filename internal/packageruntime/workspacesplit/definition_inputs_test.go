package workspacesplit_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceinput"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func TestDefinitionChangeOrReadFailureCannotReturnPartialPlan(t *testing.T) {
	for _, mode := range []string{"same-meaning-new-bytes", "entry-change", "read-error", "malformed", "unknown-field", "oversize"} {
		t.Run(mode, func(t *testing.T) {
			files, owners := sharedOwners(t)
			requests := storedOwners(t, files, owners)
			original := append([]byte(nil), files[requests[1].ManifestPath].Data...)
			if mode == "malformed" {
				files[requests[1].ManifestPath].Data = []byte("{")
			}
			if mode == "unknown-field" {
				files[requests[1].ManifestPath].Data = append([]byte(`{"unexpected":true,`), original[1:]...)
			}
			if mode == "oversize" {
				files[requests[1].ManifestPath].Data = make([]byte, workspaceinput.MaxDefinitionBytes+1)
			}
			count := 0
			reader := changingReader{read: func(path string) ([]byte, error) {
				if path == requests[1].ManifestPath {
					count++
					if count == 2 {
						if mode == "same-meaning-new-bytes" {
							return append(original, '\n'), nil
						}
						if mode == "entry-change" {
							return files[requests[0].ManifestPath].Data, nil
						}
						if mode == "read-error" {
							return nil, os.ErrNotExist
						}
					}
				}
				return files.ReadFile(path)
			}}
			plan, err := workspacesplit.PrepareFiles(reader, requests)
			if err == nil || !reflect.DeepEqual(plan, workspacesplit.Plan{}) {
				t.Fatal("partial plan", plan, err)
			}
			if (mode == "entry-change" || mode == "same-meaning-new-bytes") && !errors.Is(err, workspacesplit.ErrDefinitionChanged) {
				t.Fatal("lost definition origin", err)
			}
			if len(files) != 3 {
				t.Fatal("preparation wrote files")
			}
		})
	}
}

func TestStoredDefinitionsBoundWorkBeforeSourcePreparation(t *testing.T) {
	files, owners := sharedOwners(t)
	requests := storedOwners(t, files, owners)
	raw := append([]byte(nil), files[requests[0].ManifestPath].Data...)
	raw = append(raw, []byte(strings.Repeat(" ", workspaceinput.MaxDefinitionBytes-len(raw)))...)
	many := make([]workspacesplit.FileOwner, 17)
	for i := range many {
		name := fmt.Sprintf("owner-%d.json", i)
		files[name] = &fstest.MapFile{Data: raw}
		many[i] = workspacesplit.FileOwner{ManifestPath: name, Request: requests[0].Request}
	}
	sourceReads := 0
	reader := changingReader{read: func(path string) ([]byte, error) {
		if path == "shared.gooo" {
			sourceReads++
		}
		return files.ReadFile(path)
	}}
	plan, err := workspacesplit.PrepareFiles(reader, many)
	if err == nil || !strings.Contains(err.Error(), "aggregate byte limit") || !reflect.DeepEqual(plan, workspacesplit.Plan{}) || sourceReads != 0 {
		t.Fatal(plan, err, sourceReads)
	}
	for _, requested := range [][]workspacesplit.FileOwner{nil, make([]workspacesplit.FileOwner, workspacesplit.MaxOwners+1)} {
		if plan, err := workspacesplit.PrepareFiles(reader, requested); err == nil || !reflect.DeepEqual(plan, workspacesplit.Plan{}) {
			t.Fatal(plan, err)
		}
	}
	if plan, err := workspacesplit.PrepareFiles(nil, requests); err == nil || !reflect.DeepEqual(plan, workspacesplit.Plan{}) {
		t.Fatal(plan, err)
	}
}
