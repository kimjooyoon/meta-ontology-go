package bodypathstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCommandArgumentsAndSetup(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		code int
		want string
	}{
		{"help", []string{"--help"}, 0, "omit for deterministic"},
		{"unknown", []string{"--unknown"}, 2, "flag provided but not defined"},
		{"missing", []string{"--model"}, 2, "flag needs an argument"},
		{"positional", []string{"requests.jsonl"}, 2, "usage:"},
		{"zero", []string{"--workers", "0"}, 2, "usage:"},
		{"unbounded", []string{"--workers", "9"}, 2, "usage:"},
		{"tool-without-execution", []string{"--go-bin", "go"}, 2, "usage:"},
		{"missing-model", []string{"--model", "missing-model.json"}, 1, "missing-model.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, diagnostics bytes.Buffer
			code := RunCommand(context.Background(), "worker", tc.args,
				io.NopCloser(strings.NewReader("")), nopWriteCloser{&out}, &diagnostics)
			if code != tc.code || !strings.Contains(diagnostics.String(), tc.want) || out.Len() != 0 {
				t.Fatalf("code=%d diagnostics=%s stdout=%s", code, &diagnostics, &out)
			}
		})
	}
}

func TestCommandRequiresCancellableTransport(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := RunCommand(context.Background(), "worker", nil, strings.NewReader(""), &out, &diagnostics); code != 1 || !strings.Contains(diagnostics.String(), "support Close") {
		t.Fatalf("code=%d diagnostics=%s", code, &diagnostics)
	}
}

func TestCommandRespondsBeforeEOFAndRecoversAfterRejectedRecord(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	input, upstream := io.Pipe()
	output, downstream := io.Pipe()
	defer upstream.Close()
	defer downstream.Close()
	var diagnostics bytes.Buffer
	valid := append(requestLine(t, "next", string(source)), '\n')
	done := make(chan int, 1)
	go func() { done <- RunCommand(ctx, "worker", nil, input, downstream, &diagnostics) }()
	go func() {
		_, _ = upstream.Write([]byte("{}\n"))
		_, _ = upstream.Write(valid)
	}()
	decoder := json.NewDecoder(output)
	for _, want := range []string{"rejected", "completed"} {
		var result Result
		if err := decoder.Decode(&result); err != nil || result.Status != want {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		if want == "completed" && (result.CorrelationID != "next" || result.Response == nil) {
			t.Fatalf("missing correlated construction: %+v", result)
		}
	}
	_ = upstream.Close()
	select {
	case code := <-done:
		if code != 0 || !json.Valid(bytes.TrimSpace(diagnostics.Bytes())) {
			t.Fatalf("code=%d setup=%s", code, &diagnostics)
		}
	case <-ctx.Done():
		t.Fatal("worker did not finish after EOF")
	}
}

func TestCommandCancellationUnblocksInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input, upstream := io.Pipe()
	defer upstream.Close()
	setup, setupWriter := io.Pipe()
	defer setup.Close()
	defer setupWriter.Close()
	var out bytes.Buffer
	done := make(chan int, 1)
	go func() { done <- RunCommand(ctx, "worker", nil, input, nopWriteCloser{&out}, setupWriter) }()
	var record map[string]any
	if err := json.NewDecoder(setup).Decode(&record); err != nil {
		t.Fatal(err)
	}
	// Keep consuming diagnostics so error reporting cannot block cancellation.
	go func() { _, _ = io.Copy(io.Discard, setup) }()
	cancel()
	select {
	case code := <-done:
		if code != 1 {
			t.Fatalf("cancellation exit=%d", code)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancellation did not interrupt stdin")
	}
}
