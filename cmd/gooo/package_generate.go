package main

import (
	"fmt"
	"io"
)

const packageGenerateUsage = "usage: gooo package generate --go-package <name> [--export-stage <Name=activity-id>] [--embed-origin] [--out <directory>] <workspace.json>"

// Emit only complete source. Source loading and package resolution use the same
// bounded workspace path as inspection, without compiling or executing a binary.
func runPackageGenerate(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	options, ok := parsePackageGenerateArgs(args)
	if !ok {
		return reportUsage(false, stdout, stderr, "package generate", packageGenerateUsage)
	}
	program, manifestDigest, err := inspectWorkspaceSource(reader, options.manifest)
	var source string
	if err == nil {
		source, err = generatePackageLibrary(program, manifestDigest, options)
	}
	if err != nil {
		fmt.Fprintf(stderr, "gooo package generate: %v\n", err)
		return exitFailure
	}
	if options.outputDir != "" {
		output, err := writePackageLibrary(options, program, []byte(source))
		if err != nil {
			fmt.Fprintf(stderr, "gooo package generate: %v\n", err)
			return exitFailure
		}
		if _, err := fmt.Fprintf(stdout, "generated: %s\n", output); err != nil {
			return exitFailure
		}
		return exitOK
	}
	_, err = io.WriteString(stdout, source)
	if err != nil {
		return exitFailure
	}
	return exitOK
}
