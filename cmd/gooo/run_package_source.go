package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageexecution"
)

func maybeRunSourcePackage(args []string, stdout, stderr io.Writer) (bool, int) {
	options, err := parseRunSourceArguments(args)
	if err != nil {
		return false, 0
	}
	info, err := os.Stat(options.filename)
	if err != nil || !info.IsDir() {
		return false, 0
	}
	sources, err := packageexecution.LoadDirectory(options.filename)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: run package: %v\n", err)
		return true, exitFailure
	}
	receipt := packageexecution.Execute(packageexecution.Request{
		PackagePath: filepath.Base(filepath.Clean(options.filename)),
		Entry:       options.entry,
		Sources:     sources,
	})
	data, err := packageexecution.Marshal(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: run package: %v\n", err)
		return true, exitFailure
	}
	if _, err := stdout.Write(data); err != nil {
		fmt.Fprintf(stderr, "gooo: run package: %v\n", err)
		return true, exitFailure
	}
	if receipt.Decision != "PASS" {
		return true, exitFailure
	}
	return true, exitOK
}
