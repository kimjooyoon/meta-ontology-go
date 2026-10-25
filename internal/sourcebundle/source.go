// Package sourcebundle lowers an explicit, bounded same-package source set.
// Imports belong to the package compiler; this boundary never guesses their scope.
package sourcebundle

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"strings"
)

const (
	MaxSources     = 64
	MaxSourceBytes = 1 << 20
	MaxTotalBytes  = 16 << 20
)

type Source struct {
	Filename string
	Content  string
}

// Digest binds ordered member names and exact bytes, including comments/spacing.
// It deliberately differs from a single-file content digest.
func Digest(sources []Source) ([32]byte, error) {
	if len(sources) == 0 || len(sources) > MaxSources {
		return [32]byte{}, fmt.Errorf("source bundle requires 1 to %d members", MaxSources)
	}
	seen, total := map[string]bool{}, 0
	h := sha256.New()
	io.WriteString(h, "gooo/source-bundle/v1\x00")
	for _, source := range sources {
		if err := checkName(source.Filename, seen); err != nil {
			return [32]byte{}, err
		}
		total += len(source.Content)
		if len(source.Content) > MaxSourceBytes || total > MaxTotalBytes {
			return [32]byte{}, fmt.Errorf("source bundle byte limit exceeded")
		}
		writePart(h, source.Filename)
		writePart(h, source.Content)
	}
	var digest [32]byte
	copy(digest[:], h.Sum(nil))
	return digest, nil
}

func checkName(name string, seen map[string]bool) error {
	key := strings.ToLower(name)
	if !fs.ValidPath(name) || !strings.HasSuffix(name, ".gooo") ||
		strings.ContainsAny(name, "\\\x00:") || seen[key] {
		return fmt.Errorf("source bundle member name is invalid or duplicated: %q", name)
	}
	seen[key] = true
	return nil
}

func writePart(h hash.Hash, value string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	h.Write(size[:])
	io.WriteString(h, value)
}
