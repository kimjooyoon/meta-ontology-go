package workspaceexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestImportedScalarIdentitySurvivesNativeExecutionAndReplay(t *testing.T) {
	manifest := importedActivityWorkspace()
	for i := range manifest.Packages {
		source := &manifest.Packages[i].Sources[0].Content
		*source = strings.ReplaceAll(strings.ReplaceAll(*source, "example://core/text", "urn:gooo:type:string"), "Text", "문자열")
	}
	value := json.RawMessage(`"한글 and English"`)
	suite := bodyexecution.CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []bodyexecution.CompositionCase{{
		Inputs: map[string]json.RawMessage{"example/core:Normalize": value}, Expected: map[string]json.RawMessage{"example/core:Normalize": value, "example/app:Main": value},
	}}}
	result, err := ExecuteWorkspace(context.Background(), manifest, suite, "", "")
	if err != nil || result.Runtime.FinitePassed != 2 || result.Runtime.FiniteTotal != 2 {
		t.Fatal("imported scalar alias did not execute", err)
	}
	for _, node := range result.Composition.Plan.Activities {
		if node.InputEntityID != "urn:gooo:type:string" || node.OutputEntityID != "urn:gooo:type:string" {
			t.Fatal("imported scalar identity changed", node)
		}
	}
	again, err := ReplayWorkspace(context.Background(), manifest, savedWorkspaceResult(t, result), suite, "")
	if err != nil || again.Runtime.FinitePassed != 2 || again.Replay == nil || again.Replay.ModelCalls != 0 {
		t.Fatal("imported scalar alias saved replay differed", err)
	}
}
