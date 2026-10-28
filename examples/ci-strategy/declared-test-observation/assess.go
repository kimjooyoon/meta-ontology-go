// Package declaredtest transports named Go test observations to Gooo's
// source-owned assessment. It grants no CI or repository-write authority.
package declaredtest

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/testevents"
)

//go:embed main.gooo
var source []byte

const sourcePath = "examples/ci-strategy/declared-test-observation/main.gooo"

// Capture carries the caller's observations. Completed means the process
// exited, including a nonzero exit; StdoutComplete describes the capture.
// SourceRevision and Arguments are supplied observations, not inferred facts.
type Capture struct {
	TrialID, SourceRevision string
	Arguments               []string
	ExitCode                *int
	Completed               bool
	StdoutComplete          bool
	Stdout                  []byte
}

// Assess preserves the caller's Capture. On a malformed suffix it evaluates
// the decoded prefix and returns both the source result and the decoding error.
func Assess(ctx context.Context, plan json.RawMessage, captured Capture) (bodycodegen.FixedValue, error) {
	events, decodeErr := testevents.Decode(captured.Stdout)
	tests := make([]map[string]string, 0, len(events))
	for _, event := range events {
		tests = append(tests, map[string]string{"test_id": event.Name, "outcome": event.Action})
	}
	exit := -1
	if captured.ExitCode != nil {
		exit = *captured.ExitCode
	}
	observation := map[string]any{
		"trial_id": captured.TrialID, "source_revision": captured.SourceRevision,
		"process_exit": exit, "completed": captured.Completed && captured.ExitCode != nil,
		"stream_complete": captured.StdoutComplete && decodeErr == nil,
		"argv_captured":   len(captured.Arguments) != 0, "tests": tests,
	}
	raw, err := json.Marshal([]any{map[string]any{"plan": plan, "observation": observation}})
	if err != nil {
		return bodycodegen.FixedValue{}, errors.Join(decodeErr, err)
	}
	program, err := bodycodegen.NewFixedProgram(ctx, sourcePath, source, "Assess")
	if err != nil {
		return bodycodegen.FixedValue{}, errors.Join(decodeErr, err)
	}
	result, err := program.Evaluate(ctx, raw)
	return result, errors.Join(decodeErr, err)
}
