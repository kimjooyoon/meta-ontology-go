package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxInputBytes     int64 = 1 << 20
	maxOutputBytes    int64 = 16 << 20
	generatedFileName       = "semantic.gooo.go"
)

func readSource(reader SourceReader, filename string) ([]byte, error) {
	source, err := reader.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if int64(len(source)) > maxInputBytes {
		return nil, inputLimitError(maxInputBytes)
	}
	return source, nil
}

func readRegularFile(path string, limit int64) ([]byte, error) {
	file, err := openRegularFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", path, err)
	}
	if info.Size() > limit {
		return nil, inputLimitError(limit)
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	if int64(len(data)) > limit {
		return nil, inputLimitError(limit)
	}
	return data, nil
}

func inputLimitError(limit int64) error {
	return fmt.Errorf("input exceeds maximum size of %d bytes", limit)
}

func validateRegularFile(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%q is a symbolic link", path)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%q is not a regular file", path)
	}
	return nil
}

func canonicalOutputRoot(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("output root is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("absolute path: %w", err)
	}
	root := filepath.Clean(abs)
	if err := ensureOutputDirectory(root); err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("canonical path: %w", err)
	}
	return filepath.Clean(canonical), nil
}

func ensureOutputDirectory(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
		info, err = os.Lstat(path)
	}
	if err != nil {
		return fmt.Errorf("inspect directory: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("output root %q is a symbolic link", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("output root %q is not a directory", path)
	}
	return nil
}

func resolveOutputPath(root, name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || filepath.Base(name) != name {
		return "", fmt.Errorf("output name %q escapes its root", name)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("absolute root: %w", err)
	}
	target := filepath.Clean(filepath.Join(absRoot, name))
	if !pathContained(absRoot, target) {
		return "", fmt.Errorf("output name %q escapes its root", name)
	}
	if err := validateOutputTarget(target); err != nil {
		return "", err
	}
	return target, nil
}

func pathContained(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil || filepath.IsAbs(rel) || rel == "." || rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func validateOutputTarget(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect output: %w", err)
	}
	return validateRegularFile(path, info)
}

func writeGeneratedOutput(path string, data []byte) error {
	if int64(len(data)) > maxOutputBytes {
		return outputLimitError(maxOutputBytes)
	}
	same, err := digestMatches(path, data)
	if err != nil {
		return err
	}
	if same {
		return nil
	}
	if err := validateOutputTarget(path); err != nil {
		return err
	}
	return writeAtomicFile(path, data)
}

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
