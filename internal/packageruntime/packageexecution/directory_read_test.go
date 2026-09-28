package packageexecution

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
)

type sourceReadCounter struct{ read int }

func (source *sourceReadCounter) Read(data []byte) (int, error) {
	clear(data)
	source.read += len(data)
	return len(data), nil
}

func TestReadBoundedSourceByteBudget(t *testing.T) {
	for _, size := range []int{0, maxSourceBytes - 1, maxSourceBytes, maxSourceBytes + 1, 4 * maxSourceBytes} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			source := &sourceReadCounter{}
			data, err := readBoundedSource(io.LimitReader(source, int64(size)))
			if err != nil {
				t.Fatal(err)
			}
			want := min(size, maxSourceBytes+1)
			if len(data) != want || source.read != want {
				t.Fatalf("buffer=%d consumed=%d, want exactly %d bytes", len(data), source.read, want)
			}
		})
	}
}

func TestReadBoundedSourcePreservesReadError(t *testing.T) {
	want := errors.New("source read interrupted")
	source := io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(want))
	_, err := readBoundedSource(source)
	if !errors.Is(err, want) {
		t.Fatalf("read error=%v, want %v", err, want)
	}
}

func TestLoadDirectorySourceReadLimit(t *testing.T) {
	for _, size := range []int{maxSourceBytes, maxSourceBytes + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "a.gooo"), []byte(strings.Repeat("x", size)), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "z.gooo"), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
			sources, err := LoadDirectory(directory)
			if size > maxSourceBytes {
				if err == nil || sources != nil {
					t.Fatalf("oversize source must return an error and no package: sources=%d err=%v", len(sources), err)
				}
				if !strings.Contains(err.Error(), "exceeds "+strconv.Itoa(maxSourceBytes)+" bytes") {
					t.Fatalf("unexpected oversize diagnostic: %v", err)
				}
				return
			}
			if err != nil || len(sources) != 2 {
				t.Fatalf("exact-limit package: sources=%d err=%v", len(sources), err)
			}
			if sources[0].Filename != "a.gooo" || len(sources[0].Content) != size {
				t.Fatal("accepted source bytes or deterministic source order changed")
			}
		})
	}
}
