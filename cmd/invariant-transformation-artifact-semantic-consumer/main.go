package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/reportconsumer"
)

func main() {
	artifactPath := flag.String("artifact", "", "generated artifact to consume")
	expectedPath := flag.String("expected-projection", "", "expected artifact semantic projection")
	headSHA := flag.String("head-sha", "", "exact subject commit")
	flag.Parse()
	if *artifactPath == "" || *expectedPath == "" || *headSHA == "" {
		fail("-artifact, -expected-projection, and -head-sha are required")
	}
	var expected reportconsumer.ArtifactSemanticProjection
	expectedRaw, err := os.ReadFile(*expectedPath)
	if err != nil {
		fail(err.Error())
	}
	if err := json.Unmarshal(expectedRaw, &expected); err != nil {
		fail(fmt.Sprintf("decode expected artifact projection: %v", err))
	}
	actual, err := reportconsumer.ProjectArtifactFile(*artifactPath, *headSHA)
	if err != nil {
		fail(err.Error())
	}
	if err := reportconsumer.CompareArtifactSemanticProjection(expected, actual); err != nil {
		fail(err.Error())
	}
	fmt.Printf("artifact semantic projection matches: case=%s execution=%s raw=%s semantic=%s\n", actual.CaseID, actual.ExecutionID, actual.RawDigest, actual.SemanticDigest)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
