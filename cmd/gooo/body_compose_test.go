package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBodyCompositionCLIProducesRunnableAndReplayableFiles(t *testing.T) {
	root := t.TempDir()
	source := "../../examples/body-codegen/native-composition.gooo.fixture"
	cases := "../../examples/body-codegen/native-composition-cases.json"
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	directory := filepath.Join(root, "first")
	args := []string{"body-compose", "--source", source, "--cases", cases, "--go-bin", tool, "--out", directory}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatalf("compose(%d): %s", code, stderr.String())
	}
	var first bodyCompositionOutput
	if err := json.Unmarshal(stdout.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if !first.GeneratedNow || first.Runtime.FinitePassed != 49 || !first.Runtime.RuntimeReplayed {
		t.Fatalf("composition CLI observation: %+v", first.Runtime)
	}
	for _, name := range []string{"original.gooo", "cases.json", "composition.json", "runtime.json", "realized.gooo", "generated.go", "main.go", "go.mod"} {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil || info.Size() == 0 {
			t.Fatal("missing output", name, err)
		}
	}
	stdout.Reset()
	stderr.Reset()
	args = []string{"body-compose", "--source", source, "--cases", cases, "--composition", filepath.Join(directory, "composition.json"), "--go-bin", tool}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatalf("replay(%d): %s", code, stderr.String())
	}
	var replay bodyCompositionOutput
	if err := json.Unmarshal(stdout.Bytes(), &replay); err != nil {
		t.Fatal(err)
	}
	if replay.GeneratedNow || replay.Runtime.ModelCalls != 0 || !replay.Runtime.RuntimeReplayed || replay.Runtime.FinitePassed != 49 || replay.Composition.GeneratedSHA256 != first.Composition.GeneratedSHA256 {
		t.Fatal("saved composition did not execute without generation")
	}
	stdout.Reset()
	stderr.Reset()
	args = []string{"--source", source, "--cases", cases, "--out", directory, "--model", "missing-model"}
	if code := runBodyComposeContext(context.Background(), args, &stdout, &stderr); code != exitFailure || stdout.Len() != 0 || !bytes.Contains(stderr.Bytes(), []byte("new directory")) {
		t.Fatal("existing output was not rejected before model load")
	}
}

func TestBodyCompositionUsageAndFailureRecords(t *testing.T) {
	for _, args := range [][]string{nil, {"--source"}, {"--source", "x", "--cases", "y", "--model", "z", "--composition", "q"}, {"--unknown", "x"}} {
		var stdout, stderr bytes.Buffer
		if code := runBodyComposeContext(context.Background(), args, &stdout, &stderr); code != exitUsage || stdout.Len() != 0 {
			t.Fatalf("usage(%v): %d %s", args, code, stderr.String())
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	args := []string{"--source", "../../examples/body-codegen/native-composition.gooo.fixture", "--cases", "../../examples/body-codegen/native-composition-cases.json"}
	if code := runBodyComposeContext(ctx, args, &stdout, &stderr); code != exitFailure {
		t.Fatal("cancelled composition succeeded")
	}
	var failed bodyCompositionOutput
	if err := json.Unmarshal(stdout.Bytes(), &failed); err != nil || failed.Composition.Stage != "PLAN" || failed.Composition.Failure == "" || failed.Runtime.Build.Started {
		t.Fatalf("missing cancelled generation frontier: %v %+v", err, failed)
	}
}
