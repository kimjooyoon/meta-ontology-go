package workspaceexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func TestFreshLocalFileReplay12673(t *testing.T) {
	const source = `package replayfiles
namespace replayfiles12673
entity Integer id "urn:gooo:type:integer"
activity Offset(Integer) -> Integer id "urn:replayfiles12673:offset" computes "let folded=input+907;return folded"
activity Build(Integer) -> Integer id "urn:replayfiles12673:build" computes "let base=Offset(input);if input>0 {let value=base+31;return value};let value=base-37;return value"
`
	manifest := packageruntime.Manifest{Schema: packageruntime.ManifestSchema, Entry: packageruntime.EntrySpec{PackagePath: "origin/replay12673", Activity: "Build"}, Packages: []packageruntime.PackageSpec{{Path: "origin/replay12673", Name: "replayfiles", Sources: []packageruntime.Source{{Filename: "branches12673.gooo", Content: source}}}}}
	firstCases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/replay12673:Build":911},"expected":{"origin/replay12673:Build":1849}},{"inputs":{"origin/replay12673:Build":-919},"expected":{"origin/replay12673:Build":-49}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	first, err := ExecuteWorkspace(context.Background(), manifest, firstCases, "", "")
	if err != nil || first.Runtime.FinitePassed != 2 {
		t.Fatal("fresh first execution", err)
	}
	checkReplayFileLocations12673(t, first.Runtime, source)
	saveReplayFile12673(t, "first12673.json", first)
	prior := first
	prior.Runtime.LocalFileSources = nil
	encoded, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &prior); err != nil {
		t.Fatal(err)
	}
	newCases, err := bodyexecution.DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"origin/replay12673:Build":929},"expected":{"origin/replay12673:Build":1867}},{"inputs":{"origin/replay12673:Build":-937},"expected":{"origin/replay12673:Build":-67}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	next, err := ReplayWorkspace(context.Background(), manifest, prior, newCases, "")
	if err != nil || next.Runtime.FinitePassed != 2 || next.Replay == nil || next.Replay.ModelCalls != 0 {
		t.Fatal("new-input replay", err)
	}
	checkReplayFileLocations12673(t, next.Runtime, source)
	if prior.Runtime.LocalFileSources != nil || first.Runtime.LocalFileSources.SHA256 != next.Runtime.LocalFileSources.SHA256 {
		t.Fatal("old absence mutated or new static origins changed")
	}
	saveReplayFile12673(t, "replay12673.json", next)
	t.Log("new branch/helper source first2/2 and replay2/2 on different inputs;shadowed same-name declarations stay distinct;fresh executions acquire file origins while old absent metadata remains nil;model0")
}

func checkReplayFileLocations12673(t *testing.T, runtime bodyexecution.CompositionRuntime, source string) {
	t.Helper()
	if runtime.LocalFileSources == nil {
		t.Fatal("file table absent")
	}
	declarations := map[int]bool{}
	for _, binding := range runtime.LocalFileSources.Bindings {
		for _, loc := range []bodyexecution.CompositionLocalFileLocation{binding.Declaration, binding.Value} {
			if loc.Status != "MAPPED" || loc.Filename != "branches12673.gooo" || loc.SourceSHA256 != sourceSHA256([]byte(source)) || loc.Start < 0 || loc.End <= loc.Start || loc.End > len(source) {
				t.Fatal("physical input source", loc)
			}
		}
		if binding.Declaration.Expression == "value" {
			declarations[binding.Declaration.Start] = true
		}
	}
	if len(declarations) != 2 {
		t.Fatal("same-name lexical declarations merged", declarations)
	}
	rows, _, err := bodyexecution.SelectRecordedNativeOrigins(runtime, bodyexecution.NativeOriginQuery{Kind: "local-input-locations", Offset: 21, Limit: 3})
	if err != nil || len(rows) != 3 {
		t.Fatal("second-case file page", err)
	}
	for _, raw := range rows {
		var fact bodyexecution.CompositionLocalFileFact
		if err := json.Unmarshal(raw, &fact); err != nil || fact.CaseIndex != 1 || fact.Mapping != "MAPPED" {
			t.Fatal("file index did not cross case boundary", err, fact)
		}
	}
}

func saveReplayFile12673(t *testing.T, name string, value any) {
	t.Helper()
	if root := os.Getenv("GOOO_FRESH_CLI_OUTPUT"); root != "" {
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
