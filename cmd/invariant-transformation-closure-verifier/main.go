package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/closureverifier"
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
	firstReportRaw, err := os.ReadFile(*firstReportPath)
	if err != nil {
		fail(err.Error())
	}
	secondReportRaw, err := os.ReadFile(*secondReportPath)
	if err != nil {
		fail(err.Error())
	}
	firstProjectionRaw, err := os.ReadFile(*firstProjectionPath)
	if err != nil {
		fail(err.Error())
	}
	secondProjectionRaw, err := os.ReadFile(*secondProjectionPath)
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
	if err := closureverifier.Verify(closureRaw, firstReportRaw, secondReportRaw, firstProjectionRaw, secondProjectionRaw, source, *headSHA, *outputTamperPath, *authorizationTamperPath, interventionReport, interventionConsumer); err != nil {
		fail(err.Error())
	}
	fmt.Println("closure verified: independent package decision=PASS metrics=11/11")
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
