package declaredtest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCapturedEventsKeepExistingTypedCollectionLimit(t *testing.T) {
	var raw strings.Builder
	for i := range 257 {
		fmt.Fprintf(&raw, "{\"Action\":\"pass\",\"Test\":\"Test%d\"}\n", i)
	}
	exit := 0
	capture := Capture{TrialID: "bound", SourceRevision: "fixture:bound", ExitCode: &exit,
		Completed: true, StdoutComplete: true, Stdout: []byte(raw.String())}
	plan := json.RawMessage(`{"trial_id":"bound","source_revision":"fixture:bound","process_exit":0,
"tests":[{"test_id":"Test0","outcome":"pass"}]}`)
	result, err := Assess(context.Background(), plan, capture)
	if err == nil || len(result.Value) != 0 {
		t.Fatal("oversized transport reached source execution or returned an accepted prefix", result, err)
	}
}
