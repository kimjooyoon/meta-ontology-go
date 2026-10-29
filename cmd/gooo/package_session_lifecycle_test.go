package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"
)

func TestPackageSessionEOFFinishesRequestAndReleasesArtifact(t *testing.T) {
	ownedRoot := t.TempDir()
	t.Setenv("TMPDIR", ownedRoot)
	t.Setenv("TMP", ownedRoot)
	t.Setenv("TEMP", ownedRoot)
	manifest, _, root := packageSessionFixture(t)
	request := packageSessionRequest{Schema: packageSessionRequestSchema, ID: "last", Operation: "execute",
		Arguments: []string{"--timing", "--cases", sessionCases(t, root, "cases.json", 3, 4), manifest}}
	var input, output, diagnostics bytes.Buffer
	if err := json.NewEncoder(&input).Encode(request); err != nil {
		t.Fatal(err)
	}
	if code := servePackageSession(context.Background(), &input, OSFileReader{}, &output, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var response packageSessionResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil || response.Receipt == nil ||
		response.Receipt.Result == nil || response.Receipt.Result.Runtime.FinitePassed != 1 {
		t.Fatal("EOF canceled the last queued request", err)
	}
	var timing packageWallDiagnostic
	if err := json.Unmarshal(diagnostics.Bytes(), &timing); err != nil || timing.Wall.Status != "OBSERVED" ||
		timing.Operation != "execute" || timing.SourceSHA256 != response.Receipt.Result.SourceSHA256 {
		t.Fatal("session timing lost its current observation scope", err)
	}
	assertSessionOwnedWorkspaces(t, ownedRoot, 0)
}

func TestPackageSessionCancellationInterruptsIdleInput(t *testing.T) {
	input, writer := io.Pipe()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan int, 1)
	go func() { done <- servePackageSession(ctx, input, OSFileReader{}, io.Discard, io.Discard) }()
	cancel()
	select {
	case code := <-done:
		if code != exitFailure {
			t.Fatal("canceled session reported a completed stream", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("session cancellation left its input reader blocked")
	}
}
