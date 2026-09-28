package languagesyntax

import (
	"errors"
	"strconv"
	"strings"
)

const (
	LanguageSyntaxDeclarationProvenanceSchema  = "gooo/language-syntax-declaration-provenance/v1"
	LanguageSyntaxDeclarationProvenanceClosed  = "CLOSED"
	LanguageSyntaxDeclarationProvenanceUnknown = "UNKNOWN"
	LanguageSyntaxDeclarationProvenanceRefuted = "REFUTED"
)

type LanguageSyntaxDeclarationProvenanceObservation struct {
	Schema                  string
	Decision                string
	Reason                  string
	RegistryDigest          string
	CorpusDigest            string
	SourceDigest            string
	GoooFilesDigest         string
	MissingRegisteredDigest string
	UnregisteredGoooDigest  string
	RegisteredCount         int
	MissingCount            int
	UnregisteredCount       int
	NonAuthorizing          bool
	ObservationDigest       string
}

func ObserveLanguageSyntaxDeclarationProvenance(source Source) LanguageSyntaxDeclarationProvenanceObservation {
	observation := LanguageSyntaxDeclarationProvenanceObservation{
		Schema:            LanguageSyntaxDeclarationProvenanceSchema,
		Decision:          LanguageSyntaxDeclarationProvenanceUnknown,
		Reason:            "LANGUAGE_SYNTAX_DECLARATION_PROVENANCE_UNKNOWN",
		RegistryDigest:    source.RegistryDigest,
		CorpusDigest:      source.CorpusDigest,
		RegisteredCount:   len(source.GoooFiles),
		MissingCount:      len(source.MissingRegistered),
		UnregisteredCount: len(source.UnregisteredGooo),
		NonAuthorizing:    true,
	}
	if source.RegistryDigest == "" {
		return finalizeLanguageSyntaxDeclarationProvenance(observation, LanguageSyntaxDeclarationProvenanceUnknown, "MISSING_DECLARATION_REGISTRY_DIGEST")
	}
	if !validDigest(source.RegistryDigest) {
		return finalizeLanguageSyntaxDeclarationProvenance(observation, LanguageSyntaxDeclarationProvenanceRefuted, "MALFORMED_DECLARATION_REGISTRY_DIGEST")
	}
	if len(source.GoooFiles) == 0 {
		return finalizeLanguageSyntaxDeclarationProvenance(observation, LanguageSyntaxDeclarationProvenanceUnknown, "MISSING_GOOO_DECLARATION_INVENTORY")
	}
	observation.SourceDigest = digestJSON(source)
	observation.GoooFilesDigest = digestJSON(source.GoooFiles)
	observation.MissingRegisteredDigest = digestJSON(source.MissingRegistered)
	observation.UnregisteredGoooDigest = digestJSON(source.UnregisteredGooo)
	if len(source.MissingRegistered) > 0 || len(source.UnregisteredGooo) > 0 {
		return finalizeLanguageSyntaxDeclarationProvenance(observation, LanguageSyntaxDeclarationProvenanceRefuted, "GOOO_DECLARATION_REGISTRY_DRIFT")
	}
	return finalizeLanguageSyntaxDeclarationProvenance(observation, LanguageSyntaxDeclarationProvenanceClosed, "GOOO_DECLARATION_REGISTRY_BOUND")
}

func finalizeLanguageSyntaxDeclarationProvenance(observation LanguageSyntaxDeclarationProvenanceObservation, decision, reason string) LanguageSyntaxDeclarationProvenanceObservation {
	observation.Decision = decision
	observation.Reason = reason
	observation.ObservationDigest = digestBytes([]byte(observation.Canonical()))
	return observation
}

func (observation LanguageSyntaxDeclarationProvenanceObservation) Canonical() string {
	return strings.Join([]string{
		LanguageSyntaxDeclarationProvenanceSchema,
		observation.Decision,
		observation.Reason,
		observation.RegistryDigest,
		observation.CorpusDigest,
		observation.SourceDigest,
		observation.GoooFilesDigest,
		observation.MissingRegisteredDigest,
		observation.UnregisteredGoooDigest,
		strconv.Itoa(observation.RegisteredCount),
		strconv.Itoa(observation.MissingCount),
		strconv.Itoa(observation.UnregisteredCount),
		strconv.FormatBool(observation.NonAuthorizing),
	}, "\x1f")
}

func (observation LanguageSyntaxDeclarationProvenanceObservation) Validate() error {
	if observation.Schema != LanguageSyntaxDeclarationProvenanceSchema ||
		!observation.NonAuthorizing ||
		observation.Decision == "" ||
		observation.Reason == "" {
		return errors.New("language syntax declaration provenance identity is invalid")
	}
	for _, digest := range []string{
		observation.RegistryDigest,
		observation.CorpusDigest,
		observation.SourceDigest,
		observation.GoooFilesDigest,
		observation.MissingRegisteredDigest,
		observation.UnregisteredGoooDigest,
		observation.ObservationDigest,
	} {
		if digest != "" && !validDigest(digest) {
			return errors.New("language syntax declaration provenance contains an invalid digest")
		}
	}
	if observation.Decision == LanguageSyntaxDeclarationProvenanceClosed &&
		(observation.RegistryDigest == "" ||
			observation.SourceDigest == "" ||
			observation.GoooFilesDigest == "" ||
			observation.MissingRegisteredDigest == "" ||
			observation.UnregisteredGoooDigest == "" ||
			observation.RegisteredCount == 0 ||
			observation.MissingCount != 0 ||
			observation.UnregisteredCount != 0) {
		return errors.New("closed language syntax declaration provenance is incomplete")
	}
	if observation.ObservationDigest != digestBytes([]byte(observation.Canonical())) {
		return errors.New("language syntax declaration provenance digest mismatch")
	}
	return nil
}
