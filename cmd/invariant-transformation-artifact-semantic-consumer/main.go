package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/reportconsumer"
)

func main() {
	artifactPath := flag.String("artifact", "", "generated artifact to consume")
	expectedPath := flag.String("expected-projection", "", "expected artifact semantic projection")
	reportPath := flag.String("report", "", "baseline bound report")
	sourcePath := flag.String("source", "", "original Gooo source")
	headSHA := flag.String("head-sha", "", "exact subject commit")
	flag.Parse()
	if *artifactPath == "" || *expectedPath == "" || *reportPath == "" || *sourcePath == "" || *headSHA == "" {
		fail("-artifact, -expected-projection, -report, -source, and -head-sha are required")
	}
	expected, err := decodeProjection(*expectedPath)
	if err != nil {
		fail(err.Error())
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	var report model.Report
	if err := decodeStrictFile(*reportPath, &report); err != nil {
		fail(err.Error())
	}
	baseline, err := reportconsumer.ArtifactProjection(report, source, *headSHA)
	if err != nil {
		fail(err.Error())
	}
	if err := reportconsumer.ValidateArtifactProjectionBinding(expected, baseline); err != nil {
		fail(err.Error())
	}
	actual, err := reportconsumer.ProjectArtifactFile(*artifactPath, *headSHA)
	if err != nil {
		fail(err.Error())
	}
	if err := reportconsumer.CompareArtifactSemanticProjection(expected, actual); err != nil {
		fail(err.Error())
	}
	if err := reportconsumer.CompareArtifactProvenance(expected, actual); err != nil {
		fail(err.Error())
	}
	fmt.Printf("artifact semantic projection matches: case=%s execution=%s raw=%s semantic=%s\n", actual.CaseID, actual.ExecutionID, actual.RawDigest, actual.SemanticDigest)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func decodeProjection(path string) (reportconsumer.ArtifactSemanticProjection, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return reportconsumer.ArtifactSemanticProjection{}, err
	}
	return reportconsumer.DecodeArtifactSemanticProjection(raw)
}

func decodeStrictFile(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("strict JSON decode %s: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("strict JSON decode %s: trailing JSON", path)
		}
		return fmt.Errorf("strict JSON decode %s: %w", path, err)
	}
	return nil
}
