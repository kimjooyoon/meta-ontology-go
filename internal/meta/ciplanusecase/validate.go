package ciplanusecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func DecodeReport(raw []byte) (Report, error) {
	report := Report{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return Report{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Report{}, fmt.Errorf("scorecard contains trailing content")
	}
	return report, nil
}

func Validate(report Report) error {
	if report.Schema != ReportSchema {
		return fmt.Errorf("scorecard schema %q is not supported", report.Schema)
	}
	if report.Decision != "PASS" || report.Resolution != "EXACT" {
		return fmt.Errorf("scorecard is not an exact pass")
	}
	if len(report.Cases) != 12 || len(report.Indicators) != 22 || len(report.ReaderViews) != 3 || len(report.Proofs) != 3 {
		return fmt.Errorf("scorecard fixed denominator changed")
	}
	for _, indicator := range report.Indicators {
		if indicator.Status != "SATISFIED" || indicator.Producer == "" || indicator.Consumer == "" || indicator.MetaOperation == "" {
			return fmt.Errorf("indicator %q is not meta-bound and satisfied", indicator.ID)
		}
	}
	if report.ReportDigest != seal(report).ReportDigest {
		return fmt.Errorf("scorecard digest mismatch")
	}
	return nil
}
