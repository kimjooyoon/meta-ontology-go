package main

import (
	"fmt"
	"os"
)

func run(mode, contract, observation, producer, metadata, archiveDigest, repository,
	subjectSHA, checkoutSHA, workflowRef, workflowSHA string, runID int64, runAttempt int,
	job, artifactName, output string, check, checkReadOnly bool) error {
	contractFS, contractName, err := contractFileSystem(contract)
	if err != nil {
		return err
	}
	observationRaw, err := os.ReadFile(observation)
	if err != nil {
		return fmt.Errorf("read observation: %w", err)
	}
	switch mode {
	case "produce":
		return runProduce(contractFS, contractName, observationRaw, repository, subjectSHA,
			checkoutSHA, workflowRef, workflowSHA, runID, runAttempt, job, artifactName, output, check)
	case "consume":
		return runConsume(contractFS, contractName, observationRaw, producer, metadata,
			archiveDigest, repository, runID, output, check, checkReadOnly)
	default:
		return fmt.Errorf("mode must be produce or consume")
	}
}
