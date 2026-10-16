//go:build darwin || linux

package originpack

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestSmallOriginAllocationFresh1813(t *testing.T) {
	raw := bytes.Repeat([]byte("small source-bound hint record 1813\n"), 19)
	packed, err := Pack(raw)
	if err != nil {
		t.Fatal(err)
	}
	blocks := []int64{}
	for index, content := range [][]byte{raw, packed} {
		path := filepath.Join(t.TempDir(), []string{"original", "compressed"}[index])
		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			t.Fatal("native allocation observation unavailable")
		}
		blocks = append(blocks, stat.Blocks*512)
	}
	t.Logf("ORIGIN_ALLOCATION_METRIC raw_bytes=%d stored_bytes=%d raw_allocated_bytes=%d stored_allocated_bytes=%d",
		len(raw), len(packed), blocks[0], blocks[1])
}
