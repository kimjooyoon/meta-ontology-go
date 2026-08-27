package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/reportconsumer"
)

func main() {
	reportPath := flag.String("report", "", "unbound witness report")
	beforePath := flag.String("before", "", "before snapshot metadata")
	afterPath := flag.String("after", "", "after snapshot metadata")
	sourcePath := flag.String("source", "", "Gooo source")
	headSHA := flag.String("head-sha", "", "exact subject commit")
	outputPath := flag.String("output", "", "bound report output")
	artifactProjectionPath := flag.String("artifact-projection", "", "independent generated artifact semantic projection output")
	flag.Parse()
	if *reportPath == "" || *beforePath == "" || *afterPath == "" || *sourcePath == "" || *headSHA == "" || *outputPath == "" || *artifactProjectionPath == "" {
		fail("-report, -before, -after, -source, -head-sha, -output, and -artifact-projection are required")
	}
	reportRaw, err := os.ReadFile(*reportPath)
	if err != nil {
		fail(err.Error())
	}
	beforeRaw, err := os.ReadFile(*beforePath)
	if err != nil {
		fail(err.Error())
	}
	afterRaw, err := os.ReadFile(*afterPath)
	if err != nil {
		fail(err.Error())
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	report, err := reportconsumer.Bind(reportRaw, beforeRaw, afterRaw, source, *headSHA)
	if err != nil {
		fail(err.Error())
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fail(fmt.Sprintf("encode bound report: %v", err))
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, append(raw, '\n'), 0o644); err != nil {
		fail(err.Error())
	}
	projection, err := reportconsumer.ArtifactProjection(report, source, *headSHA)
	if err != nil {
		fail(err.Error())
	}
	projectionRaw, err := json.MarshalIndent(projection, "", "  ")
	if err != nil {
		fail(fmt.Sprintf("encode artifact projection: %v", err))
	}
	if err := os.MkdirAll(filepath.Dir(*artifactProjectionPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*artifactProjectionPath, append(projectionRaw, '\n'), 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("independent report consumer: execution=%s state=%s snapshots=%d/%d\n", report.ExecutionID, report.RepositoryObservation.State, report.Summary.RepositoryNetSnapshotObservations, report.Summary.RepositoryNetSnapshotDenominator)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
