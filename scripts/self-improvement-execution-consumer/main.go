package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "", "reserve or consume")
	grantPath := flag.String("grant", "", "verified v26 grant report")
	contractPath := flag.String("contract", "", "v25 contract report")
	metadataPath := flag.String("metadata", "", "workflow-bound source metadata")
	reservationPath := flag.String("reservation", "", "uploaded one-use reservation")
	outputPath := flag.String("output", "", "output JSON path")
	flag.Parse()
	if *mode != "reserve" && *mode != "consume" {
		fatal("-mode must be reserve or consume")
	}
	if *grantPath == "" || *contractPath == "" || *metadataPath == "" || *outputPath == "" {
		fatal("-grant, -contract, -metadata, and -output are required")
	}
	grantReport, contractReport, metadata, err := loadAndValidate(*grantPath, *contractPath, *metadataPath)
	if err != nil {
		fatal(err.Error())
	}
	if *mode == "reserve" {
		err = writeJSON(*outputPath, newReservation(grantReport, contractReport, metadata))
	} else {
		var reserved reservation
		if *reservationPath == "" {
			fatal("-reservation is required for consume mode")
		}
		if err = readJSON(*reservationPath, &reserved); err == nil {
			err = validateReservation(reserved)
		}
		if err == nil && (reserved.GrantID != grantReport.Resolution.Receipt.GrantID ||
			reserved.RequestDigest != grantReport.Request.Digest || reserved.SubjectSHA != metadata.SubjectSHA ||
			reserved.SourceRunID != metadata.GrantRunID || reserved.SourceAttempt != metadata.GrantAttempt) {
			err = fmt.Errorf("reservation identity does not match the exact grant")
		}
		if err == nil {
			err = consume(grantReport, contractReport, metadata, reserved, *outputPath)
		}
	}
	if err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
