package main

import (
	"bufio"
	"bytes"
	"go/build/constraint"
	pathpkg "path"
	"strings"
)

func validateGenericBuildHeader(header []byte) error {
	var goExpr, legacyExpr string
	scanner := bufio.NewScanner(bytes.NewReader(header))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "//go:build ") && !strings.HasPrefix(line, "// +build ") {
			continue
		}
		expr, err := constraint.Parse(line)
		if err != nil {
			return extractionError("validate-ast", "build-constraints", "BUILD_TAG_CONFLICT", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
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
		return extractionError("validate-ast", "build-constraints", "BUILD_TAG_CONFLICT", "KNOWN_CONTRADICTION", "report-contradiction", []string{})
	}
	return nil
}

var genericGOOS = map[string]bool{"aix": true, "android": true, "darwin": true, "dragonfly": true, "freebsd": true, "illumos": true, "ios": true, "js": true, "linux": true, "netbsd": true, "openbsd": true, "plan9": true, "solaris": true, "wasip1": true, "windows": true}
var genericGOARCH = map[string]bool{"386": true, "amd64": true, "arm": true, "arm64": true, "loong64": true, "mips": true, "mips64": true, "mips64le": true, "mipsle": true, "ppc64": true, "ppc64le": true, "riscv64": true, "s390x": true, "wasm": true}

func genericHelperPath(logical string) string {
	dir, base := pathpkg.Split(logical)
	stem := strings.TrimSuffix(base, ".go")
	test := ""
	if strings.HasSuffix(stem, "_test") {
		stem, test = strings.TrimSuffix(stem, "_test"), "_test"
	}
	var suffix []string
	parts := strings.Split(stem, "_")
	for len(parts) > 1 {
		last := parts[len(parts)-1]
		if !genericGOOS[last] && !genericGOARCH[last] {
			break
		}
		suffix = append([]string{last}, suffix...)
		parts = parts[:len(parts)-1]
	}
	name := strings.Join(parts, "_") + "_extracted"
	if len(suffix) > 0 {
		name += "_" + strings.Join(suffix, "_")
	}
	return dir + name + test + ".go"
}
