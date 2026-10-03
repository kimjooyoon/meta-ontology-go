package bodyexecution

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFileDigestBufferMatchesAllBytes(t *testing.T) {
	for _, size := range []int{0, 1, fileHashBufferBytes - 1, fileHashBufferBytes, fileHashBufferBytes + 1, 3*fileHashBufferBytes + 17} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			data := bytes.Repeat([]byte{0x37, 0xe1, 0x00}, (size+2)/3)[:size]
			path := filepath.Join(t.TempDir(), "file")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("sha256:%x", sha256.Sum256(data))
			var scratch [fileHashBufferBytes]byte
			got, err := fileDigestBuffer(path, scratch[:])
			if err != nil || got != want {
				t.Fatal(got, want, err)
			}
			var standalone *Executor
			if got, err := standalone.bindFile(path); err != nil || got != want {
				t.Fatal("standalone binding", got, err)
			}
		})
	}
}

func TestFileDigestBufferRejectsInvalidInputs(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "large")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(executableFileLimit + 1); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	var scratch [fileHashBufferBytes]byte
	for _, name := range []string{root, filepath.Join(root, "missing"), path} {
		if got, err := fileDigestBuffer(name, scratch[:]); err == nil || got != "" {
			t.Fatal("invalid file accepted", name, got, err)
		}
	}
	if got, err := fileDigestBuffer(path, nil); err == nil || got != "" {
		t.Fatal("empty buffer accepted", got, err)
	}
}

func TestExecutorOwnsAndReleasesHashScratch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("current bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	e := NewExecutor()
	if e.hashScratch != nil {
		t.Fatal("unused executor allocated scratch")
	}
	<-e.gate
	first, err := e.bindFile(path)
	if err != nil {
		t.Fatal(err)
	}
	owned := e.hashScratch
	if err := os.WriteFile(path, []byte("changed bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := e.bindFile(path)
	if err != nil || first == second || owned == nil || owned != e.hashScratch {
		t.Fatal("bytes or scratch were reused incorrectly", err)
	}
	e.gate <- struct{}{}
	if err := e.Close(); err != nil || e.hashScratch != nil {
		t.Fatal("close retained scratch", err)
	}
}

// The default copy recreates the previous implementation for allocation comparison.
func previousFileDigest(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 256<<20 {
		return "", fmt.Errorf("invalid executable file")
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	return fmt.Sprintf("sha256:%x", h.Sum(nil)), err
}

func BenchmarkFileDigest(b *testing.B) {
	path := filepath.Join(b.TempDir(), "file")
	if err := os.WriteFile(path, bytes.Repeat([]byte{0x37}, 1<<20), 0600); err != nil {
		b.Fatal(err)
	}
	e := NewExecutor()
	defer e.Close()
	if _, err := e.bindFile(path); err != nil {
		b.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		bind func(string) (string, error)
	}{{"previous-default", previousFileDigest}, {"owned-32KiB", e.bindFile}} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(1 << 20)
			for b.Loop() {
				if got, err := tc.bind(path); err != nil || got == "" {
					b.Fatal(got, err)
				}
			}
		})
	}
}
