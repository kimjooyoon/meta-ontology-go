package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func runConsume(contractFS fs.FS, contractName string, observationRaw []byte, producer,
	metadata, archiveDigest, repository string, runID int64, output string, check,
	checkReadOnly bool) error {
	producerRaw, err := os.ReadFile(producer)
	if err != nil {
		return fmt.Errorf("read producer receipt: %w", err)
	}
	metadataRaw, err := os.ReadFile(metadata)
	if err != nil {
		return fmt.Errorf("read transport metadata: %w", err)
	}
	report := selfimprovementtransport.Evaluate(
		contractFS, contractName, repository, runID,
		observationRaw, producerRaw, metadataRaw, archiveDigest,
	)
	if err := writeJSON(output, report); err != nil {
		return err
	}
	if check || checkReadOnly {
		if err := selfimprovementtransport.ValidateReport(report); err != nil {
			return err
		}
	}
	if checkReadOnly {
		if err := selfimprovementtransport.CheckReadOnly(report); err != nil {
			return err
		}
	} else if check && report.Decision != selfimprovementtransport.DecisionPass {
		return fmt.Errorf("transport: %s / %s", report.Decision, report.Reason)
	}
	fmt.Printf("transport consumer: %s %d/%d (%d bps)\n", report.Decision,
		report.Metrics.VerifiedTotal, report.Metrics.FixedObligationTotal,
		report.Metrics.CoverageBasisPoints)
	return nil
}
