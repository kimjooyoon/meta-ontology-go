package main

import (
	"flag"
	"log"
)

type options struct {
	mode, contractPath, inputPath, reportPath, outputPath string
	check                                                 bool
}

func main() {
	if err := run(parseOptions()); err != nil {
		log.Fatal(err)
	}
}

func parseOptions() options {
	mode := flag.String("mode", "live", "live, cases, or verify")
	contract := flag.String("contract", "examples/self-improvement-ci-continuation/continuation.gooo", "caller-selected continuation policy")
	input := flag.String("input", "", "typed CI_CONTINUATION_REQUEST input JSON")
	report := flag.String("report", "", "continuation report JSON for verify mode")
	output := flag.String("output", "", "caller-owned output artifact")
	check := flag.Bool("check", false, "validate the emitted artifact")
	flag.Parse()
	return options{mode: *mode, contractPath: *contract, inputPath: *input, reportPath: *report, outputPath: *output, check: *check}
}
