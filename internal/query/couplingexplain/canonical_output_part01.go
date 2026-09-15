package couplingexplain

import (
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"sort"
)

type canonicalExplanation struct {
	Status         Status             `json:"status"`
	EvidenceDigest string             `json:"evidence_digest"`
	Binding        SnapshotBinding    `json:"binding"`
	Upstream       *canonicalUpstream `json:"upstream,omitempty"`
	Link           *canonicalLink     `json:"link,omitempty"`
	NoLink         *NoLink            `json:"no_link,omitempty"`
	Diagnostics    []Diagnostic       `json:"diagnostics,omitempty"`
}
type canonicalLink struct {
	CodeBinding   canonicalCodeBinding `json:"code_binding"`
	SemanticOwner string               `json:"semantic_owner"`
	Term          canonicalTerm        `json:"term"`
	OriginPath    canonicalPath        `json:"origin_path"`
	Receipt       canonicalReceipt     `json:"receipt"`
	Verifier      canonicalVerifier    `json:"verifier"`
}
type canonicalUpstream struct {
	SourceMapDigest      string            `json:"source_map_digest,omitempty"`
	ManifestDigest       string            `json:"manifest_digest"`
	DetectorInputDigest  string            `json:"detector_input_digest"`
	DetectorResultDigest string            `json:"detector_result_digest"`
	DetectorStatus       detector.Status   `json:"detector_status"`
	DetectorReasons      []detector.Reason `json:"detector_reasons,omitempty"`
	ManifestStatus       string            `json:"manifest_status,omitempty"`
	ManifestReason       string            `json:"manifest_reason,omitempty"`
}

func toCanonicalUpstream(value *UpstreamEvidence) *canonicalUpstream {
	if value == nil {
		return nil
	}
	reasons := append([]detector.Reason(nil), value.DetectorReasons...)
	sort.Slice(reasons, func(i, j int) bool {
		if reasons[i].Code != reasons[j].Code {
			return reasons[i].Code < reasons[j].Code
		}
		return reasons[i].Detail < reasons[j].Detail
	})
	return &canonicalUpstream{
		SourceMapDigest: value.SourceMapDigest, ManifestDigest: value.ManifestDigest,
		DetectorInputDigest: value.DetectorInputDigest, DetectorResultDigest: value.DetectorResultDigest,
		DetectorStatus: value.DetectorStatus, DetectorReasons: reasons,
		ManifestStatus: string(value.ManifestStatus), ManifestReason: string(value.ManifestReason),
	}
}
func toCanonicalLink(value ExplanationLink, expanded bool) *canonicalLink {
	path := toCanonicalPath(value.OriginPath)
	if !expanded {
		path.Steps = nil
	}
	return &canonicalLink{CodeBinding: toCanonicalCodeBinding(value.CodeBinding), SemanticOwner: value.SemanticOwner,
		Term: toCanonicalTerm(value.Term), OriginPath: path, Receipt: toCanonicalReceipt(value.Receipt), Verifier: toCanonicalVerifier(value.Verifier)}
}
func canonicalDiagnostics(values []Diagnostic) []Diagnostic {
	diagnostics := make([]Diagnostic, 0, len(values))
	for _, value := range values {
		diagnostics = append(diagnostics, Diagnostic{Code: value.Code, IDs: sortedStrings(value.IDs)})
	}
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Code != diagnostics[j].Code {
			return diagnostics[i].Code < diagnostics[j].Code
		}
		return joinIDs(diagnostics[i].IDs) < joinIDs(diagnostics[j].IDs)
	})
	return diagnostics
}
