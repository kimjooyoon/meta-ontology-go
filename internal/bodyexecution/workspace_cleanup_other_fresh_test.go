//go:build !windows

package bodyexecution

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestNonWindowsWorkspaceErrorsFresh1713(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"numeric-overlap", &os.PathError{Op: "remove", Path: "owned", Err: syscall.Errno(32)}},
		{"permission", os.ErrPermission},
		{"message-only", errors.New("The process cannot access the file because it is being used by another process.")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if retryableWorkspaceRemoval(tc.err) {
				t.Fatal("non-Windows error was mistaken for a Win32 sharing identity")
			}
		})
	}
}
