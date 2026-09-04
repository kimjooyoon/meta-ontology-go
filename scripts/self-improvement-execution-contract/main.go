package main

import (
	"flag"
	"log"
)

type options struct {
	mode           string
	contractPath   string
	requestPath    string
	resolutionPath string
	outputPath     string
	check          bool
}

func main() {
	if err := run(parseOptions()); err != nil {
		log.Fatal(err)
	}
}

func parseOptions() options {
	mode := flag.String("mode", "live", "live, cases, or verify")
	contractPath := flag.String("contract", "examples/self-improvement-execution-contract/contract.gooo", "caller-selected policy source")
	requestPath := flag.String("request", "", "optional v24 authorization request JSON")
	resolutionPath := flag.String("resolution", "", "resolution JSON for verify mode")
	outputPath := flag.String("output", "", "caller-owned output artifact")
	check := flag.Bool("check", false, "validate the emitted artifact")
	flag.Parse()
	return options{
		mode: *mode, contractPath: *contractPath, requestPath: *requestPath,
		resolutionPath: *resolutionPath, outputPath: *outputPath, check: *check,
	}
}
