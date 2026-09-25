package lsp

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const CodeActionProvenanceSchema = "gooo/lsp-code-action-provenance/v1"

const (
	CodeActionProvenanceClosed  = "CLOSED"
	CodeActionProvenanceUnknown = "UNKNOWN"
)

// CodeActionProvenanceEdit binds one proposed edit to the document snapshot
// from which it was derived. It is evidence, not an authorization grant.
type CodeActionProvenanceEdit struct {
	URI        string   `json:"uri"`
	Version    int      `json:"version"`
	Edit       TextEdit `json:"edit"`
	EditDigest string   `json:"edit_digest"`
}

// CodeActionProvenanceObservation describes a safe-to-present edit surface.
// The edit is never applied by this package.
type CodeActionProvenanceObservation struct {
	Schema            string                     `json:"schema"`
	URI               string                     `json:"uri"`
	Version           int                        `json:"version"`
	SourceDigest      string                     `json:"source_digest,omitempty"`
	OriginDigest      string                     `json:"origin_digest,omitempty"`
	EnvironmentDigest string                     `json:"environment_digest,omitempty"`
	Edits             []CodeActionProvenanceEdit `json:"edits"`
	EditMapDigest     string                     `json:"edit_map_digest"`
	Decision          string                     `json:"decision"`
	Reason            string                     `json:"reason"`
	NonAuthorizing    bool                       `json:"non_authorizing"`
	ObservationDigest string                     `json:"observation_digest"`
}

// ObserveCodeActionProvenance records a deterministic edit proposal without
// applying it. A missing or malformed boundary remains UNKNOWN.
func ObserveCodeActionProvenance(uri string, version int, sourceDigest, originDigest, environmentDigest string, edits []TextEdit) CodeActionProvenanceObservation {
	observation := CodeActionProvenanceObservation{
		Schema:            CodeActionProvenanceSchema,
		URI:               strings.TrimSpace(uri),
		Version:           version,
		SourceDigest:      strings.TrimSpace(sourceDigest),
		OriginDigest:      strings.TrimSpace(originDigest),
		EnvironmentDigest: strings.TrimSpace(environmentDigest),
		Edits:             codeActionProvenanceEdits(uri, version, edits),
		Decision:          CodeActionProvenanceUnknown,
		Reason:            "MISSING_SOURCE_DIGEST",
		NonAuthorizing:    true,
	}
	switch {
	case observation.URI == "":
		observation.Reason = "MISSING_DOCUMENT_URI"
	case version < 0:
		observation.Reason = "MISSING_DOCUMENT_VERSION"
	case !cache.Digest(observation.SourceDigest).Known():
		observation.Reason = "MISSING_SOURCE_DIGEST"
	case !cache.Digest(observation.OriginDigest).Known():
		observation.Reason = "MISSING_ORIGIN_DIGEST"
	case !cache.Digest(observation.EnvironmentDigest).Known():
		observation.Reason = "MISSING_ENVIRONMENT_DIGEST"
	case len(observation.Edits) == 0:
		observation.Reason = "NO_SAFE_EDIT"
	default:
		observation.Decision = CodeActionProvenanceClosed
		observation.Reason = "CODE_ACTION_SURFACE_BOUND"
	}
	return finalizeCodeActionProvenance(observation)
}

