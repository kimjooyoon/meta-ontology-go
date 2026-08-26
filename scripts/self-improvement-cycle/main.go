package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	var opts options
	flag.StringVar(&opts.headSHA, "head-sha", "", "exact checked-out commit SHA")
	flag.StringVar(&opts.branch, "branch", "", "CI head branch")
	flag.StringVar(&opts.conclusion, "ci-conclusion", "", "source CI conclusion")
	flag.Int64Var(&opts.runID, "ci-run-id", 0, "source CI workflow run ID")
	flag.StringVar(&opts.metrics, "metrics", "", "source metrics JSON")
	flag.StringVar(&opts.plan, "plan", "", "generation plan JSON")
	flag.StringVar(&opts.execution, "execution", "", "execution manifest JSON")
	flag.StringVar(&opts.receipts, "receipts", "", "receipt report JSON")
	flag.StringVar(&opts.provenance, "provenance", "", "artifact provenance JSON")
	flag.StringVar(&opts.contract, "contract", "", "Gooo contract report JSON")
	flag.StringVar(&opts.output, "output", "", "JSON output path; stdout when empty")
	flag.BoolVar(&opts.check, "check", false, "exit non-zero for an open envelope")
	flag.Parse()

	in, err := loadInputs(opts)
	if err != nil {
		exitError(err)
	}
	envelope := buildEnvelope(in, opts)
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		exitError(err)
	}
	data = append(data, '\n')
	if opts.output == "" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(opts.output, data, 0o644)
	}
	if err != nil {
		exitError(err)
	}
	if opts.check && envelope.Status != "BOUND" {
		os.Exit(1)
	}
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}

func validContractIndicators(contract contractDocument) bool {
	counts := map[string]int{}
	for _, indicator := range contract.Indicators {
		if indicator.Verdict != "PASS" {
			return false
		}
		counts[indicator.Route]++
	}
	return len(contract.Indicators) == 8 &&
		counts["FOUNDATION"] == 3 && counts["COHERENCE"] == 4 &&
		counts["REGRESSION"] == 1
}
