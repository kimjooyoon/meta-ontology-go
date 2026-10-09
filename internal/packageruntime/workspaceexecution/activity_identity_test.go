package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestWorkspaceExplicitActivityIdentitySurvivesNativeExecutionAndReplay(t *testing.T) {
	for _, name := range []string{"Normalize", "정규화"} {
		manifest := importedActivityWorkspace()
		manifest.Packages[0].Sources[0].Content = strings.ReplaceAll(manifest.Packages[0].Sources[0].Content, "Normalize", name)
		source := &manifest.Packages[1].Sources[0].Content
		*source = strings.ReplaceAll(*source, "Normalize", name)
		*source = strings.ReplaceAll(*source, "activity "+name+"(Text) -> Text", "activity "+name+`(Text) -> Text id "urn:gooo:activity:normalize"`)
		key := "example/core:" + name
		value := json.RawMessage(`"한글 and English"`)
		suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
			Inputs: map[string]json.RawMessage{key: value}, Expected: map[string]json.RawMessage{key: value, "example/app:Main": value},
		}}}
		result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
		if err != nil || result.Runtime.FinitePassed != 2 || result.Runtime.FiniteTotal != 2 {
			t.Fatal("native explicit-ID execution failed", err, result.Runtime)
		}
		if !strings.Contains(result.Program.Source, `id "urn:gooo:activity:normalize"`) ||
			!strings.Contains(result.Composition.Source, `id="urn:gooo:activity:normalize"`) {
			t.Fatal("workspace flattening or generated markers lost stable identity")
		}
		if len(result.Runtime.Traces) != 1 || result.Runtime.Traces[0].Deliveries[0].ActivityID != "urn:gooo:activity:normalize" {
			t.Fatal("native delivery trace lost stable identity", result.Runtime.Traces)
		}
		again, err := ReplayWorkspace(context.Background(), manifest, savedWorkspaceResult(t, result), suite, "")
		if err != nil || again.Runtime.FinitePassed != 2 || again.Replay == nil || again.Replay.ModelCalls != 0 {
			t.Fatal("saved receipt did not replay with zero model calls", err, again.Replay)
		}
	}
}
