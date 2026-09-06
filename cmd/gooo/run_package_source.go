package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"
)

func maybeRunSourcePackage(args []string, stdout, stderr io.Writer) (bool, int) {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseRunSourceArguments(args)
	if err != nil {
		if directory, ok := packageDirectoryArgument(args); ok {
			receipt := packageexecution.Reject(packageexecution.Request{
				PackagePath: filepath.Base(filepath.Clean(directory)),
				Entry:       packageEntryArgument(args),
			}, "PACKAGE_INVOCATION_UNSUPPORTED", []packageexecution.Diagnostic{{
				Stage:    "INPUT",
				Code:     "PACKAGE_INVOCATION_UNSUPPORTED",
				Filename: directory,
				Message:  runSourceUsage,
			}})
			return true, writeSourcePackageResult(receipt, directory, jsonMode, stdout, stderr, exitUsage)
		}
		return false, 0
	}
	info, err := os.Stat(options.filename)
	if err != nil || !info.IsDir() {
		return false, 0
	}
	if options.input != "" {
		receipt := packageexecution.Reject(packageexecution.Request{
			PackagePath: filepath.Base(filepath.Clean(options.filename)),
			Entry:       options.entry,
		}, "PACKAGE_VALUE_INPUT_UNSUPPORTED", []packageexecution.Diagnostic{{
			Stage:    "INPUT",
			Code:     "PACKAGE_VALUE_INPUT_UNSUPPORTED",
			Filename: options.filename,
			Message:  "value-plan input is supported only for a native .gooo source file",
		}})
		return true, writeSourcePackageResult(receipt, options.filename, jsonMode, stdout, stderr, exitFailure)
	}
	sources, err := packageexecution.LoadDirectory(options.filename)
	if err != nil {
		receipt := packageexecution.Reject(packageexecution.Request{
			PackagePath: filepath.Base(filepath.Clean(options.filename)),
			Entry:       options.entry,
		}, "PACKAGE_SOURCE_DIRECTORY_UNAVAILABLE", []packageexecution.Diagnostic{{
			Stage:    "INPUT",
			Code:     "PACKAGE_SOURCE_DIRECTORY_UNAVAILABLE",
			Filename: options.filename,
			Message:  err.Error(),
		}})
		return true, writeSourcePackageResult(receipt, options.filename, jsonMode, stdout, stderr, exitFailure)
	}
	receipt := packageexecution.Execute(packageexecution.Request{
		PackagePath: filepath.Base(filepath.Clean(options.filename)),
		Entry:       options.entry,
		Sources:     sources,
	})
	return true, writeSourcePackageResult(receipt, options.filename, jsonMode, stdout, stderr, exitFailure)
}

func writeSourcePackageResult(receipt packageexecution.Receipt, filename string, jsonMode bool, stdout, stderr io.Writer, failureCode int) int {
	data, err := packageexecution.Marshal(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: run package: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		if _, err := stdout.Write(data); err != nil {
			fmt.Fprintf(stderr, "gooo: run package: %v\n", err)
			return exitFailure
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
		fmt.Fprintf(stderr, "%s: %s: %s\n", filename, diagnostic.Code, diagnostic.Message)
	}
	if receipt.Decision != "PASS" {
		return failureCode
	}
	return exitOK
}

func packageDirectoryArgument(args []string) (string, bool) {
	for index, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if index > 0 && args[index-1] == "--entry" {
			continue
		}
		info, err := os.Stat(arg)
		if err == nil && info.IsDir() {
			return arg, true
		}
	}
	return "", false
}

func packageEntryArgument(args []string) string {
	for index := 0; index+1 < len(args); index++ {
		if args[index] == "--entry" {
			return args[index+1]
		}
	}
	return ""
}
