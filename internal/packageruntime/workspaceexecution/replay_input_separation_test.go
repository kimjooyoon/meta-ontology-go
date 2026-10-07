package workspaceexecution

import (
	"context"
	"encoding/json"
	"maps"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestReplayWorkspaceRecomputesEarlierFillInputSeparation(t *testing.T) {
	manifest, suite := workspaceInputsFixture(t)
	ctx := context.Background()
	prior, err := ExecuteWorkspace(ctx, manifest, suite, "", "")
	if err != nil {
		t.Fatal(err)
	}
	stages := prior.Runtime.InputSeparation.EarlierStages
	prior.Runtime.InputSeparation.DisjointInputs = 999
	prior.Runtime.InputSeparation.DisjointCasesPassed = 999
	prior = savedWorkspaceResult(t, prior)
	for _, mode := range []string{"pass", "duplicate-failure", "inputs-only"} {
		t.Run(mode, func(t *testing.T) {
			current := suite
			current.Cases = append([]bodyexecution.CompositionCase(nil), suite.Cases...)
			if mode == "duplicate-failure" {
				current.Cases[3].Expected = maps.Clone(current.Cases[3].Expected)
				current.Cases[3].Expected["example/construction:Lift"] = json.RawMessage("999")
			} else if mode == "inputs-only" {
				current.Schema = bodyexecution.CompositionInputsSchema
				for i := range current.Cases {
					current.Cases[i].Expected = nil
				}
			}
			again, err := ReplayWorkspace(ctx, manifest, prior, current, "")
			if err != nil {
				t.Fatal(err)
			}
			checkWorkspaceInputCounts(t, again.Runtime.InputSeparation, mode)
			if again.Replay.ModelCalls != 0 || again.Runtime.ModelCalls != 0 || !again.Runtime.RuntimeReplayed || len(again.Runtime.Runs) != 2 {
				t.Fatal("replay did not perform fresh native observations without inference", again.Replay)
			}
			if !reflect.DeepEqual(stages, again.Runtime.InputSeparation.EarlierStages) {
				t.Fatal("replay changed the construction input identity")
			}
		})
	}
}
