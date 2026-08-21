package semantic

import (
	"strconv"
	"strings"
)

func (e InferenceEvidence) Canonical() string {
	if normalized, err := e.normalized(); err == nil {
		e = normalized
	}
	var b strings.Builder
	b.WriteString("evidence-record\t")
	writeCanonicalField(&b, e.ID.String())
	writeCanonicalField(&b, e.Digest)
	writeSnapshot(&b, e.Before)
	writeSnapshot(&b, e.After)
	writeCanonicalField(&b, strconv.FormatBool(e.SourceBacked))
	writeCanonicalField(&b, strconv.FormatBool(e.Independent))
	writeCanonicalField(&b, e.Controls.CatalogDigest)
	writeCanonicalField(&b, e.Controls.PolicyDigest)
	writeCanonicalField(&b, e.Controls.Profile.ID)
	writeCanonicalField(&b, e.Controls.Profile.Version)
	writeCanonicalField(&b, e.Controls.Profile.Digest)
	return b.String()
}
func (e InferenceEvidence) StableHash() string { return StableHashString(e.Canonical()) }
func (p InferencePathV1) Canonical() string {
	if normalized, err := p.Normalized(); err == nil {
		p = normalized
	}
	var b strings.Builder
	b.WriteString("inference-path\t")
	writeCanonicalField(&b, p.Version)
	b.WriteString("edges\t")
	b.WriteString(strconv.Itoa(len(p.Edges)))
	b.WriteByte('\n')
	for _, edge := range p.Edges {
		b.WriteString(edge.Canonical())
		b.WriteByte('\n')
	}
	b.WriteString("claims\t")
	b.WriteString(strconv.Itoa(len(p.Claims)))
	b.WriteByte('\n')
	for _, claim := range p.Claims {
		b.WriteString(claim.Canonical())
		b.WriteByte('\n')
	}
	b.WriteString("evidence-records\t")
	b.WriteString(strconv.Itoa(len(p.Evidence)))
	b.WriteByte('\n')
	for _, evidence := range p.Evidence {
		b.WriteString(evidence.Canonical())
		b.WriteByte('\n')
	}
	return b.String()
}
