package workspaceexecution

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func ownedWorkspaceSuite(t *testing.T, input, expected string) bodyexecution.CompositionCases {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": []any{map[string]any{
		"inputs":   map[string]string{"example/core:Normalize": input},
		"expected": map[string]string{"example/core:Normalize": expected, "example/app:Main": expected},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	suite, err := bodyexecution.DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	return suite
}

func ownedWorkspaceTool() string {
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(runtime.GOROOT(), "bin", name)
}

func TestWorkspaceExecutorRetainsBuildWithFreshReplayInputsAndOutcomes(t *testing.T) {
	owner := NewExecutor()
	defer owner.Close()
	manifest := importedActivityWorkspace()
	first, err := owner.Execute(context.Background(), manifest, ownedWorkspaceSuite(t, "alpha", "alpha"),
		ExecuteOptions{GoBinary: ownedWorkspaceTool()})
	if err != nil || first.Runtime.Artifact == nil || first.Runtime.Artifact.Reused || !first.Runtime.Build.Completed {
		t.Fatal("first workspace request did not build", err)
	}
	second, err := owner.Replay(context.Background(), manifest, savedWorkspaceResult(t, first),
		ownedWorkspaceSuite(t, "한글 beta", "한글 beta"), ownedWorkspaceTool())
	assertOwnedWorkspaceReplay(t, first, second, err, 2)
	third, err := owner.Replay(context.Background(), manifest, savedWorkspaceResult(t, second),
		ownedWorkspaceSuite(t, "gamma", "deliberately wrong"), ownedWorkspaceTool())
	assertOwnedWorkspaceReplay(t, second, third, err, 0)
	for _, result := range []Result{first, second, third} {
		if len(result.Runtime.Runs) != 2 || result.Runtime.ModelCalls != 0 || result.Runtime.FiniteTotal != 2 {
			t.Fatal("workspace request did not keep two fresh runs and its current denominator")
		}
	}
}

func assertOwnedWorkspaceReplay(t *testing.T, prior, current Result, err error, passed int) {
	t.Helper()
	r := current.Runtime
	encoded, encodeErr := json.Marshal(prior)
	if err != nil || encodeErr != nil || current.Replay == nil || current.Replay.PriorResultSHA256 != sourceSHA256(encoded) ||
		r.Artifact == nil || !r.Artifact.Reused || !r.Artifact.ExecutableVerified || r.Build.Started ||
		r.Artifact.SourceBuildSHA256 != prior.Runtime.Artifact.SourceBuildSHA256 ||
		r.RuntimeSuiteSHA256 == prior.Runtime.RuntimeSuiteSHA256 || r.FinitePassed != passed {
		t.Fatal("workspace replay reused stale work or lost its parent", err, r.Stage)
	}
}

func TestWorkspaceExecutorRebuildsChangedSourceAndRefusesStaleReplay(t *testing.T) {
	owner := NewExecutor()
	defer owner.Close()
	manifest := importedActivityWorkspace()
	suite := ownedWorkspaceSuite(t, "alpha", "alpha")
	first, err := owner.Execute(context.Background(), manifest, suite, ExecuteOptions{GoBinary: ownedWorkspaceTool()})
	if err != nil {
		t.Fatal(err)
	}
	manifest.Packages[0].Sources[0].Content = strings.Replace(manifest.Packages[0].Sources[0].Content,
		"computes "+strconv.Quote("return input"), "computes "+strconv.Quote(`return input + ":changed"`), 1)
	if _, err := owner.Replay(context.Background(), manifest, first, suite, "must-not-execute"); !errors.Is(err, ErrReplaySourceMismatch) {
		t.Fatal("changed workspace did not retain its typed replay mismatch", err)
	}
	suite.Cases[0].Expected["example/app:Main"] = json.RawMessage(`"alpha:changed"`)
	second, err := owner.Execute(context.Background(), manifest, suite, ExecuteOptions{GoBinary: ownedWorkspaceTool()})
	if err != nil || second.Runtime.Artifact == nil || second.Runtime.Artifact.Reused || !second.Runtime.Build.Completed ||
		second.Runtime.FinitePassed != 2 || second.Runtime.Artifact.KeySHA256 == first.Runtime.Artifact.KeySHA256 {
		t.Fatal("changed workspace used an old artifact", err)
	}
}

func TestWorkspaceExecutorQueuedCancellationAndClosedLifetime(t *testing.T) {
	owner := NewExecutor()
	<-owner.gate
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := owner.Execute(ctx, importedActivityWorkspace(), ownedWorkspaceSuite(t, "x", "x"), ExecuteOptions{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("queued workspace request ignored its context", err)
	}
	owner.gate <- struct{}{}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := owner.Execute(context.Background(), importedActivityWorkspace(), ownedWorkspaceSuite(t, "x", "x"), ExecuteOptions{}); !errors.Is(err, bodyexecution.ErrExecutorClosed) {
		t.Fatal("closed workspace executor accepted work", err)
	}
	for _, invalid := range []*Executor{nil, {}} {
		if _, err := invalid.request(context.Background(), nil); err == nil {
			t.Fatal("uninitialized workspace executor accepted work")
		}
	}
}

func TestWorkspaceExecutorCloseCancelsActiveRequestBeforeCleanup(t *testing.T) {
	owner := NewExecutor()
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		_, err := owner.request(context.Background(), func(ctx context.Context) (Result, error) {
			close(started)
			<-ctx.Done()
			return Result{}, ctx.Err()
		})
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("workspace request did not start")
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("active workspace request was not canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Close left the active request blocked")
	}
}
