package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution/assuranceeligibility"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(arguments []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(arguments, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	paths := []string{options.parentReport, options.parentObservation, options.parentSuite,
		options.capabilityReport, options.capabilityObservation, options.capabilitySuite}
	payloads := make([][]byte, len(paths))
	for index, path := range paths {
		payloads[index], err = read(path)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	input := assuranceeligibility.NewInput(options.subjectSHA, payloads[0], payloads[1], payloads[2],
		payloads[3], payloads[4], payloads[5])
	report := assuranceeligibility.Evaluate(input)
	if !assuranceeligibility.Validate(report, input) {
		fmt.Fprintln(stderr, "eligibility report does not replay")
		return 1
	}
	suite := assuranceeligibility.RunSuite(input)
	if !assuranceeligibility.ValidateSuite(suite, input) {
		fmt.Fprintln(stderr, "eligibility suite does not replay")
		return 1
	}
	if write(options.report, report, stdout, stderr) != 0 {
		return 1
	}
	return write(options.suite, suite, stdout, stderr)
}
