package languagesyntax

import (
	"errors"
	"strings"
)

const (
	LanguageSyntaxReadinessProvenanceSchema  = "gooo/language-syntax-readiness-provenance/v1"
	LanguageSyntaxReadinessProvenanceClosed  = "CLOSED"
	LanguageSyntaxReadinessProvenanceUnknown = "UNKNOWN"
	LanguageSyntaxReadinessProvenanceRefuted = "REFUTED"
)

type LanguageSyntaxReadinessProvenanceObservation struct {
	Schema                       string
	Decision                     string
	Reason                       string
	ReportDecision               string
	ReportReason                 string
	DeclarationDecision          string
	DeclarationReason            string
	ReportObservationDigest      string
	DeclarationObservationDigest string
	ReverseObservationDigest     string
	NonAuthorizing               bool
	ObservationDigest            string
}

func ObserveLanguageSyntaxReadinessProvenance(report Report) LanguageSyntaxReadinessProvenanceObservation {
	reportObservation := ObserveLanguageSyntaxProvenance(report)
	declarationObservation := ObserveLanguageSyntaxDeclarationProvenance(report.Source)
	observation := LanguageSyntaxReadinessProvenanceObservation{
		Schema:                       LanguageSyntaxReadinessProvenanceSchema,
		Decision:                     LanguageSyntaxReadinessProvenanceUnknown,
		Reason:                       "LANGUAGE_SYNTAX_READINESS_PROVENANCE_UNKNOWN",
		ReportDecision:               reportObservation.Decision,
		ReportReason:                 reportObservation.Reason,
		DeclarationDecision:          declarationObservation.Decision,
		DeclarationReason:            declarationObservation.Reason,
		ReportObservationDigest:      reportObservation.ObservationDigest,
		DeclarationObservationDigest: declarationObservation.ObservationDigest,
		NonAuthorizing:               true,
	}
	if reportObservation.Decision == LanguageSyntaxProvenanceRefuted ||
		declarationObservation.Decision == LanguageSyntaxDeclarationProvenanceRefuted {
		observation.Decision = LanguageSyntaxReadinessProvenanceRefuted
		observation.Reason = "LANGUAGE_SYNTAX_READINESS_EVIDENCE_REFUTED"
		return finalizeLanguageSyntaxReadinessProvenance(observation)
	}
	if reportObservation.Decision != LanguageSyntaxProvenanceClosed ||
		declarationObservation.Decision != LanguageSyntaxDeclarationProvenanceClosed {
		observation.Decision = LanguageSyntaxReadinessProvenanceUnknown
		observation.Reason = "LANGUAGE_SYNTAX_READINESS_EVIDENCE_UNKNOWN"
		return finalizeLanguageSyntaxReadinessProvenance(observation)
	}
	observation.ReverseObservationDigest = languageSyntaxReadinessReverseObservationDigest(
		reportObservation.ObservationDigest,
		declarationObservation.ObservationDigest,
	)
	observation.Decision = LanguageSyntaxReadinessProvenanceClosed
	observation.Reason = "LANGUAGE_SYNTAX_READINESS_EVIDENCE_COMPOSED"
	return finalizeLanguageSyntaxReadinessProvenance(observation)
}

func languageSyntaxReadinessReverseObservationDigest(reportDigest, declarationDigest string) string {
	return digestBytes([]byte(strings.Join([]string{
		"gooo/language-syntax-readiness-reverse-observation/v1",
		reportDigest,
		declarationDigest,
	}, "\x1f")))
}

func finalizeLanguageSyntaxReadinessProvenance(observation LanguageSyntaxReadinessProvenanceObservation) LanguageSyntaxReadinessProvenanceObservation {
	observation.ObservationDigest = digestBytes([]byte(observation.Canonical()))
	return observation
}

func (observation LanguageSyntaxReadinessProvenanceObservation) Canonical() string {
	return strings.Join([]string{
		LanguageSyntaxReadinessProvenanceSchema,
		observation.Decision,
		observation.Reason,
		observation.ReportDecision,
		observation.ReportReason,
		observation.DeclarationDecision,
		observation.DeclarationReason,
		observation.ReportObservationDigest,
		observation.DeclarationObservationDigest,
		observation.ReverseObservationDigest,
		strconvBool(observation.NonAuthorizing),
	}, "\x1f")
}

func strconvBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (observation LanguageSyntaxReadinessProvenanceObservation) Validate() error {
	if observation.Schema != LanguageSyntaxReadinessProvenanceSchema ||
		!observation.NonAuthorizing ||
		observation.Decision == "" ||
		observation.Reason == "" {
		return errors.New("language syntax readiness provenance identity is invalid")
	}
	for _, digest := range []string{
		observation.ReportObservationDigest,
		observation.DeclarationObservationDigest,
		observation.ReverseObservationDigest,
		observation.ObservationDigest,
	} {
		if digest != "" && !validDigest(digest) {
			return errors.New("language syntax readiness provenance contains an invalid digest")
		}
	}
	if observation.Decision == LanguageSyntaxReadinessProvenanceClosed &&
		(observation.ReportObservationDigest == "" ||
			observation.DeclarationObservationDigest == "" ||
			observation.ReverseObservationDigest == "") {
		return errors.New("closed language syntax readiness provenance is incomplete")
	}
	if observation.ReverseObservationDigest != "" &&
		observation.ReverseObservationDigest != languageSyntaxReadinessReverseObservationDigest(
			observation.ReportObservationDigest,
			observation.DeclarationObservationDigest,
		) {
		return errors.New("language syntax readiness reverse observation mismatch")
	}
	if observation.ObservationDigest != digestBytes([]byte(observation.Canonical())) {
		return errors.New("language syntax readiness observation digest mismatch")
	}
	return nil
}
