package main

import (
	"flag"
	"fmt"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	verify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify"
)

func run() error {
	mode := flag.String("mode", "", "generate or verify")
	repository := flag.String("repository", "", "subject repository")
	subjectSHA := flag.String("subject-sha", "", "exact subject commit")
	output := flag.String("output", "", "output JSON path")
	ledgerPath := flag.String("ledger", "", "ledger JSON path")
	flag.Parse()
	if *output == "" {
		return fmt.Errorf("-output is required")
	}
	switch *mode {
	case "generate":
		ledger, err := metric.Generate(*repository, *subjectSHA)
		if err != nil {
			return err
		}
		return artifact.WriteJSON(*output, ledger)
	case "verify":
		if *ledgerPath == "" {
			return fmt.Errorf("-ledger is required")
		}
		ledger, err := artifact.ReadJSON[metric.Ledger](*ledgerPath)
		if err != nil {
			return err
		}
		receipt, err := verify.Replay(ledger)
		if err != nil {
			return err
		}
		return artifact.WriteJSON(*output, receipt)
	default:
		return fmt.Errorf("-mode must be generate or verify")
	}
}
