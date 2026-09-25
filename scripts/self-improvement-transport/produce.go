package main

import (
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func runProduce(contractFS fs.FS, contractName string, observationRaw []byte, repository,
	subjectSHA, checkoutSHA, workflowRef, workflowSHA string, runID int64, runAttempt int,
	job, artifactName, output string, check bool) error {
	receipt, err := selfimprovementtransport.Produce(
		contractFS, contractName,
		selfimprovementtransport.ProducerInput{
			Repository: repository, SubjectSHA: subjectSHA, CheckoutSHA: checkoutSHA,
			WorkflowRef: workflowRef, WorkflowSHA: workflowSHA, RunID: runID,
			RunAttempt: runAttempt, Job: job, ArtifactName: artifactName,
		},
		observationRaw,
	)
	if err != nil {
		return err
	}
	if err := writeJSON(output, receipt); err != nil {
		return err
	}
	if check && receipt.Decision != "BOUND" {
		return fmt.Errorf("producer receipt was not bound")
	}
	fmt.Printf("transport producer: %s %s\n", receipt.SubjectSHA, receipt.Subject.Digest)
	return nil
}
