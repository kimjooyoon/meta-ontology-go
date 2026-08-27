package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	sandbox "github.com/kimjooyoon/meta-ontology-go/internal/meta/reflectivequerysandbox"
)

func main() {
	source := flag.String("source", "", "Gooo source")
	subject := flag.String("subject-sha", "", "exact subject commit")
	checkoutEvidencePath := flag.String("subject-checkout-evidence", "", "producer shell evidence binding subject SHA to checkout HEAD")
	output := flag.String("output", "", "observation receipt")
	repositoryBefore := flag.String("repository-before", "", "repository status before observation")
	repositoryAfter := flag.String("repository-after", "", "repository status after observation")
	proposalDir := flag.String("proposal-dir", "", "ephemeral generated proposal directory")
	proposalReceipt := flag.String("proposal-receipt", "", "portable proposal receipt")
	flag.Parse()
	if *source == "" || *output == "" {
		fail("usage: producer -source FILE -subject-sha SHA -subject-checkout-evidence FILE -repository-before FILE -repository-after FILE -output FILE")
	}
	checkoutEvidence := ""
	if *checkoutEvidencePath != "" {
		data, readErr := os.ReadFile(*checkoutEvidencePath)
		if readErr != nil {
			fail("read subject checkout evidence: %v", readErr)
		}
		checkoutEvidence = string(data)
	}
	observation, err := sandbox.ObserveWithCheckoutEvidence(*source, *subject, checkoutEvidence, *repositoryBefore, *repositoryAfter)
	if err != nil {
		fail("observe source: %v", err)
	}
	data, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		fail("encode observation: %v", err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0o644); err != nil {
		fail("write observation: %v", err)
	}
	if *proposalDir != "" || *proposalReceipt != "" {
		if *proposalDir == "" || *proposalReceipt == "" {
			fail("proposal-dir and proposal-receipt must be provided together")
		}
		if err := writeProposalArtifacts(*source, *proposalDir, *proposalReceipt, observation); err != nil {
			fail("write proposal artifacts: %v", err)
		}
	}
	fmt.Printf("producer observation: %s nodes=%d facts=%d attempts=%d transitions=%d\n", observation.Schema, observation.Source.NodeCount, observation.Source.FactCount, len(observation.Attempts), len(observation.Claims))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
