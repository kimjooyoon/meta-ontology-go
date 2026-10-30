package main

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"
)

type blockedSessionOutput struct {
	started chan struct{}
	closed  chan struct{}
	write   sync.Once
	close   sync.Once
}

func (w *blockedSessionOutput) Write([]byte) (int, error) {
	w.write.Do(func() { close(w.started) })
	<-w.closed
	return 0, io.ErrClosedPipe
}

func (w *blockedSessionOutput) Close() error {
	w.close.Do(func() { close(w.closed) })
	return nil
}

func TestPackageSessionCancellationInterruptsBlockedOutput(t *testing.T) {
	output := &blockedSessionOutput{started: make(chan struct{}), closed: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan int, 1)
	go func() {
		done <- servePackageSession(ctx, bytes.NewBufferString("{\n"), OSFileReader{}, output, io.Discard)
	}()
	select {
	case <-output.started:
	case <-time.After(3 * time.Second):
		t.Fatal("session response did not enter the controlled output wait")
	}
	cancel()
	select {
	case code := <-done:
		if code != exitFailure {
			t.Fatal("canceled output was reported as complete", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("session cancellation left its response writer blocked")
	}
}