// ValidateCodeActionProvenance verifies the immutable observation envelope.
func ValidateCodeActionProvenance(value CodeActionProvenanceObservation) error {
	if value.Schema != CodeActionProvenanceSchema || value.URI == "" || value.Version < 0 || !value.NonAuthorizing || value.Reason == "" {
		return errors.New("code action provenance identity is invalid")
	}
	if value.Decision != CodeActionProvenanceClosed && value.Decision != CodeActionProvenanceUnknown {
		return errors.New("code action provenance decision is invalid")
	}
	for _, digest := range []string{value.SourceDigest, value.OriginDigest, value.EnvironmentDigest} {
		if digest != "" && !cache.Digest(digest).Known() {
			return errors.New("code action provenance binding is invalid")
		}
	}
	if !cache.Digest(value.EditMapDigest).Known() || value.EditMapDigest != codeActionProvenanceEditMapDigest(value.Edits) {
		return errors.New("code action provenance edit map is invalid")
	}
	for _, edit := range value.Edits {
		if edit.URI != value.URI || edit.Version != value.Version || !cache.Digest(edit.EditDigest).Known() ||
			edit.EditDigest != codeActionProvenanceEditDigest(edit, value.SourceDigest, value.OriginDigest, value.EnvironmentDigest) {
			return errors.New("code action provenance edit is invalid")
		}
	}
	if value.Decision == CodeActionProvenanceClosed && (len(value.Edits) == 0 ||
		!cache.Digest(value.SourceDigest).Known() || !cache.Digest(value.OriginDigest).Known() || !cache.Digest(value.EnvironmentDigest).Known()) {
		return errors.New("code action provenance closed binding is incomplete")
	}
	if !cache.Digest(value.ObservationDigest).Known() || value.ObservationDigest != codeActionProvenanceObservationDigest(value) {
		return errors.New("code action provenance observation is invalid")
	}
	return nil
}

// ApplyCodeActionProvenance is a stale-snapshot guard. It validates the
// observation and current context, but deliberately performs no edit.
func ApplyCodeActionProvenance(value CodeActionProvenanceObservation, currentVersion int, currentSourceDigest, currentOriginDigest, currentEnvironmentDigest string) error {
	if err := ValidateCodeActionProvenance(value); err != nil {
		return err
	}
	if value.Decision != CodeActionProvenanceClosed {
		return errors.New("code action provenance is unknown")
	}
	if value.Version != currentVersion || value.SourceDigest != currentSourceDigest ||
		value.OriginDigest != currentOriginDigest || value.EnvironmentDigest != currentEnvironmentDigest {
		return errors.New("code action provenance context is stale")
	}
	return nil
}

func codeActionProvenanceEdits(uri string, version int, values []TextEdit) []CodeActionProvenanceEdit {
	result := make([]CodeActionProvenanceEdit, 0, len(values))
	for _, value := range values {
		item := CodeActionProvenanceEdit{URI: strings.TrimSpace(uri), Version: version, Edit: value}
		item.EditDigest = codeActionProvenanceEditDigest(item, "", "", "")
		result = append(result, item)
	}
	sort.SliceStable(result, func(left, right int) bool {
		first, second := result[left], result[right]
		if first.Edit.Range.Start != second.Edit.Range.Start {
			return positionLess(first.Edit.Range.Start, second.Edit.Range.Start)
		}
		if first.Edit.Range.End != second.Edit.Range.End {
			return positionLess(first.Edit.Range.End, second.Edit.Range.End)
		}
		return first.Edit.NewText < second.Edit.NewText
	})
	return result
}

func codeActionProvenanceEditDigest(value CodeActionProvenanceEdit, sourceDigest, originDigest, environmentDigest string) string {
	value.EditDigest = ""
	payload, _ := json.Marshal(struct {
		Edit              CodeActionProvenanceEdit `json:"edit"`
		SourceDigest      string                   `json:"source_digest"`
		OriginDigest      string                   `json:"origin_digest"`
		EnvironmentDigest string                   `json:"environment_digest"`
	}{value, sourceDigest, originDigest, environmentDigest})
	return cache.HashBytes(payload).String()
}

func codeActionProvenanceEditMapDigest(values []CodeActionProvenanceEdit) string {
	payload, _ := json.Marshal(values)
	return cache.HashBytes(payload).String()
}

func finalizeCodeActionProvenance(value CodeActionProvenanceObservation) CodeActionProvenanceObservation {
	for index := range value.Edits {
		value.Edits[index].EditDigest = codeActionProvenanceEditDigest(value.Edits[index], value.SourceDigest, value.OriginDigest, value.EnvironmentDigest)
	}
	value.EditMapDigest = codeActionProvenanceEditMapDigest(value.Edits)
	value.ObservationDigest = codeActionProvenanceObservationDigest(value)
	return value
}

func codeActionProvenanceObservationDigest(value CodeActionProvenanceObservation) string {
	value.ObservationDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}
