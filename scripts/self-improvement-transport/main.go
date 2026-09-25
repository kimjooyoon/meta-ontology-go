package main

import (
	"flag"
	"fmt"
	"os"

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
