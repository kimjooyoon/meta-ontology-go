package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func retainedFieldFixture(t *testing.T) ([]byte, CompositionCases, Composition) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/record-field-assembly-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "fields.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	return source, suite, prior
}

func TestCompositionExecutorReusesBuildAndObservesCurrentRecordCases(t *testing.T) {
	source, suite, prior := retainedFieldFixture(t)
	executor := NewExecutor()
	defer executor.Close()
	first, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	if first.Artifact == nil || first.Artifact.Reused || !first.Build.Completed || first.FinitePassed != 14 ||
		first.Schema != "gooo/body-composition-runtime/v2" || first.ToolchainReference == nil {
		t.Fatal(first)
	}
	root := executor.artifact.root
	*first.Artifact.SourceBuild.ExitCode = 99
	suite.Cases = []CompositionCase{{
		Inputs:   map[string]json.RawMessage{"Select.input0": json.RawMessage(`{"title":"다른 입력","state":"queued","reason":"새 실행"}`), "Select.input1": json.RawMessage(`true`)},
		Expected: map[string]json.RawMessage{"Select": json.RawMessage(`{"title":"다른 입력","state":"ready","reason":"새 실행:accepted"}`), "Label": json.RawMessage(`"different expectation"`)},
	}}
	second, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	if !second.Artifact.Reused || second.Build.Started || *second.Artifact.SourceBuild.ExitCode != 0 ||
		second.FinitePassed != 1 || second.FiniteTotal != 2 || second.RuntimeSuiteSHA256 == first.RuntimeSuiteSHA256 ||
		len(second.Runs) != 2 || second.ModelCalls != 0 || second.Toolchain.Started || !second.ToolchainReference.Reused ||
		!bytes.Contains(second.Traces[0].Deliveries[0].Actual, []byte("다른 입력")) {
		t.Fatal(second)
	}
	if err := executor.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("closed executor retained workspace", err)
	}
	if _, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool()); !errors.Is(err, ErrExecutorClosed) {
		t.Fatal(err)
	}
}

func TestCompositionExecutorRebuildsChangedProgramAndRetainsCurrentValues(t *testing.T) {
	source, suite, prior := retainedFieldFixture(t)
	executor := NewExecutor()
	defer executor.Close()
	first, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	oldRoot := executor.artifact.root
	source = []byte(strings.Replace(string(source), "return input.title +", `return "changed:" + input.title +`, 1))
	for _, test := range suite.Cases {
		var label string
		if err := json.Unmarshal(test.Expected["Label"], &label); err != nil {
			t.Fatal(err)
		}
		test.Expected["Label"], err = json.Marshal("changed:" + label)
		if err != nil {
			t.Fatal(err)
		}
	}
	prior, err = GenerateComposition(context.Background(), "fields.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	if second.Artifact.Reused || !second.Build.Completed || second.Artifact.MissReason != "key_changed" ||
		second.Artifact.KeySHA256 == first.Artifact.KeySHA256 || second.FinitePassed != 14 || second.FiniteTotal != 14 {
		t.Fatal(second)
	}
	if _, err := os.Stat(oldRoot); !os.IsNotExist(err) {
		t.Fatal("previous graph workspace retained", err)
	}
}

func TestCompositionExecutorGateCancelsAndRejectsUninitializedCalls(t *testing.T) {
	source, suite, prior := retainedFieldFixture(t)
	executor := NewExecutor()
	<-executor.gate
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	r, err := executor.ExecuteComposition(ctx, "fields.gooo", source, prior, suite, nativeTool())
	executor.gate <- struct{}{}
	if !errors.Is(err, context.DeadlineExceeded) || r.Stage != "EXECUTOR_GATE" || r.Build.Started {
		t.Fatal(r, err)
	}
	if err := executor.Close(); err != nil {
		t.Fatal(err)
	}
	for _, owner := range []*Executor{nil, {}} {
		if _, err := owner.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool()); err == nil {
			t.Fatal("uninitialized executor accepted")
		}
	}
}

func TestCompositionExecutorSerializesConcurrentCurrentRecordRequests(t *testing.T) {
	source, suite, prior := retainedFieldFixture(t)
	executor := NewExecutor()
	defer executor.Close()
	if _, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool()); err != nil {
		t.Fatal(err)
	}
	type completed struct {
		result CompositionRuntime
		err    error
		title  string
	}
	done := make(chan completed, 4)
	for i := range 4 {
		go func() {
			title := fmt.Sprintf("caller-%d", i)
			input, _ := json.Marshal(map[string]string{"title": title, "state": "queued", "reason": "work"})
			expected, _ := json.Marshal(map[string]string{"title": title, "state": "ready", "reason": "work:accepted"})
			label, _ := json.Marshal(title + ":ready:work:accepted")
			current := CompositionCases{Schema: suite.Schema, Cases: []CompositionCase{{
				Inputs:   map[string]json.RawMessage{"Select.input0": input, "Select.input1": json.RawMessage(`true`)},
				Expected: map[string]json.RawMessage{"Select": expected, "Label": label},
			}}}
			r, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, current, nativeTool())
			done <- completed{r, err, title}
		}()
	}
	for range 4 {
		select {
		case call := <-done:
			if call.err != nil || !call.result.Artifact.Reused || call.result.Build.Started ||
				call.result.FinitePassed != 2 || !bytes.Contains(call.result.Traces[0].Deliveries[0].Actual, []byte(call.title)) {
				t.Fatal(call)
			}
		case <-time.After(30 * time.Second):
			t.Fatal("concurrent composition requests did not finish")
		}
	}
}

func TestCompositionExecutorCloseReleasesWorkspaceWhileCallersWait(t *testing.T) {
	source, suite, prior := retainedFieldFixture(t)
	executor := NewExecutor()
	if _, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool()); err != nil {
		t.Fatal(err)
	}
	root := executor.artifact.root
	<-executor.gate
	closed := make(chan error, 1)
	go func() { closed <- executor.Close() }()
	select {
	case <-executor.lifetime.Done():
	case <-time.After(time.Second):
		t.Fatal("close did not cancel waiting lifetime")
	}
	if _, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, suite, nativeTool()); !errors.Is(err, ErrExecutorClosed) {
		t.Fatal(err)
	}
	executor.gate <- struct{}{}
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("close did not finish after gate release")
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("closed graph workspace retained", err)
	}
}
