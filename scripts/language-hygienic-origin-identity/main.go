package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/hygienicoriginidentity"
)

func main() {
	source := flag.String("source", "examples/language-hygienic-origin-identity/main.gooo", "Gooo experiment source")
	head := flag.String("head-sha", "", "exact source commit")
	output := flag.String("output", "", "receipt output outside the repository")
	expectUnknown := flag.Bool("expect-unknown", false, "require the explicit UNKNOWN guardrail")
	check := flag.Bool("check", false, "validate the fixed experiment contract")
	flag.Parse()
	if err := run(*source, *head, *output, *expectUnknown, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(source, head, output string, expectUnknown, check bool) error {
	if output == "" {
		return fmt.Errorf("output is required")
	}
	report, err := hygienicoriginidentity.Evaluate(os.DirFS("."), source, head)
	if err != nil {
		return err
	}
	if check {
		if err := hygienicoriginidentity.Validate(report, expectUnknown, head); err != nil {
			return err
		}
	}
	encoded, err := hygienicoriginidentity.Encode(report)
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, encoded, 0o644); err != nil {
		return fmt.Errorf("write receipt: %w", err)
	}
	fmt.Printf("hygienic origin identity: %s %d/%d classified, %d discharged, %d refuted, %d unknown\n",
		report.Decision, report.Metrics.ClassifiedClaimTotal, report.Metrics.FixedClaimDenominator,
		report.Metrics.DischargedClaimTotal, report.Metrics.RefutedClaimTotal, report.Metrics.UnknownPathTotal)
	return nil
}
