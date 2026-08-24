package main

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/sourceauthorityupstream"
)

func run(ctx context.Context, arguments []string, fetcher sourceauthorityupstream.Fetcher) error {
	options, err := parseOptions(arguments)
	if err != nil {
		return err
	}
	suite := sourceauthorityupstream.RunSuite(ctx, options.expectedSHA, fetcher)
	if err := writeArtifacts(options.outputDir, suite); err != nil {
		return err
	}
	fmt.Printf("source-authority-upstream: cases=%d passed=%d coverage_bps=%d\n",
		suite.Summary.CasesTotal, suite.Summary.CasesPassed, suite.Summary.CoverageBPS)
	if suite.Decision != "PASS" {
		return fmt.Errorf("upstream conformance blocked: %s", suite.Reason)
	}
	return nil
}
