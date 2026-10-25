package main

import (
	"context"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const packageGenerateUsage = "usage: gooo package generate --go-package <name> <workspace.json>"

// Emit only complete source. Source loading and package resolution use the same
// bounded workspace path as inspection, without compiling or executing a binary.
func runPackageGenerate(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) != 3 || args[0] != "--go-package" {
		return reportUsage(false, stdout, stderr, "package generate", packageGenerateUsage)
	}
	program, manifestDigest, err := inspectWorkspaceSource(reader, args[2])
	var source string
	if err == nil {
		source, err = bodyexecution.GenerateLibrary(context.Background(), "workspace.gooo",
			[]byte(program.Source), program.Entry.LoweredName, args[1])
	}
	if err != nil {
		fmt.Fprintf(stderr, "gooo package generate: %v\n", err)
		return exitFailure
	}
	header := fmt.Sprintf("// Workspace manifest: %s\n// Workspace image: %s\n", manifestDigest, program.Workspace.Digest)
	for _, pkg := range program.Workspace.Packages {
		for _, source := range pkg.Sources {
			header += fmt.Sprintf("// Gooo source package=%q file=%q digest=%q\n", pkg.Path, source.Filename, source.SourceDigest)
		}
	}
	_, err = io.WriteString(stdout, header+source)
	if err != nil {
		return exitFailure
	}
	return exitOK
}
