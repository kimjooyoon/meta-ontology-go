package workspacesplit_test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func discoveryFixture(t *testing.T) fstest.MapFS {
	t.Helper()
	files, definition, _ := sharedFixture(t)
	raw, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	files["first.workspace.json"] = &fstest.MapFile{Data: raw}
	files["second.json"] = &fstest.MapFile{Data: raw}
	files["cases.json"] = &fstest.MapFile{Data: []byte(`[{"input":1}]`)}
	return files
}

func TestDiscoveryFindsSameOccurrenceAndRetainsOtherDefinitions(t *testing.T) {
	files := discoveryFixture(t)
	files["copy.gooo"] = &fstest.MapFile{Data: files["shared.gooo"].Data}
	files["unrelated.json"] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(
		string(files["first.workspace.json"].Data), "shared.gooo", "copy.gooo"))}
	got, err := workspacesplit.DiscoverFiles(files, "shared.gooo")
	if err != nil || !reflect.DeepEqual(got.Consumers, []string{"first.workspace.json", "second.json"}) ||
		len(got.Definitions) != 3 || got.JSONFiles != 4 {
		t.Fatal("consumer population or nonconsumer origin lost", got, err)
	}
	// Saving a proposal inside this scope must not invalidate its own catalog.
	files["proposal.json"] = &fstest.MapFile{Data: []byte(`{"schema":"gooo/workspace-source-split-receipt/v1"}`)}
	again, err := workspacesplit.DiscoverFiles(files, "shared.gooo")
	if err != nil || !reflect.DeepEqual(again.Definitions, got.Definitions) ||
		!reflect.DeepEqual(again.Consumers, got.Consumers) || again.JSONFiles != got.JSONFiles+1 {
		t.Fatal("non-workspace output changed the definition catalog", again, err)
	}
}

func TestDiscoveryNeverReturnsPartialOwnersForUnreadableScope(t *testing.T) {
	cases := map[string]*fstest.MapFile{
		"trailing.json":  {Data: []byte(`{} {}`)},
		"duplicate.json": {Data: []byte(`{"schema":"other","schema":"hidden"}`)},
		"large.json":     {Data: []byte(`"` + strings.Repeat("x", 1<<20) + `"`)},
		"linked":         {Data: []byte("."), Mode: fs.ModeSymlink},
		"invalid.json":   {Data: []byte(`{"schema":"gooo/package-workspace-manifest/v1"}`)},
		"future.json":    {Data: []byte(`{"schema":"gooo/package-workspace-manifest/v2"}`)},
	}
	for name, file := range cases {
		t.Run(name, func(t *testing.T) {
			files := discoveryFixture(t)
			files["z-"+name] = file
			got, err := workspacesplit.DiscoverFiles(files, "shared.gooo")
			if err == nil || !reflect.DeepEqual(got, workspacesplit.Discovery{}) {
				t.Fatal("unreadable tail returned a usable consumer prefix", got, err)
			}
		})
	}
}

func TestDiscoveryCannotHideMissingOrCaseAliasedReferences(t *testing.T) {
	for _, name := range []string{"missing.gooo", "SHARED.gooo", "../shared.gooo"} {
		files := discoveryFixture(t)
		files["second.json"].Data = []byte(strings.ReplaceAll(string(files["second.json"].Data), "shared.gooo", name))
		if _, err := workspacesplit.DiscoverFiles(files, "shared.gooo"); err == nil {
			t.Fatal("unobserved reference was silently ignored", name)
		}
	}
}

func TestDiscoveryResolvesEachDefinitionRelativeToItsOwnDirectory(t *testing.T) {
	files := discoveryFixture(t)
	files["nested/shared.gooo"] = files["shared.gooo"]
	delete(files, "shared.gooo")
	files["nested/second.json"] = files["second.json"]
	delete(files, "second.json")
	files["first.workspace.json"].Data = []byte(strings.ReplaceAll(
		string(files["first.workspace.json"].Data), "shared.gooo", "nested/shared.gooo"))
	got, err := workspacesplit.DiscoverFiles(files, "nested/shared.gooo")
	if err != nil || !reflect.DeepEqual(got.Consumers, []string{"first.workspace.json", "nested/second.json"}) {
		t.Fatal("shared occurrence lost manifest-relative origins", got, err)
	}
}

func TestDiscoveryBoundsCandidateEnumerationBeforeReading(t *testing.T) {
	files := discoveryFixture(t)
	for index := range 4097 {
		files[fmt.Sprintf("candidates/%04d.json", index)] = &fstest.MapFile{Data: []byte(`{}`)}
	}
	got, err := workspacesplit.DiscoverFiles(files, "shared.gooo")
	if err == nil || !reflect.DeepEqual(got, workspacesplit.Discovery{}) {
		t.Fatal("excessive candidate catalog accepted", got, err)
	}
}
