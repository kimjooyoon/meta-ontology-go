package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkspacePureNullaryImportedHelperAndSavedReplay(t *testing.T) {
	manifest, suite := pureCallWorkspace(), pureCallCases(t)
	core := &manifest.Packages[0]
	core.Sources[0].Content = strings.Replace(core.Sources[0].Content, "Triple(input) - 1", "Triple() + input - 1", 1)
	core.Sources[1].Content = strings.Replace(core.Sources[1].Content, "Triple(Integer) -> Integer computes `return input * 3`", "Triple() -> Integer computes `return 7`", 1)
	for i, value := range []string{"14", "2"} {
		suite.Cases[i].Expected["example/app:Main"] = json.RawMessage(value)
	}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 || !result.Runtime.RuntimeReplayed ||
		len(result.Composition.Plan.Activities) != 1 || result.Program.PureCalls == nil ||
		len(result.Program.PureCalls.Activities) != 4 || len(result.Program.PureCalls.Sites) != 3 {
		t.Fatal("nullary imported helper changed the entry inputs or call origins", err)
	}
	raw, _ := json.Marshal(result)
	var saved Result
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	replayed, err := ReplayWorkspace(context.Background(), manifest, saved, suite, "")
	if err != nil || replayed.Runtime.FinitePassed != 2 || replayed.Replay.ModelCalls != 0 {
		t.Fatal("saved nullary helper did not replay", err)
	}
	core.Sources[1].Content = strings.Replace(core.Sources[1].Content, "return 7", "return 8", 1)
	if _, err := ReplayWorkspace(context.Background(), manifest, saved, suite, ""); err == nil {
		t.Fatal("changed nullary helper consumed an old receipt")
	}
}
