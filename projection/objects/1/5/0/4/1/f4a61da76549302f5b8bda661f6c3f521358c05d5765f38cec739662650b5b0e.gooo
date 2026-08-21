package main

import (
	"go/ast"
	"go/format"
	"go/token"
	"sort"
	"strings"
)

func renderDecl(fset *token.FileSet, decl ast.Decl) string {
	buf := &strings.Builder{}
	_ = format.Node(buf, fset, decl)
	return buf.String()
}

func lineCount(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	lines := bytesCount(buf, '\n')
	if buf[len(buf)-1] != '\n' {
		lines++
	}
	return lines
}

func bytesCount(source []byte, target byte) int {
	n := 0
	for _, ch := range source {
		if ch == target {
			n++
		}
	}
	return n
}
func dedupe(values []string) []string {
	sort.Strings(values)
	out := make([]string, 0, len(values))
	for _, value := range values {
		if len(out) > 0 && out[len(out)-1] == value {
			continue
		}
		out = append(out, value)
	}
	return out
}
