package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/reportconsumer"
)

func main() {
	sourcePath := flag.String("source", "", "original Gooo source")
	headSHA := flag.String("head-sha", "", "exact subject commit")
	firstReportPath := flag.String("first-report", "", "first preliminary report")
	secondReportPath := flag.String("second-report", "", "second preliminary report")
	firstProjectionPath := flag.String("first-projection", "", "first artifact semantic projection")
	secondProjectionPath := flag.String("second-projection", "", "second artifact semantic projection")
	outputTamperPath := flag.String("output-tamper", "", "artifact with semantic output tamper")
	authorizationTamperPath := flag.String("authorization-tamper", "", "artifact with authorization digest tamper")
	interventionReportPath := flag.String("intervention-report", "", "producer intervention report")
	interventionConsumerPath := flag.String("intervention-consumer-receipt", "", "independent intervention consumer receipt")
	outputPath := flag.String("output", "", "final closure receipt")
	flag.Parse()
	if *sourcePath == "" || *headSHA == "" || *firstReportPath == "" || *secondReportPath == "" || *firstProjectionPath == "" || *secondProjectionPath == "" || *outputTamperPath == "" || *authorizationTamperPath == "" || *interventionReportPath == "" || *interventionConsumerPath == "" || *outputPath == "" {
		fail("-source, -head-sha, -first-report, -second-report, -first-projection, -second-projection, -output-tamper, -authorization-tamper, -intervention-report, -intervention-consumer-receipt, and -output are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	var firstReport, secondReport model.Report
	if err := decodeStrictFile(*firstReportPath, &firstReport); err != nil {
		fail(err.Error())
	}
	if err := decodeStrictFile(*secondReportPath, &secondReport); err != nil {
		fail(err.Error())
	}
	interventionReport, err := os.ReadFile(*interventionReportPath)
	if err != nil {
		fail(err.Error())
	}
	interventionConsumer, err := os.ReadFile(*interventionConsumerPath)
	if err != nil {
		fail(err.Error())
	}
	firstProjection, err := decodeProjection(*firstProjectionPath)
	if err != nil {
		fail(err.Error())
	}
	secondProjection, err := decodeProjection(*secondProjectionPath)
	if err != nil {
		fail(err.Error())
	}
	closure, err := reportconsumer.Close(firstReport, secondReport, firstProjection, secondProjection, source, *headSHA, *outputTamperPath, *authorizationTamperPath, interventionReport, interventionConsumer)
	if err != nil {
		fail(err.Error())
	}
	raw, err := json.MarshalIndent(closure, "", "  ")
	if err != nil {
		fail(fmt.Sprintf("encode closure receipt: %v", err))
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, append(raw, '\n'), 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("artifact closure: decision=%s observed=%d/%d semantic=%d/%d raw=%d/%d authorization=%d/%d tamper=%d/%d\n", closure.Decision, closure.ArtifactBytesObserved, closure.ExpectedArtifactBytes, closure.SemanticEqualityObserved, closure.ExpectedSemanticEquality, closure.RawProvenanceBindings, closure.ExpectedRawProvenanceBindings, closure.AuthorizationBindings, closure.ExpectedAuthorizationBindings, closure.OutputSemanticTamperRejected+closure.AuthorizationTamperRejected, closure.ExpectedOutputSemanticTamperRejections+closure.ExpectedAuthorizationTamperRejections)
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

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
