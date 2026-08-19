package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"strings"
)

func parseCheckArguments(args []string) (checkOptions, error) {
	options := checkOptions{}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--semantic":
			if options.semantic {
				return checkOptions{}, errors.New(checkUsage)
			}
			options.semantic = true
		case "--provenance-store":
			if options.provenanceStore != "" || index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
				return checkOptions{}, errors.New(checkUsage)
			}
			options.provenanceStore = args[index+1]
			index++
		default:
			if strings.HasPrefix(args[index], "-") || options.filename != "" {
				return checkOptions{}, errors.New(checkUsage)
			}
			options.filename = args[index]
		}
	}
	if options.filename == "" || (options.provenanceStore != "" && !options.semantic) {
		return checkOptions{}, errors.New(checkUsage)
	}
	return options, nil
}
func checkArguments(args []string) (semanticMode bool, filename string, ok bool) {
	options, err := parseCheckArguments(args)
	if err != nil || options.provenanceStore != "" {
		return false, "", false
	}
	return options.semantic, options.filename, true
}
func reportSemanticDiagnostic(filename string, file *syntax.File, err error, stderr io.Writer) bool {
	span := syntax.Span{Filename: filename}
	if file != nil {
		span = file.Span
	}
	output, formatErr := formatSemanticDiagnostics(span, err)
	if formatErr != nil {
		fmt.Fprintln(stderr, "gooo: diagnostic resource limit exceeded")
		return false
	}
	_, writeErr := stderr.Write(output)
	return writeErr == nil
}

type semanticDiagnostic struct {
	Code    string
	Message string
}
