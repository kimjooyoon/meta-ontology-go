package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"io"
	"strings"
)

const (
	fixPlanSchemaVersion   = "gooo-fix-plan/v1"
	fixPlanReady           = "ready"
	fixPlanSyntaxInvalid   = "syntax-invalid"
	fixPlanSemanticInvalid = "semantic-invalid"
)

func runAnalyze(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	clean, _ := parseJSONFlag(args)
	if analyzeDeltaArguments(clean) {
		return runAnalyzeDelta(clean, reader, parser, stdout, stderr)
	}
	return runAnalyzeWithLowerer(args, reader, parser, stdout, stderr, bidir.Lower)
}

const analyzeDeltaUsage = "usage: gooo analyze <authority.gooo> <generated.go> [<generated.go>...] or gooo analyze <authority.gooo> --go <generated.go> [--go <generated.go>...]"

func analyzeDeltaArguments(args []string) bool {
	if len(args) > 1 {
		return true
	}
	for _, arg := range args {
		if strings.HasPrefix(arg, "--go") || arg == "--generated-go" || arg == "--input" {
			return true
		}
	}
	return false
}
