package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func main() {
	mode := flag.String("mode", "", "produce or consume")
	contract := flag.String("contract", "examples/self-improvement/transport.gooo", "Gooo transport contract")
	observation := flag.String("observation", "", "logical observation payload")
	producer := flag.String("producer", "", "producer receipt")
	metadata := flag.String("metadata", "", "GitHub artifact metadata")
	archiveDigest := flag.String("archive-digest", "", "consumer-calculated archive digest")
	repository := flag.String("repository", "", "owner/repository")
	subjectSHA := flag.String("subject-sha", "", "producer-declared logical subject")
	checkoutSHA := flag.String("checkout-sha", "", "actual producer checkout")
	workflowRef := flag.String("workflow-ref", "", "producer workflow ref")
	workflowSHA := flag.String("workflow-sha", "", "producer workflow file commit")
	runID := flag.Int64("run-id", 0, "producer workflow run id")
	runAttempt := flag.Int("run-attempt", 0, "producer workflow run attempt")
	job := flag.String("job", "", "producer job id")
	artifactName := flag.String("artifact-name", selfimprovementtransport.ArtifactName, "opaque artifact lookup label")
	output := flag.String("output", "", "receipt output")
	check := flag.Bool("check", false, "validate the generated receipt")
	checkReadOnly := flag.Bool("check-read-only", false, "allow only the single unsigned attestation unknown")
	flag.Parse()
	if err := run(*mode, *contract, *observation, *producer, *metadata, *archiveDigest,
		*repository, *subjectSHA, *checkoutSHA, *workflowRef, *workflowSHA, *runID,
		*runAttempt, *job, *artifactName, *output, *check, *checkReadOnly); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

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
		receipt, err := selfimprovementtransport.Produce(contractFS, contractName, selfimprovementtransport.ProducerInput{
			Repository: repository, SubjectSHA: subjectSHA, CheckoutSHA: checkoutSHA,
			WorkflowRef: workflowRef, WorkflowSHA: workflowSHA, RunID: runID,
			RunAttempt: runAttempt, Job: job, ArtifactName: artifactName,
		}, observationRaw)
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
	case "consume":
		producerRaw, err := os.ReadFile(producer)
		if err != nil {
			return fmt.Errorf("read producer receipt: %w", err)
		}
		metadataRaw, err := os.ReadFile(metadata)
		if err != nil {
			return fmt.Errorf("read transport metadata: %w", err)
		}
		report := selfimprovementtransport.Evaluate(contractFS, contractName, repository, runID,
			observationRaw, producerRaw, metadataRaw, archiveDigest)
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
			report.Metrics.VerifiedTotal, report.Metrics.FixedObligationTotal, report.Metrics.CoverageBasisPoints)
	default:
		return fmt.Errorf("mode must be produce or consume")
	}
	return nil
}

func contractFileSystem(path string) (fs.FS, string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, "", fmt.Errorf("resolve contract: %w", err)
	}
	return os.DirFS(filepath.Dir(absolute)), filepath.Base(absolute), nil
}

func writeJSON(path string, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode receipt: %w", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write receipt: %w", err)
	}
	return nil
}
