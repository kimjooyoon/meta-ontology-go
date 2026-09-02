package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"
)

func maybeRunSourcePackage(args []string, stdout, stderr io.Writer) (bool, int) {
	args, jsonMode := parseJSONFlag(args)
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
	if jsonMode {
		if _, err := stdout.Write(data); err != nil {
			fmt.Fprintf(stderr, "gooo: run package: %v\n", err)
			return true, exitFailure
		}
	} else if receipt.Decision == "PASS" && receipt.Execution != nil {
		fmt.Fprintf(stdout, "executed package: %s.%s(%s) -> %s sources=%d digest=%s\n",
			receipt.Execution.Entry.Package,
			receipt.Execution.Entry.Activity,
			inputNames(receipt.Execution.Entry.Inputs),
			receipt.Execution.Entry.Output.Name,
			len(receipt.Sources), receipt.Digest)
	} else if len(receipt.Diagnostics) > 0 {
		diagnostic := receipt.Diagnostics[0]
		fmt.Fprintf(stderr, "%s: %s: %s\n", options.filename, diagnostic.Code, diagnostic.Message)
	}
	if receipt.Decision != "PASS" {
		return true, exitFailure
	}
	return true, exitOK
}
