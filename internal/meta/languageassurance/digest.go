package languageassurance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
)

func digest(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func seal(report *Report) {
	copy := *report
	copy.ReportDigest = ""
	report.ReportDigest = digest(copy)
}

func detectSnapshotMismatches(subjectSHA string, transaction Transaction) []Finding {
	var findings []Finding
	for _, binding := range transaction.SnapshotBindings {
		if binding.SubjectSHA != subjectSHA {
			findings = append(findings, Finding{MetricID: MetricSnapshotBinding, PathID: "snapshot/" + binding.EvidenceID, EvidenceID: binding.EvidenceID, ExpectedSHA: subjectSHA, ObservedSHA: binding.SubjectSHA})
		}
	}
	return findings
}

func observeSnapshotBindings(bindings []SnapshotBinding, mismatches int) (*int, *int) {
	if len(bindings) != len(snapshotEvidenceIDs) {
		return nil, nil
	}
	exact, paths := len(snapshotEvidenceIDs)-mismatches, mismatches
	bps := exact * 10000 / len(snapshotEvidenceIDs)
	return &bps, &paths
}

func Validate(report Report) error {
	if report.Schema != ReportSchema || report.DenominatorID != DenominatorID {
		return fmt.Errorf("language assurance report identity is malformed")
	}
	if len(report.Denominator) != 12 || len(report.Obligations) != 12 || len(report.MetaOperations) != 7 || len(report.Indicators) != 7 {
		return fmt.Errorf("language assurance report cardinality is malformed")
	}
	if report.ReportDigest == "" {
		return fmt.Errorf("language assurance report digest is missing")
	}
	expected, err := Evaluate(report.SubjectSHA, report.Transaction)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(report, expected) {
		return fmt.Errorf("language assurance report does not replay")
	}
	return nil
}

func ValidateForSubject(report Report, subjectSHA string) error {
	if err := Validate(report); err != nil {
		return err
	}
	if report.SubjectSHA != subjectSHA {
		return fmt.Errorf("language assurance report does not bind the exact subject")
	}
	return nil
}
