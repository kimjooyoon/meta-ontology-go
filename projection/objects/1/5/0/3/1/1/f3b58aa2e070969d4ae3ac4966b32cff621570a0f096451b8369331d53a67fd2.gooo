package cache

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReceiptLockRejectsSymlinkWithoutMutatingTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires platform privileges on Windows")
	}
	cache, _, _, _, receipt := projectionHitFixture(t)
	target := filepath.Join(t.TempDir(), "outside.lock")
	original := []byte("outside lock\n")
	if err := os.WriteFile(target, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, cache.receipts+".lock"); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(receipt); !errors.Is(err, ErrUnsafeReceiptLog) {
		t.Fatalf("append through lock symlink = %v, want ErrUnsafeReceiptLog", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("lock symlink target mutated: %q", got)
	}
}
