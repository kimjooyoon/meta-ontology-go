package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementtransport"
)

func main() {
	runPath := flag.String("run", "", "GitHub workflow run API response")
	artifactsPath := flag.String("artifacts", "", "GitHub artifact list API response")
	repository := flag.String("repository", "", "owner/repository")
	runID := flag.Int64("expected-run-id", 0, "expected producer workflow run id")
	runAttempt := flag.Int("expected-run-attempt", 0, "expected producer run attempt")
	artifactName := flag.String("artifact-name", selfimprovementtransport.ArtifactName, "artifact lookup label")
	output := flag.String("output", "", "selected transport metadata output")
	flag.Parse()
	if err := run(*runPath, *artifactsPath, *repository, *runID, *runAttempt,
		*artifactName, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
