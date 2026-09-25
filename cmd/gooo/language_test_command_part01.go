package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagetest"
)

const languageTestUsage = "usage: gooo test [--json] <file.gooo>"

func runLanguageTest(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	if len(args) != 1 {
		return reportUsage(jsonMode, stdout, stderr, "test", languageTestUsage)
	}
	filename := args[0]
	source, err := readSource(reader, filename)
	if err != nil {
		fmt.Fprintf(stderr, "gooo test: %v\n", err)
		return exitFailure
	}
	receipt := languagetest.Observe(languagetest.Request{Filename: filename, Source: string(source)})
	return writeLanguageTestReceipt(receipt, jsonMode, stdout, stderr)
}
