package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	metaff "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/toolchainformatfix"
)

const maxSavedReportBytes = 32 << 20

func readComparedReport(filename string) (metaff.Report, error) {
	return readComparedReportLimit(filename, maxSavedReportBytes)
}

func readComparedReportLimit(filename string, limit int) (metaff.Report, error) {
	var report metaff.Report
	if limit <= 0 || limit > maxSavedReportBytes {
		return report, fmt.Errorf("saved report read limit is out of range")
	}
	file, err := os.Open(filename)
	if err != nil {
		return report, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if err != nil {
		return report, err
	}
	// Reuse JSON key, UTF-8, depth and trailing-data checks for the bytes read.
	if err := completeness.CheckJSON(raw, limit); err != nil {
		return report, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&report)
	return report, err
}
