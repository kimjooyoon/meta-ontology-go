package adkcheck

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/adk/v2/agent"
)

type fixtureAgentContext struct {
	*agent.ContextMock
	context.Context
}

func (c fixtureAgentContext) Deadline() (time.Time, bool) { return c.Context.Deadline() }
func (c fixtureAgentContext) Done() <-chan struct{}       { return c.Context.Done() }
func (c fixtureAgentContext) Err() error                  { return c.Context.Err() }
func (c fixtureAgentContext) Value(key any) any           { return c.Context.Value(key) }

func expectFailure(t *testing.T, err error, stage, code string) {
	t.Helper()
	f, ok := errors.AsType[*Failure](err)
	if !ok || f.Stage != stage || f.Code != code {
		t.Fatalf("failure = %#v, want %s/%s", err, stage, code)
	}
	if f.Error() != "gooo-adk/"+stage+"/"+code {
		t.Fatalf("public failure text = %q", f.Error())
	}
}

func TestInputFailuresHaveStableStagesAndCodes(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.gooo"), []byte("package p\nnamespace p\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = check(context.Background(), "unused", root, time.Second, CheckArgs{Path: "../outside.gooo"})
	expectFailure(t, err, StageInput, CodeInvalidPath)
	_, err = discover(context.Background(), "unused", root, time.Second, DiscoverArgs{Query: ""})
	expectFailure(t, err, StageInput, CodeInvalidQuery)
	_, err = discover(context.Background(), "unused", root, time.Second, DiscoverArgs{Path: "missing.gooo", Query: "capability"})
	expectFailure(t, err, StageInput, CodeInvalidPath)
	_, err = discover(context.Background(), "unused", root, time.Second, DiscoverArgs{
		Path: "source.gooo", Query: "capability", DomainContract: "missing.gooo",
	})
	expectFailure(t, err, StageInput, CodeInvalidContractPath)
}

func TestExecutionAndOutputFailuresHaveStableStagesAndCodes(t *testing.T) {
	root := t.TempDir()
	fixture := filepath.Join(root, "fixture.sh")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(fixture, []byte("#!/bin/sh\n"+body+"\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	_, err := runJSONCommand(context.Background(), filepath.Join(root, "missing"), root, time.Second)
	expectFailure(t, err, StageExecution, CodeStartFailed)
	write("printf 'nope'")
	_, err = runJSONCommand(context.Background(), fixture, root, time.Second)
	expectFailure(t, err, StageOutput, CodeInvalidJSON)
	write("head -c 1048577 /dev/zero")
	_, err = runJSONCommand(context.Background(), fixture, root, time.Second)
	expectFailure(t, err, StageOutput, CodeOutputLimit)
	if !errors.Is(err, errReportTooLarge) {
		t.Fatalf("output limit cause was lost: %v", err)
	}
	_, err = runJSONCommand(context.Background(), "/bin/sleep", root, time.Millisecond, "1")
	expectFailure(t, err, StageExecution, CodeDeadline)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline cause was lost: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = runJSONCommand(ctx, fixture, root, time.Second)
	expectFailure(t, err, StageExecution, CodeCancelled)
	write("kill -TERM $$")
	_, err = runJSONCommand(context.Background(), fixture, root, time.Second)
	expectFailure(t, err, StageExecution, CodeTerminated)
	write("printf '{\"status\":\"error\"}'; exit 7")
	result, err := runJSONCommand(context.Background(), fixture, root, time.Second)
	if err != nil || result.ExitCode != 7 || !strings.Contains(result.ReportJSON, `"status":"error"`) {
		t.Fatalf("compiler diagnostic became tool failure: result=%#v err=%v", result, err)
	}
}

func TestConfigurationFailuresHaveStableCodes(t *testing.T) {
	_, _, _, err := resolveConfig(Config{SourceRoot: t.TempDir()})
	expectFailure(t, err, StageConfiguration, CodeInvalidBinary)
	_, _, _, err = resolveConfig(Config{GoooBinary: "/bin/true"})
	expectFailure(t, err, StageConfiguration, CodeInvalidSourceRoot)
	_, _, _, err = resolveConfig(Config{GoooBinary: "/bin/true", SourceRoot: filepath.Join(t.TempDir(), "missing")})
	expectFailure(t, err, StageConfiguration, CodeUnavailableRoot)
}

func TestADKFunctionToolPreservesBoundaryFailure(t *testing.T) {
	root := t.TempDir()
	tool, err := NewDiscover(Config{GoooBinary: "/bin/true", SourceRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	callable, ok := tool.(interface {
		Run(agent.Context, any) (map[string]any, error)
	})
	if !ok {
		t.Fatalf("ADK tool %T is not callable", tool)
	}
	result, err := callable.Run(&agent.ContextMock{}, map[string]any{"path": "source.gooo", "query": ""})
	if result != nil {
		t.Fatalf("ADK returned a success map for rejected input: %#v", result)
	}
	expectFailure(t, err, StageInput, CodeInvalidQuery)
}

func TestADKFunctionToolPreservesCompilerDiagnosticAsData(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.gooo"), []byte("package p\nnamespace p\n"), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(root, "fixture.sh")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '{\"status\":\"error\"}'; exit 4\n"), 0700); err != nil {
		t.Fatal(err)
	}
	tool, err := New(Config{GoooBinary: binary, SourceRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	callable := tool.(interface {
		Run(agent.Context, any) (map[string]any, error)
	})
	ctx := fixtureAgentContext{ContextMock: &agent.ContextMock{}, Context: context.Background()}
	result, err := callable.Run(ctx, map[string]any{"path": "source.gooo"})
	if err != nil || result["report_json"] != `{"status":"error"}` || fmt.Sprint(result["exit_code"]) != "4" {
		t.Fatalf("ADK changed compiler result: result=%#v err=%v", result, err)
	}
}
