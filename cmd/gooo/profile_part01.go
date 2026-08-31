package main

import (
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"
)

const profileUsage = "usage: gooo profile [--json] [--samples <count>] --entry <activity> <file.gooo>"

func runProfile(args []string, reader SourceReader, measurer languageprofile.Measurer, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseProfileArguments(args)
	if err != nil {
		return reportUsage(jsonMode, stdout, stderr, "profile", profileUsage)
	}
	source, err := readSource(reader, options.filename)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "profile", options.filename,
			"read", err.Error(), sourceexecutionSpan())
	}
	receipt := languageprofile.Observe(languageprofile.Request{
		Filename: options.filename, Source: string(source), Entry: options.entry, Samples: options.samples,
	}, measurer)
	payload, err := languageprofile.Marshal(receipt)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: profile receipt: %v\n", err)
		return exitFailure
	}
	if jsonMode {
		if _, err := stdout.Write(payload); err != nil {
			return exitFailure
		}
	} else if receipt.Decision == "PASS" {
		fmt.Fprintf(stdout, "profiled: %s samples=%d wall_ns=%d/%d/%d total_alloc_bytes=%d/%d/%d\n",
			receipt.ProfiledEntry.Activity, receipt.Summary.SamplesObserved,
			receipt.Summary.WallMinNanoseconds, receipt.Summary.WallMedianNanoseconds,
			receipt.Summary.WallMaxNanoseconds, receipt.Summary.TotalAllocMinBytes,
			receipt.Summary.TotalAllocMedianBytes, receipt.Summary.TotalAllocMaxBytes)
	} else {
		fmt.Fprintf(stderr, "%s: %s\n", options.filename, receipt.Reason)
	}
	if receipt.Decision == "PASS" {
		return exitOK
	}
	return exitFailure
}
