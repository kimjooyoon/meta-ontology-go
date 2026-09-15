package lsp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestServeContextClosesBlockedInputRead(t *testing.T) {
	pipeReader, pipeWriter := io.Pipe()
	input := &trackedPipeReader{reader: pipeReader, started: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- NewServer().ServeContext(ctx, input, &bytes.Buffer{}) }()
	<-input.started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ServeContext() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ServeContext() did not unblock its input read")
	}
	if _, err := pipeWriter.Write([]byte("unexpected input")); err == nil {
		t.Fatal("pipe writer remained open after canceled input read")
	}
	_ = pipeWriter.Close()
}
func TestServeContextReturnsWhenCloseDoesNotInterruptReader(t *testing.T) {
	input := newNonInterruptingCloser()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- NewServer().ServeContext(ctx, input, &bytes.Buffer{}) }()
	<-input.started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ServeContext() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ServeContext() waited for a reader that Close could not interrupt")
	}
	select {
	case <-input.closed:
	default:
		t.Fatal("ServeContext() did not close the input reader")
	}
	close(input.release)
}
