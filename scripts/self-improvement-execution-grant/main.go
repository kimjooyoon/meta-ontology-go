package main

import (
	"flag"
	"log"
)

type options struct {
	mode, contractPath, v24RequestPath, v24ResolutionPath string
	v25ContractPath, sourcePath, outputPath               string
	grantRequestPath, resolutionPath                      string
	check                                                 bool
}

func main() {
	if err := run(parseOptions()); err != nil {
		log.Fatal(err)
	}
}

func parseOptions() options {
	mode := flag.String("mode", "live", "live, cases, or verify")
	contract := flag.String("contract", "examples/self-improvement-execution-grant/grant.gooo", "caller-selected grant policy")
	v24Request := flag.String("v24-request", "", "optional v24 authorization request JSON")
	v24Resolution := flag.String("v24-resolution", "", "optional v24 authorization resolution JSON")
	v25Contract := flag.String("v25-contract", "", "optional v25 pre-execution contract JSON")
	grantRequest := flag.String("grant-request", "", "grant request JSON for verify mode")
	resolution := flag.String("resolution", "", "grant resolution JSON for verify mode")
	source := flag.String("source", "", "optional source artifact metadata JSON")
	output := flag.String("output", "", "caller-owned output artifact")
	check := flag.Bool("check", false, "validate the emitted artifact")
	flag.Parse()
	return options{mode: *mode, contractPath: *contract, v24RequestPath: *v24Request, v24ResolutionPath: *v24Resolution, v25ContractPath: *v25Contract, grantRequestPath: *grantRequest, resolutionPath: *resolution, sourcePath: *source, outputPath: *output, check: *check}
}
