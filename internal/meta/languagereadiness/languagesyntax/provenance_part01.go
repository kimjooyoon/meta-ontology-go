package languagesyntax

import (
	"errors"
	"strconv"
	"strings"
)

const (
	LanguageSyntaxProvenanceSchema  = "gooo/language-syntax-provenance/v1"
	LanguageSyntaxProvenanceClosed  = "CLOSED"
	LanguageSyntaxProvenanceUnknown = "UNKNOWN"
	LanguageSyntaxProvenanceRefuted = "REFUTED"
)

type LanguageSyntaxProvenanceObservation struct {
	Schema                   string
	Decision                 string
	Reason                   string
	ReportedDecision         string
	ReportedReason           string
	HeadSHA                  string
	SourceDigest             string
	ReportDigest             string
	CaseEvidenceDigest       string
	IndicatorDigest          string
	ReverseObservationDigest string
	Satisfied                int
	Total                    int
	Unresolved               int
	NonAuthorizing           bool
	ObservationDigest        string
}

func ObserveLanguageSyntaxProvenance(report Report) LanguageSyntaxProvenanceObservation {
	observation := LanguageSyntaxProvenanceObservation{
		Schema:           LanguageSyntaxProvenanceSchema,
		Decision:         LanguageSyntaxProvenanceUnknown,
		Reason:           "LANGUAGE_SYNTAX_PROVENANCE_UNKNOWN",
		ReportedDecision: report.Decision,
		ReportedReason:   report.Reason,
		HeadSHA:          report.HeadSHA,
		ReportDigest:     report.ReportDigest,
		Satisfied:        report.Summary.Satisfied,
		Total:            report.Summary.Total,
		Unresolved:       report.Summary.Unresolved,
		NonAuthorizing:   true,
	}
	if report.Schema == "" {
		return finalizeLanguageSyntaxProvenance(observation, LanguageSyntaxProvenanceUnknown, "MISSING_REPORT_SCHEMA")
	}
	if report.Schema != ReportSchema {
		return finalizeLanguageSyntaxProvenance(observation, LanguageSyntaxProvenanceRefuted, "MALFORMED_REPORT_SCHEMA")
	}
	if report.HeadSHA == "" {
		return finalizeLanguageSyntaxProvenance(observation, LanguageSyntaxProvenanceUnknown, "MISSING_HEAD_SHA")
	}
	if !validHead(report.HeadSHA) {
		return finalizeLanguageSyntaxProvenance(observation, LanguageSyntaxProvenanceRefuted, "MALFORMED_HEAD_SHA")
	}
	if report.ReportDigest == "" {
		return finalizeLanguageSyntaxProvenance(observation, LanguageSyntaxProvenanceUnknown, "MISSING_REPORT_DIGEST")
	}
	if !validDigest(report.ReportDigest) {
		return finalizeLanguageSyntaxProvenance(observation, LanguageSyntaxProvenanceRefuted, "MALFORMED_REPORT_DIGEST")
	}
	expected := report
	expected.ReportDigest = ""
	if digestJSON(expected) != report.ReportDigest {
		return finalizeLanguageSyntaxProvenance(observation, LanguageSyntaxProvenanceRefuted, "REPORT_DIGEST_MISMATCH")
	}
	observation.SourceDigest = digestJSON(report.Source)
	observation.CaseEvidenceDigest = digestJSON(report.Cases)
	observation.IndicatorDigest = digestJSON(report.Indicators)
	observation.ReverseObservationDigest = languageSyntaxReverseObservationDigest(
		report.ReportDigest,
		observation.SourceDigest,
		observation.CaseEvidenceDigest,
		observation.IndicatorDigest,
	)
	return finalizeLanguageSyntaxProvenance(observation, LanguageSyntaxProvenanceClosed, "LANGUAGE_SYNTAX_EVIDENCE_BOUND")
}

func languageSyntaxReverseObservationDigest(reportDigest, sourceDigest, caseDigest, indicatorDigest string) string {
	return digestBytes([]byte(strings.Join([]string{
		"gooo/language-syntax-reverse-observation/v1",
		reportDigest,
		sourceDigest,
		caseDigest,
		indicatorDigest,
	}, "\x1f")))
}

func finalizeLanguageSyntaxProvenance(observation LanguageSyntaxProvenanceObservation, decision, reason string) LanguageSyntaxProvenanceObservation {
	observation.Decision = decision
	observation.Reason = reason
	observation.ObservationDigest = digestBytes([]byte(observation.Canonical()))
	return observation
}

func (observation LanguageSyntaxProvenanceObservation) Canonical() string {
	return strings.Join([]string{
		LanguageSyntaxProvenanceSchema,
		observation.Decision,
		observation.Reason,
		observation.ReportedDecision,
		observation.ReportedReason,
		observation.HeadSHA,
		observation.SourceDigest,
		observation.ReportDigest,
		observation.CaseEvidenceDigest,
		observation.IndicatorDigest,
		observation.ReverseObservationDigest,
		strconv.Itoa(observation.Satisfied),
		strconv.Itoa(observation.Total),
		strconv.Itoa(observation.Unresolved),
		strconv.FormatBool(observation.NonAuthorizing),
	}, "\x1f")
}

func (observation LanguageSyntaxProvenanceObservation) Validate() error {
	if observation.Schema != LanguageSyntaxProvenanceSchema {
		return errors.New("language syntax provenance schema mismatch")
	}
	if !observation.NonAuthorizing {
		return errors.New("language syntax provenance must be non-authorizing")
	}
	if observation.Decision == "" || observation.Reason == "" {
		return errors.New("language syntax provenance decision is incomplete")
	}
	if observation.HeadSHA != "" && !validHead(observation.HeadSHA) {
		return errors.New("language syntax provenance head is malformed")
	}
	for _, digest := range []string{
		observation.SourceDigest,
		observation.ReportDigest,
		observation.CaseEvidenceDigest,
		observation.IndicatorDigest,
		observation.ReverseObservationDigest,
		observation.ObservationDigest,
	} {
		if digest != "" && !validDigest(digest) {
			return errors.New("language syntax provenance digest is malformed")
		}
	}
	if observation.Decision == LanguageSyntaxProvenanceClosed {
		if observation.HeadSHA == "" ||
			observation.SourceDigest == "" ||
			observation.ReportDigest == "" ||
			observation.CaseEvidenceDigest == "" ||
			observation.IndicatorDigest == "" ||
			observation.ReverseObservationDigest == "" {
			return errors.New("closed language syntax provenance is incomplete")
		}
	}
	if observation.ReverseObservationDigest != "" &&
		observation.ReverseObservationDigest != languageSyntaxReverseObservationDigest(
			observation.ReportDigest,
			observation.SourceDigest,
			observation.CaseEvidenceDigest,
			observation.IndicatorDigest,
		) {
		return errors.New("language syntax reverse observation mismatch")
	}
	if observation.ObservationDigest != digestBytes([]byte(observation.Canonical())) {
		return errors.New("language syntax observation digest mismatch")
	}
	return nil
}
