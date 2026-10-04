//go:build unix

package bodypathstream

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestFileInputOpenDoesNotWaitBeforeDescriptorValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fifo")
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		f, err := openFileInput(path)
		if err == nil {
			err = f.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		// Join the original blocked open before recording the TDD failure.
		writer, err := os.OpenFile(path, os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			t.Fatal("cannot release FIFO reader", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("reader did not join after release")
		}
		t.Fatal("input open waited before descriptor validation")
	}
}

func TestFilesRejectFIFOInputsBeforeCreatingOutput(t *testing.T) {
	for _, index := range []int{1, 5, 7, 15} {
		args, out := fileArgs(t, true, `{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":2,"expected":5}]}`)
		path := filepath.Join(t.TempDir(), "fifo")
		if err := syscall.Mkfifo(path, 0600); err != nil {
			t.Fatal(err)
		}
		if index == 15 {
			args = append(args, "--options", path)
		} else {
			args[index] = path
		}
		var stdout, stderr bytes.Buffer
		if code := RunFilesCommand(context.Background(), "files", args, &stdout, &stderr); code != 1 ||
			stdout.Len() != 0 || !strings.Contains(stderr.String(), "input must be a regular file") {
			t.Fatal(index, code, stderr.String())
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Fatal("FIFO created output", index, err)
		}
	}
}

func TestFileInputRegularSymlinkKeepsExactBytes(t *testing.T) {
	root := t.TempDir()
	path, alias := filepath.Join(root, "file"), filepath.Join(root, "alias")
	want := []byte("원본 UTF-8 bytes\n")
	if err := os.WriteFile(path, want, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(path, alias); err != nil {
		t.Fatal(err)
	}
	if got, err := readFileInput(alias, 128); err != nil || !bytes.Equal(got, want) {
		t.Fatal("regular symlink changed source bytes", got, err)
	}
}
