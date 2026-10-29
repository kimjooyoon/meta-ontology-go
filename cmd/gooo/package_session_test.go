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

type packageSessionClient struct {
	input  *io.PipeWriter
	output *io.PipeReader
	write  *json.Encoder
	read   *json.Decoder
	done   chan int
	cancel context.CancelFunc
}

func newPackageSessionClient(t *testing.T) *packageSessionClient {
	t.Helper()
	input, inputWriter := io.Pipe()
	output, outputWriter := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	c := &packageSessionClient{input: inputWriter, output: output, write: json.NewEncoder(inputWriter),
		read: json.NewDecoder(output), done: make(chan int, 1), cancel: cancel}
	go func() {
		code := servePackageSession(ctx, input, OSFileReader{}, outputWriter, io.Discard)
		_ = input.Close()
		_ = outputWriter.Close()
		c.done <- code
	}()
	t.Cleanup(func() {
		cancel()
		_ = inputWriter.Close()
		_ = output.Close()
	})
	return c
}

func (c *packageSessionClient) request(t *testing.T, id, operation string, arguments ...string) packageSessionResponse {
	t.Helper()
	if err := c.write.Encode(packageSessionRequest{Schema: packageSessionRequestSchema,
		ID: id, Operation: operation, Arguments: arguments}); err != nil {
		t.Fatal("write session request", err)
	}
	var response packageSessionResponse
	if err := c.read.Decode(&response); err != nil {
		t.Fatal("read session response", err)
	}
	if response.Schema != packageSessionResponseSchema || response.ID != id {
		t.Fatal("session lost request correlation", response.ID, id)
	}
	return response
}

func packageSessionFixture(t *testing.T) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	source := "package sessionapp\nnamespace sessionapp\nentity Integer id \"urn:gooo:type:integer\"\n" +
		"activity Main(Integer) -> Integer id \"urn:session:main\" computes \"return input + 1\"\n"
	manifest := `{"schema":"gooo/package-workspace-manifest/v1","entry":{"package_path":"session/app","activity":"Main"},"packages":[{"path":"session/app","sources":["main.gooo"]}]}`
	for name, content := range map[string]string{"main.gooo": source, "gooo.workspace.json": manifest} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(root, "gooo.workspace.json"), filepath.Join(root, "main.gooo"), root
}

func sessionCases(t *testing.T, root, name string, input, expected int) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": []any{map[string]any{
		"inputs": map[string]int{"session/app:Main": input}, "expected": map[string]int{"session/app:Main": expected},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func saveSessionReceipt(t *testing.T, root string, response packageSessionResponse) string {
	t.Helper()
	if response.ExitCode != exitOK || response.Receipt == nil || response.Receipt.Result == nil {
		t.Fatal("session did not return an execution receipt", response.ExitCode, response.Error)
	}
	raw, err := json.Marshal(response.Receipt)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, response.ID+"-receipt.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPackageSessionExecuteInspectReplayUsesFreshInputsAndResults(t *testing.T) {
	manifest, source, root := packageSessionFixture(t)
	client := newPackageSessionClient(t)
	first := client.request(t, "first", "execute", "--json", "--cases", sessionCases(t, root, "first.json", 1, 2), manifest)
	prior := saveSessionReceipt(t, root, first)
	if first.Receipt.Result.Runtime.Artifact == nil || first.Receipt.Result.Runtime.Artifact.Reused ||
		!first.Receipt.Result.Runtime.Build.Completed || first.Receipt.Result.Runtime.FinitePassed != 1 {
		t.Fatal("session did not build the first workspace")
	}
	second := client.request(t, "second", "replay", "--receipt", prior, "--cases", sessionCases(t, root, "second.json", 5, 6), manifest)
	assertSessionFreshReplay(t, first, second, 1)
	prior = saveSessionReceipt(t, root, second)
	third := client.request(t, "third", "replay", "--receipt", prior, "--cases", sessionCases(t, root, "third.json", 9, 0), manifest)
	assertSessionFreshReplay(t, second, third, 0)
	if third.Receipt.Decision != "PROGRESS" {
		t.Fatal("current wrong expectation reused a previous PASS")
	}
	assertSessionCurrentSource(t, client, manifest, source, prior, root)
	assertSessionClose(t, client)
}

func assertSessionFreshReplay(t *testing.T, prior, current packageSessionResponse, passed int) {
	t.Helper()
	if current.ExitCode != exitOK || current.Receipt == nil || current.Receipt.Result == nil {
		t.Fatal("session replay did not return a current receipt", current.Error)
	}
	r := current.Receipt.Result.Runtime
	before := prior.Receipt.Result.Runtime
	if r.Artifact == nil || !r.Artifact.Reused || r.Build.Started || r.Artifact.SourceBuildSHA256 != before.Artifact.SourceBuildSHA256 ||
		r.FinitePassed != passed || r.FiniteTotal != 1 || len(r.Runs) != 2 || r.ModelCalls != 0 ||
		r.RuntimeSuiteSHA256 == before.RuntimeSuiteSHA256 || current.Receipt.ReplayedFrom == "" || current.Receipt.Result.Replay == nil {
		t.Fatal("session reused an outcome or lost fresh-run provenance", r.Stage)
	}
}

func assertSessionCurrentSource(t *testing.T, client *packageSessionClient, manifest, source, prior, root string) {
	t.Helper()
	file, err := os.OpenFile(source, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := file.WriteString("\n")
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal(writeErr, closeErr)
	}
	response := client.request(t, "changed", "replay", "--receipt", prior, "--cases",
		sessionCases(t, root, "changed.json", 1, 2), "--go", "must-not-execute", manifest)
	if response.ExitCode != exitFailure || response.Receipt == nil || response.Receipt.Failure == nil ||
		response.Receipt.Failure.Code != "REPLAY_SOURCE_MISMATCH" || response.Receipt.Decision != "FAIL_CLOSED" {
		t.Fatal("session did not reload current workspace source", response.Error)
	}
}

func assertSessionClose(t *testing.T, client *packageSessionClient) {
	t.Helper()
	response := client.request(t, "closed", "close")
	if !response.Closed || response.ExitCode != exitOK {
		t.Fatal("session cleanup failed", response.Error)
	}
	select {
	case code := <-client.done:
		if code != exitOK {
			t.Fatal("session did not exit cleanly", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("closed session did not exit")
	}
}

func TestPackageSessionMalformedRequestPreservesFollowingClose(t *testing.T) {
	input := "{\n" + `{"schema":"gooo/workspace-session-request/v1","id":"close","operation":"close"}` + "\n"
	var output, diagnostics bytes.Buffer
	if code := runWithInput([]string{"package", "session"}, bytes.NewBufferString(input), &output, &diagnostics); code != exitOK {
		t.Fatal("public session dispatch failed", code, diagnostics.String())
	}
	decoder := json.NewDecoder(&output)
	var invalid, closed packageSessionResponse
	if decoder.Decode(&invalid) != nil || decoder.Decode(&closed) != nil || invalid.ExitCode != exitUsage ||
		invalid.Receipt != nil || !closed.Closed || closed.ID != "close" {
		t.Fatal("malformed request changed a later lifecycle command")
	}
}
