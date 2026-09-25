package cache

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestReceiptLogRejectsSymlinkWithoutMutatingTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires platform privileges on Windows")
	}
	cache, _, _, _, receipt := projectionHitFixture(t)
	target := filepath.Join(t.TempDir(), "outside.jsonl")
	original := []byte("outside\n")
	if err := os.WriteFile(target, original, 0o600); err != nil {
		t.Fatal(err)
	}
	receiptLog := filepath.Join(cache.Root(), receiptsFileName)
	if err := os.Symlink(target, receiptLog); err != nil {
		t.Fatal(err)
	}
	if _, err := cache.AppendReceipt(receipt); !errors.Is(err, ErrUnsafeReceiptLog) {
		t.Fatalf("append through receipt symlink = %v, want ErrUnsafeReceiptLog", err)
	}
	if _, err := cache.Receipts(); !errors.Is(err, ErrUnsafeReceiptLog) {
		t.Fatalf("read through receipt symlink = %v, want ErrUnsafeReceiptLog", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("symlink target mutated: %q", got)
	}
}
func TestReceiptAppendSerializesAcrossCacheHandles(t *testing.T) {
	cache, _, _, _, receipt := projectionHitFixture(t)
	second, err := Open(cache.Root())
	if err != nil {
		t.Fatal(err)
	}
	release, err := acquireReceiptFileLock(cache.receipts)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		_, err := second.AppendReceipt(receipt)
		done <- err
	}()
	<-started
	select {
	case err := <-done:
		release()
		t.Fatalf("append completed while receipt lock was held: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	release()
	if err := <-done; err != nil {
		t.Fatalf("serialized append = %v", err)
	}
	if _, err := cache.AppendReceipt(receipt); !errors.Is(err, ErrReceiptReplay) {
		t.Fatalf("cross-handle replay = %v, want ErrReceiptReplay", err)
	}
	if records, err := second.Receipts(); err != nil || len(records) != 1 {
		t.Fatalf("receipt records = %d, %v; want one", len(records), err)
	}
}
