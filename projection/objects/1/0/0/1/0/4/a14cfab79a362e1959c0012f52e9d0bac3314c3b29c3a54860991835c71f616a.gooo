package main

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestRunLSPDoesNotWriteFilesystem(t *testing.T) {
	directory := t.TempDir()
	before := directoryEntries(t, directory)
	input := lspTranscript(
		lspRequest(1, "initialize", nil),
		lspRequest(2, "shutdown", nil),
		lspNotification("exit", nil),
	)
	output, code, stderr := runLSPTranscript(t, input)
	if code != exitOK || len(output) == 0 || stderr != "" {
		t.Fatalf("read-only lifecycle = code %d, stdout=%q, stderr=%q", code, output, stderr)
	}
	if after := directoryEntries(t, directory); !reflect.DeepEqual(before, after) {
		t.Fatalf("LSP launcher wrote filesystem entries: before=%v after=%v", before, after)
	}
}

func TestRunLSPContextCancellationAndEOFTerminateWithoutProtocolNoise(t *testing.T) {
	var eofOutput, eofStderr bytes.Buffer
	if code := runWithInput([]string{"lsp"}, bytes.NewReader(nil), &eofOutput, &eofStderr); code != exitOK || eofOutput.Len() != 0 || eofStderr.Len() != 0 {
		t.Fatalf("EOF lifecycle = code %d, stdout=%q, stderr=%q", code, eofOutput.Bytes(), eofStderr.Bytes())
	}

	reader, writer := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stdout, stderr bytes.Buffer
	done := make(chan int, 1)
	go func() { done <- runLSPContext(ctx, reader, &stdout, &stderr) }()
	cancel()
	select {
	case code := <-done:
		if code != exitFailure || stdout.Len() != 0 || stderr.Len() != 0 {
			t.Fatalf("canceled lifecycle = code %d, stdout=%q, stderr=%q", code, stdout.Bytes(), stderr.Bytes())
		}
	case <-time.After(time.Second):
		_ = writer.Close()
		t.Fatal("canceled LSP launcher did not terminate")
	}
	_ = writer.Close()
}
