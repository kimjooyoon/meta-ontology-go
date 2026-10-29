package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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

func TestPackageSessionInputOnlyReplayRetainsUnscoredCurrentValues(t *testing.T) {
	manifest, _, root := packageSessionFixture(t)
	inputPath := filepath.Join(root, "inputs.json")
	writeInputs := func(value int) {
		raw, err := json.Marshal(map[string]any{"schema": "gooo/body-composition-inputs/v1",
			"inputs": []any{map[string]int{"session/app:Main": value}}})
		if err != nil {
			t.Fatal("write current session inputs", err)
		}
		if err := os.WriteFile(inputPath, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	client := newPackageSessionClient(t)
	writeInputs(7)
	first := client.request(t, "input-first", "execute", "--inputs", inputPath, manifest)
	prior := saveSessionReceipt(t, root, first)
	writeInputs(10)
	second := client.request(t, "input-second", "replay", "--receipt", prior, "--inputs", inputPath, manifest)
	if second.ExitCode != exitOK || second.Receipt == nil || second.Receipt.Result == nil {
		t.Fatal("input-only session replay failed")
	}
	r := second.Receipt.Result.Runtime
	if first.Receipt.Decision != "OBSERVED" || second.Receipt.Decision != "OBSERVED" ||
		second.Receipt.InputsDigest == first.Receipt.InputsDigest || second.Receipt.CasesDigest != "" ||
		r.FiniteTotal != 0 || r.FinitePassed != 0 || r.Artifact == nil || !r.Artifact.Reused ||
		len(r.Traces) != 1 || len(r.Traces[0].Deliveries) != 1 || string(r.Traces[0].Deliveries[0].Actual) != "11" {
		t.Fatal("session reused old values or invented an input-only score")
	}
	assertSessionReplayParent(t, second, prior)
	assertSessionClose(t, client)
}
