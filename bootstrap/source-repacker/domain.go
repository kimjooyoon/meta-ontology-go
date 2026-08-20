package main

import (
	"go/ast"
	"path/filepath"
	"strings"
)

type buildDomain struct {
	Package     string
	Test        bool
	FileSuffix  string
	Constraints string
}

var knownGOOS = words("aix android darwin dragonfly freebsd illumos ios js linux netbsd openbsd plan9 solaris wasip1 windows")
var knownGOARCH = words("386 amd64 arm arm64 loong64 mips mips64 mips64le mipsle ppc64 ppc64le riscv64 s390x wasm")

func domainFor(subject string, source []byte, file *ast.File) buildDomain {
	base := strings.TrimSuffix(filepath.Base(subject), ".go")
	test := strings.HasSuffix(base, "_test")
	if test {
		base = strings.TrimSuffix(base, "_test")
	}
	parts, suffix := strings.Split(base, "_"), ""
	last := parts[len(parts)-1]
	if knownGOARCH[last] {
		suffix, parts = "_"+last, parts[:len(parts)-1]
		if len(parts) > 1 && knownGOOS[parts[len(parts)-1]] {
			suffix, parts = "_"+parts[len(parts)-1]+suffix, parts[:len(parts)-1]
		}
	} else if knownGOOS[last] {
		suffix = "_" + last
	}
	prefix := source[:fileOffset(file.Package, file)]
	constraints := make([]string, 0)
	for _, line := range strings.Split(string(prefix), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//go:build ") || strings.HasPrefix(trimmed, "// +build ") {
			constraints = append(constraints, trimmed)
		}
	}
	return buildDomain{Package: file.Name.Name, Test: test, FileSuffix: suffix, Constraints: strings.Join(constraints, "\n")}
}

func fileOffset(position interface{ IsValid() bool }, file *ast.File) int {
	if !position.IsValid() {
		return 0
	}
	return int(file.Package) - 1
}

func words(value string) map[string]bool {
	result := make(map[string]bool)
	for word := range strings.FieldsSeq(value) {
		result[word] = true
	}
	return result
}
