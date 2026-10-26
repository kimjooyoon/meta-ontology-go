package main

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
)

const packageGenerateUsage = "usage: gooo package generate --go-package <name> [--out <directory>] <workspace.json>"

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
		source, err = bodylibrary.Generate(context.Background(), "workspace.gooo",
			[]byte(program.Source), program.Entry.LoweredName, options.packageName)
	}
	if err != nil {
		fmt.Fprintf(stderr, "gooo package generate: %v\n", err)
		return exitFailure
	}
	var header strings.Builder
	info, _ := debug.ReadBuildInfo()
	compiler := versionBuildMetadata(info, runtime.Version())
	fmt.Fprintf(&header, "// Compiler version=%q source=%q source_status=%q go=%q\n",
		compiler.Version, compiler.CompilerSourceSHA, compiler.SourceStatus, compiler.GoVersion)
	fmt.Fprintf(&header, "// Workspace manifest: %s\n// Workspace image: %s\n", manifestDigest, program.Workspace.Digest)
	for _, pkg := range program.Workspace.Packages {
		for _, source := range pkg.Sources {
			fmt.Fprintf(&header, "// Gooo source package=%q file=%q source_json_string_sha256=%q\n",
				pkg.Path, source.Filename, source.SourceDigest)
		}
	}
	header.WriteString(source)
	if options.outputDir != "" {
		output, err := writePackageLibrary(options, program, []byte(header.String()))
		if err != nil {
			fmt.Fprintf(stderr, "gooo package generate: %v\n", err)
			return exitFailure
		}
		if _, err := fmt.Fprintf(stdout, "generated: %s\n", output); err != nil {
			return exitFailure
		}
		return exitOK
	}
	_, err = io.WriteString(stdout, header.String())
	if err != nil {
		return exitFailure
	}
	return exitOK
}
