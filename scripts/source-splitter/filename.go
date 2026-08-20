package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

var knownGOOS = wordSet("aix android darwin dragonfly freebsd illumos ios js linux netbsd openbsd plan9 solaris wasip1 windows")
var knownGOARCH = wordSet("386 amd64 arm arm64 loong64 mips mips64 mips64le mipsle ppc64 ppc64le riscv64 s390x wasm")

func secureSourcePath(root, subject string) (string, error) {
	if subject == "" || filepath.IsAbs(subject) {
		return "", fmt.Errorf("subject must be a relative file path")
	}
	clean := filepath.Clean(filepath.FromSlash(subject))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("subject escapes repository: %q", subject)
	}
	rootPath, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	target, err := filepath.EvalSymlinks(filepath.Join(rootPath, clean))
	if err != nil {
		return "", fmt.Errorf("resolve subject: %w", err)
	}
	relative, err := filepath.Rel(rootPath, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("subject escapes repository: %q", subject)
	}
	return target, nil
}

func splitPartPath(subject string, index int) (string, error) {
	if index == 1 {
		return subject, nil
	}
	directory, base := filepath.Split(filepath.FromSlash(subject))
	if !strings.HasSuffix(base, ".go") {
		return "", fmt.Errorf("unsupported Go filename %q", subject)
	}
	stem := strings.TrimSuffix(base, ".go")
	suffix := ""
	if before, ok := strings.CutSuffix(stem, "_test"); ok {
		stem, suffix = before, "_test"
	}
	parts := strings.Split(stem, "_")
	last := parts[len(parts)-1]
	if knownGOARCH[last] {
		suffix, parts = "_"+last+suffix, parts[:len(parts)-1]
		if len(parts) > 1 && knownGOOS[parts[len(parts)-1]] {
			last = parts[len(parts)-1]
			suffix, parts = "_"+last+suffix, parts[:len(parts)-1]
		}
	} else if knownGOOS[last] {
		suffix, parts = "_"+last+suffix, parts[:len(parts)-1]
	}
	stem = strings.Join(parts, "_")
	name := fmt.Sprintf("%s_split%02d%s.go", stem, index, suffix)
	return filepath.ToSlash(filepath.Join(directory, name)), nil
}

func wordSet(words string) map[string]bool {
	result := make(map[string]bool)
	for word := range strings.FieldsSeq(words) {
		result[word] = true
	}
	return result
}
