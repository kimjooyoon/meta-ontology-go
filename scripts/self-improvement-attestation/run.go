package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementattestation"
)

func run(options options) error {
	request, err := selfimprovementattestation.LoadRequest(
		options.receiptPath, options.archivePath, options.verification,
		options.verifierExit, options.verifierVersion,
	)
	if err != nil {
		return err
	}
	receipt, resolveErr := selfimprovementattestation.Resolve(request)
	if err := selfimprovementattestation.WriteReceipt(options.output, receipt); err != nil {
		return err
	}
	if options.check {
		if err := selfimprovementattestation.ValidateReceipt(receipt); err != nil {
			return err
		}
	}
	if resolveErr != nil {
		return resolveErr
	}
	fmt.Printf("producer attestation: %s %d/%d (%d bps)\n", receipt.Decision,
		receipt.Metrics.VerifiedTotal, receipt.Metrics.FixedObligationTotal,
		receipt.Metrics.CoverageBasisPoints)
	return nil
}
