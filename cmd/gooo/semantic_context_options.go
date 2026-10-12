package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func parseSemanticContextOptions(args []string) (semanticContextLimits, string, error) {
	limits := semanticContextLimits{Depth: 1, MaxNodes: 32, MaxBytes: 16384}
	if len(args) < 2 || len(args)%2 != 0 || args[0] == "" || strings.HasPrefix(args[0], "-") {
		return limits, "", errors.New(semanticContextUsage)
	}
	if id, err := semantic.ParseIdentity(args[1]); err != nil || id.String() != args[1] {
		return limits, "", errors.New("use a canonical stable declaration ID from gooo inspect")
	}
	expected, seen := "", map[string]bool{}
	for i := 2; i < len(args); i += 2 {
		option, value := args[i], args[i+1]
		if seen[option] || value == "" {
			return limits, "", fmt.Errorf("duplicate or empty option %q", option)
		}
		seen[option] = true
		if option == "--expect-source" {
			expected = value
			continue
		}
		if err := setSemanticContextLimit(&limits, option, value); err != nil {
			return limits, "", err
		}
	}
	return limits, expected, nil
}

func setSemanticContextLimit(limits *semanticContextLimits, option, raw string) error {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fmt.Errorf("%s requires an integer", option)
	}
	switch {
	case option == "--depth" && value >= 0 && value <= 8:
		limits.Depth = value
	case option == "--max-nodes" && value >= 1 && value <= 256:
		limits.MaxNodes = value
	case option == "--max-bytes" && value >= 1024 && value <= maxGraphDumpBytes:
		limits.MaxBytes = value
	default:
		return errors.New(semanticContextUsage)
	}
	return nil
}

func readSemanticContext(filename, expected string, reader SourceReader, parser SourceParser,
	deadline time.Time, stderr io.Writer) ([]byte, semantic.IR, error) {
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		return nil, semantic.IR{}, err
	}
	if expected != "" && semantic.StableHash(source) != expected {
		return nil, semantic.IR{}, errors.New("source digest changed; inspect the current source before continuing")
	}
	file, diagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		return nil, semantic.IR{}, err
	}
	if !reportDiagnostics(diagnostics, stderr) || diagnostics.HasErrors() {
		return nil, semantic.IR{}, errors.New("source has diagnostics")
	}
	ir, err := lowerInspectIR(file, remainingDeadline(deadline))
	return source, ir, err
}
