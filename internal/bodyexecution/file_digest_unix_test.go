//go:build unix

package bodyexecution

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestFileDigestRejectsFIFOWithoutWaitingForWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fifo")
	if err := syscall.Mkfifo(path, 0700); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := fileDigest(path)
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, errNonRegularExecutable) {
			t.Fatal("FIFO rejection lost its type", err)
		}
	case <-time.After(time.Second):
		// Release the original blocked read before reporting the TDD failure.
		writer, err := os.OpenFile(path, os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			t.Fatal("cannot release blocked file reader", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("reader did not join after FIFO release")
		}
		t.Fatal("executable hash waited for FIFO writer")
	}
}

func TestExecutorRejectsFIFOAndClosesWithoutNativeWork(t *testing.T) {
	source, doc, prior, parent := fixture(t)
	path := filepath.Join(t.TempDir(), "go-fifo")
	if err := syscall.Mkfifo(path, 0700); err != nil {
		t.Fatal(err)
	}
	e := NewExecutor()
	defer e.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := e.Execute(ctx, "fixture.gooo", source, doc, prior, parent, doc.TestCases, path)
	if err == nil || !strings.Contains(err.Error(), "regular executable file") || ctx.Err() != nil {
		t.Fatal("FIFO rejection did not precede deadline", err, ctx.Err())
	}
	if len(r.Observation.Runs) != 0 || r.Observation.Build.Started || r.Observation.Toolchain.Started {
		t.Fatal("nonregular tool started native work", r.Observation)
	}
	if err := e.Close(); err != nil || e.hashScratch != nil {
		t.Fatal("failed hash retained scratch", err)
	}
	verifyReceipt(t, r)
}

func TestFileDigestFollowsRegularSymlink(t *testing.T) {
	root := t.TempDir()
	path, alias := filepath.Join(root, "file"), filepath.Join(root, "alias")
	if err := os.WriteFile(path, []byte("current regular bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(path, alias); err != nil {
		t.Fatal(err)
	}
	want, err := fileDigest(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := fileDigest(alias); err != nil || got != want {
		t.Fatal("regular symlink binding changed", got, want, err)
	}
}
