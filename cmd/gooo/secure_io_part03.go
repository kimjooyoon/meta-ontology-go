package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

func digestMatches(path string, want []byte) (bool, error) {
	got, err := digestFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	wantDigest := sha256.Sum256(want)
	return got == wantDigest, nil
}
func digestFile(path string) ([sha256.Size]byte, error) {
	file, err := openRegularFile(path)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	defer file.Close()
	hash := sha256.New()
	read, err := io.Copy(hash, io.LimitReader(file, maxOutputBytes+1))
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("digest %q: %w", path, err)
	}
	if read > maxOutputBytes {
		return [sha256.Size]byte{}, outputLimitError(maxOutputBytes)
	}
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return digest, nil
}
func outputLimitError(limit int64) error {
	return fmt.Errorf("generated output exceeds maximum size of %d bytes", limit)
}
