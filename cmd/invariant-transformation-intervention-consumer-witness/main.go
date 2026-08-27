package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/interventionconsumer"
)

type dependencyReport struct {
	ProducerDependencyImports        int `json:"producer_dependency_imports"`
	AllowedProducerDependencyImports int `json:"allowed_producer_dependency_imports"`
	ArtifactObservation              int `json:"artifact_observation"`
	ExpectedArtifactObservation      int `json:"expected_artifact_observation"`
}

func main() {
	sourcePath := flag.String("source", "", "original Gooo source value contract")
	headSHA := flag.String("head-sha", "", "exact subject commit")
	reportPath := flag.String("report", "", "producer intervention report")
	dependencyPath := flag.String("dependency-report", "", "production dependency boundary report")
	outputPath := flag.String("output", "", "consumer audit output")
	flag.Parse()
	if *sourcePath == "" || *headSHA == "" || *reportPath == "" || *dependencyPath == "" || *outputPath == "" {
		fail("-source, -head-sha, -report, -dependency-report, and -output are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	report, err := os.ReadFile(*reportPath)
	if err != nil {
		fail(err.Error())
	}
	dependencyRaw, err := os.ReadFile(*dependencyPath)
	if err != nil {
		fail(err.Error())
	}
	var dependency dependencyReport
	if err := json.Unmarshal(dependencyRaw, &dependency); err != nil {
		fail(fmt.Sprintf("decode dependency report: %v", err))
	}
	audit, err := interventionconsumer.VerifyReport(report, source, *headSHA, interventionconsumer.DependencyBoundary{
		ProducerDependencyImports:        dependency.ProducerDependencyImports,
		AllowedProducerDependencyImports: dependency.AllowedProducerDependencyImports,
		ArtifactObservation:              dependency.ArtifactObservation,
		ExpectedArtifactObservation:      dependency.ExpectedArtifactObservation,
	})
	if err != nil {
		fail(err.Error())
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	raw, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, append(raw, '\n'), 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("independent intervention consumer: %s reconstructed=%d/%d actual-replay=%d/%d artifact=%d/%d coherent-tamper-rejected=%d/%d writes=%d\n",
		audit.Decision, audit.ReconstructedCases, audit.ExpectedCases, audit.ActualReplay, audit.ExpectedActualReplay,
		audit.ArtifactObservation, audit.ExpectedArtifactObservation, audit.CoherentTamperRejected, audit.ExpectedCoherentTamperRejections, audit.RepositoryWrites)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
