package main

import (
	"fmt"
	"os"

	promotion "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagesourcebindingpromotion"
)

func run(args []string) error {
	options, err := parseOptions(args)
	if err != nil {
		return err
	}
	if options.check != "" {
		report, err := promotion.LoadReport(options.check)
		if err != nil {
			return err
		}
		return promotion.Validate(report)
	}
	contract, err := promotion.LoadContract(options.contract)
	if err != nil {
		return err
	}
	independence, err := promotion.LoadIndependence(options.independence)
	if err != nil {
		return err
	}
	read := func(path string) ([]byte, error) { return os.ReadFile(path) }
	paths := []string{options.policySource, options.policyArtifact, options.policyReplay, options.producer,
		options.receipt, options.oracle, options.unknownProducer, options.unknownOracle, options.mismatchedOracle}
	values := make([][]byte, len(paths))
	for index, path := range paths {
		values[index], err = read(path)
		if err != nil {
			return err
		}
	}
	report := promotion.Evaluate(promotion.Input{Contract: contract, HeadSHA: options.head,
		PolicySource: values[0], PolicyArtifact: values[1], PolicyReplayArtifact: values[2],
		Producer: values[3], Receipt: values[4], Oracle: values[5], UnknownProducer: values[6],
		UnknownOracle: values[7], MismatchedOracle: values[8], Independence: independence})
	if err := promotion.WriteReport(options.output, report); err != nil {
		return err
	}
	if err := promotion.Validate(report); err != nil {
		return err
	}
	fmt.Printf("source binding promotion: %s %d/%d\n", report.Decision, report.Summary.CasesSatisfied, report.Summary.CasesTotal)
	return nil
}
