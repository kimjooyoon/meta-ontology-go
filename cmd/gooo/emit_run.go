package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/artifactemit"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"
)

func runEmit(args []string, stdout, stderr io.Writer) int {
	options, err := parseEmitArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, emitUsage)
		return exitUsage
	}
	info, err := os.Stat(options.directory)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(stderr, "gooo: emit: package directory is unavailable: %s\n", options.directory)
		return exitFailure
	}
	sources, err := packageexecution.LoadDirectory(options.directory)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: emit: %v\n", err)
		return exitFailure
	}
	receipt := packageexecution.Execute(packageexecution.Request{
		PackagePath: filepath.Base(filepath.Clean(options.directory)), Entry: options.entry, Sources: sources,
	})
	receiptJSON, err := packageexecution.Marshal(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: emit receipt: %v\n", err)
		return exitFailure
	}
	artifact := artifactemit.Emit(options.kind, receiptJSON)
	payload, err := artifactemit.Marshal(artifact)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: emit artifact: %v\n", err)
		return exitFailure
	}
	if _, err := stdout.Write(payload); err != nil {
		return exitFailure
	}
	if artifact.Decision != "PASS" {
		return exitFailure
	}
	return exitOK
}
