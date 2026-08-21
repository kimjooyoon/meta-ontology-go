package main

import (
	"flag"
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricintervention"
	verify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricinterventionverify"
)

func run() error {
	mode := flag.String("mode", "", "generate or verify")
	metricsPath := flag.String("metrics", "", "source metrics JSON path")
	repository := flag.String("repository", "", "subject repository")
	subjectSHA := flag.String("subject-sha", "", "exact subject commit")
	ledgerPath := flag.String("ledger", "", "metric intervention ledger path")
	output := flag.String("output", "", "output JSON path")
	flag.Parse()
	if *metricsPath == "" || *output == "" {
		return fmt.Errorf("-metrics and -output are required")
	}
	switch *mode {
	case "generate":
		ledger, err := metric.Generate(*metricsPath, *repository, *subjectSHA)
		if err != nil {
			return err
		}
		return artifact.WriteJSON(*output, ledger)
	case "verify":
		if *ledgerPath == "" {
			return fmt.Errorf("-ledger is required in verify mode")
		}
		ledger, err := artifact.ReadJSON[metric.Ledger](*ledgerPath)
		if err != nil {
			return err
		}
		receipt, err := verify.Replay(*metricsPath, ledger)
		if err != nil {
			return err
		}
		return artifact.WriteJSON(*output, receipt)
	default:
		return fmt.Errorf("-mode must be generate or verify")
	}
}
