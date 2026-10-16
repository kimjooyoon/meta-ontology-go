//go:build windows

package bodyexecution

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestWindowsWorkspaceSharingErrorsFresh1713(t *testing.T) {
	for _, tc := range []struct {
		name  string
		err   error
		retry bool
	}{
		{"sharing", windowsSharingViolation, true},
		{"lock", windowsLockViolation, true},
		{"wrapped", &os.PathError{Op: "remove", Path: "owned.exe", Err: windowsSharingViolation}, true},
		{"permission", syscall.Errno(5), false},
		{"message-only", errors.New("The process cannot access the file because it is being used by another process."), false},
		{"nil", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if retryableWorkspaceRemoval(tc.err) != tc.retry {
				t.Fatal("cleanup confused a Win32 identity with another error")
			}
		})
	}
}

func TestWindowsWorkspaceLockReleaseFresh1715(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "owned.exe")
	if err := os.WriteFile(path, []byte("owned file"), 0600); err != nil {
		t.Fatal(err)
	}
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(name, syscall.GENERIC_READ, syscall.FILE_SHARE_READ,
		nil, syscall.OPEN_EXISTING, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = syscall.CloseHandle(handle)
		}
	}()
	attempts, waits := 0, 0
	remove := func(root string) error {
		attempts++
		err := os.RemoveAll(root)
		if attempts == 1 && !retryableWorkspaceRemoval(err) {
			t.Fatalf("native open handle did not produce a sharing refusal: %v", err)
		}
		return err
	}
	err = removeWorkspaceWithRetry(root, remove, retryableWorkspaceRemoval, func(delay time.Duration) {
		waits++
		if delay != 10*time.Millisecond || waits != 1 {
			t.Fatal("native sharing release exceeded its first retry")
		}
		if err := syscall.CloseHandle(handle); err != nil {
			t.Fatal(err)
		}
		closed = true
	})
	if err != nil || attempts != 2 || !closed {
		t.Fatalf("native sharing cleanup did not recover: %v, %d", err, attempts)
	}
}
