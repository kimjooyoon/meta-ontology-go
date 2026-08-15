package semantic

import (
	"strconv"
	"strings"
)

func (r InferenceRecord) canonical(b *strings.Builder, prefix string) {
	b.WriteString(prefix)
	b.WriteByte('\t')
	writeCanonicalField(b, r.RecordID.String())
	writeCanonicalField(b, r.SubjectID.String())
	writeCanonicalField(b, r.ObjectID.String())
	writeCanonicalField(b, r.Rule.ID.String())
	writeCanonicalField(b, r.Rule.Version)
	writeCanonicalField(b, r.Rule.Digest)
	writeCanonicalField(b, string(r.Phase.Phase))
	writeCanonicalField(b, strconv.FormatUint(r.Phase.Ordinal, 10))
	writeSnapshot(b, r.Before)
	writeSnapshot(b, r.After)
	writeCanonicalField(b, string(r.Authority.Layer))
	writeCanonicalField(b, string(r.Authority.Effect))
	writeCanonicalField(b, r.Controls.CatalogDigest)
	writeCanonicalField(b, r.Controls.PolicyDigest)
	writeCanonicalField(b, r.Controls.Profile.ID)
	writeCanonicalField(b, r.Controls.Profile.Version)
	writeCanonicalField(b, r.Controls.Profile.Digest)
	b.WriteString("evidence\t")
	b.WriteString(strconv.Itoa(len(r.Evidence)))
	b.WriteByte('\t')
	for _, ref := range r.Evidence {
		writeCanonicalField(b, ref.ID.String())
		writeCanonicalField(b, ref.Digest)
	}
	b.WriteByte('\n')
}

func writeSnapshot(b *strings.Builder, snapshot SnapshotDigests) {
	writeCanonicalField(b, snapshot.Source)
	writeCanonicalField(b, snapshot.Semantic)
}

func (e InferenceEdge) Canonical() string {
	if normalized, err := e.normalized(); err == nil {
		e = normalized
	}
	var b strings.Builder
	e.InferenceRecord.canonical(&b, "edge")
	writeCanonicalField(&b, string(e.Kind))
	b.WriteString("roots\t")
	b.WriteString(strconv.Itoa(len(e.SourceRoots)))
	b.WriteByte('\t')
	for _, root := range e.SourceRoots {
		writeCanonicalField(&b, root.String())
	}
	writeCanonicalField(&b, e.AcceptanceReceipt.String())
	return b.String()
}

func (e InferenceEdge) StableHash() string { return StableHashString(e.Canonical()) }

func (c SemanticChangeClaim) Canonical() string {
	if normalized, err := c.normalized(); err == nil {
		c = normalized
	}
	var b strings.Builder
	c.InferenceRecord.canonical(&b, "claim")
	writeCanonicalField(&b, string(c.Kind))
	writeCanonicalField(&b, c.CanonicalDelta)
	writeCanonicalField(&b, c.DeltaDigest)
	return b.String()
}

func (c SemanticChangeClaim) StableHash() string { return StableHashString(c.Canonical()) }

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
