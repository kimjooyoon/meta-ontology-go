package extractor

import (
	"bufio"
	"bytes"
	"fmt"
	"go/build/constraint"
	pathpkg "path"
	"path/filepath"
	"strconv"
	"strings"
)

func safePath(root, logical string) (string, error) {
	if logical == "" || filepath.IsAbs(logical) || strings.ContainsRune(logical, 0) {
		return "", fmt.Errorf("unsafe extraction path %q", logical)
	}
	clean := filepath.Clean(filepath.FromSlash(logical))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe extraction path %q", logical)
	}
	return filepath.Join(root, clean), nil
}

func validateBuildHeader(header []byte) error {
	var goExpr, legacyExpr string
	scanner := bufio.NewScanner(bytes.NewReader(header))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "//go:build ") && !strings.HasPrefix(line, "// +build ") {
			continue
		}
		expr, err := constraint.Parse(line)
		if err != nil {
			return fail("validate-ast", "build-constraints", "BUILD_TAG_CONFLICT", "KNOWN_CONTRADICTION", "report-contradiction", nil)
		}
		if strings.HasPrefix(line, "//go:build ") {
			goExpr = expr.String()
		} else if legacyExpr == "" {
			legacyExpr = expr.String()
		} else {
			legacyExpr += " && " + expr.String()
		}
	}
	if goExpr != "" && legacyExpr != "" && goExpr != legacyExpr {
		return fail("validate-ast", "build-constraints", "BUILD_TAG_CONFLICT", "KNOWN_CONTRADICTION", "report-contradiction", nil)
	}
	return nil
}

var goos = map[string]bool{"aix": true, "android": true, "darwin": true, "dragonfly": true, "freebsd": true, "illumos": true, "ios": true, "js": true, "linux": true, "netbsd": true, "openbsd": true, "plan9": true, "solaris": true, "wasip1": true, "windows": true}
var goarch = map[string]bool{"386": true, "amd64": true, "arm": true, "arm64": true, "loong64": true, "mips": true, "mips64": true, "mips64le": true, "mipsle": true, "ppc64": true, "ppc64le": true, "riscv64": true, "s390x": true, "wasm": true}

func helperPath(logical string, index, total int) string {
	dir, base := pathpkg.Split(logical)
	stem := strings.TrimSuffix(base, ".go")
	test := ""
	if before, ok := strings.CutSuffix(stem, "_test"); ok {
		stem, test = before, "_test"
	}
	parts := strings.Split(stem, "_")
	var suffix []string
	for len(parts) > 1 {
		last := parts[len(parts)-1]
		if !goos[last] && !goarch[last] {
			break
		}
		suffix = append([]string{last}, suffix...)
		parts = parts[:len(parts)-1]
	}
	name := strings.Join(parts, "_") + "_extracted"
	if len(suffix) > 0 {
		name += "_" + strings.Join(suffix, "_")
	}
	if total > 1 {
		name += "_" + strconv.Itoa(index+1)
	}
	return dir + name + test + ".go"
}
