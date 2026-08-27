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
	closurePath := flag.String("closure", "", "final closure receipt")
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
	flag.Parse()
	if *closurePath == "" || *sourcePath == "" || *headSHA == "" || *firstReportPath == "" || *secondReportPath == "" || *firstProjectionPath == "" || *secondProjectionPath == "" || *outputTamperPath == "" || *authorizationTamperPath == "" || *interventionReportPath == "" || *interventionConsumerPath == "" {
		fail("-closure, -source, -head-sha, -first-report, -second-report, -first-projection, -second-projection, -output-tamper, -authorization-tamper, -intervention-report, and -intervention-consumer-receipt are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	closureRaw, err := os.ReadFile(*closurePath)
	if err != nil {
		fail(err.Error())
	}
	closure, err := reportconsumer.DecodeClosureReceipt(closureRaw)
	if err != nil {
		fail(err.Error())
	}
	firstReport, err := decodeReport(*firstReportPath)
	if err != nil {
		fail(err.Error())
	}
	secondReport, err := decodeReport(*secondReportPath)
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
	interventionReport, err := os.ReadFile(*interventionReportPath)
	if err != nil {
		fail(err.Error())
	}
	interventionConsumer, err := os.ReadFile(*interventionConsumerPath)
	if err != nil {
		fail(err.Error())
	}
	if err := reportconsumer.VerifyClosure(closure, firstReport, secondReport, firstProjection, secondProjection, source, *headSHA, *outputTamperPath, *authorizationTamperPath, interventionReport, interventionConsumer); err != nil {
		fail(err.Error())
	}
	fmt.Printf("closure verified: decision=%s metrics=%d/%d\n", closure.Decision, closure.ObservedMetricEvidence, closure.ExpectedMetricEvidence)
}

func decodeReport(path string) (model.Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.Report{}, err
	}
	var report model.Report
	if err := decodeStrict(raw, &report); err != nil {
		return model.Report{}, fmt.Errorf("strict report decode %s: %w", path, err)
	}
	return report, nil
}

func decodeProjection(path string) (reportconsumer.ArtifactSemanticProjection, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return reportconsumer.ArtifactSemanticProjection{}, err
	}
	return reportconsumer.DecodeArtifactSemanticProjection(raw)
}

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON")
		}
		return err
	}
	return nil
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
