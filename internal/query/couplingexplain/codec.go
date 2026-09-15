package couplingexplain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func (e VerifiedEnvelope) CanonicalJSON() ([]byte, error) {
	diagnostics := canonicalDiagnostics(e.Diagnostics)
	return json.Marshal(canonicalEnvelope{
		Schema: e.Schema, Binding: envelopeBinding(e.Binding),
		Upstream:    toCanonicalUpstream(e.Upstream),
		CodeBinding: toCanonicalCodeBinding(e.CodeBinding), SemanticOwner: e.SemanticOwner,
		Term: toCanonicalTerm(e.Term), OriginPath: toCanonicalPath(e.OriginPath),
		Receipt: toCanonicalReceipt(e.Receipt), Verifier: toCanonicalVerifier(e.Verifier),
		Verdict: e.Verdict, NoLinkReason: e.NoLinkReason,
		EvidenceDigest: e.EvidenceDigest, Diagnostics: diagnostics,
	})
}

// Digest is the exact envelope digest expected in Request. EnvelopeDigest is
// excluded to avoid a self-referential hash.
func (e VerifiedEnvelope) Digest() string {
	data, err := e.CanonicalJSON()
	if err != nil {
		return ""
	}
	return DigestBytes(data)
}

func (e Explanation) CanonicalJSON(view View) ([]byte, error) {
	if view != ViewCompact && view != ViewExpanded {
		return nil, fmt.Errorf("unknown explanation view %q", view)
	}
	value := canonicalExplanation{
		Status: e.Status, EvidenceDigest: e.EvidenceDigest, Binding: e.Binding,
		Upstream:    toCanonicalUpstream(e.Upstream),
		Diagnostics: canonicalDiagnostics(e.Diagnostics),
	}
	if e.NoLink != nil {
		value.NoLink = &NoLink{Reason: e.NoLink.Reason}
	}
	if e.Link != nil {
		value.Link = toCanonicalLink(*e.Link, view == ViewExpanded)
	}
	return json.Marshal(value)
}

func (e Explanation) CanonicalDigest(view View) (string, error) {
	data, err := e.CanonicalJSON(view)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func DigestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func envelopeBinding(value SnapshotBinding) SnapshotBinding {
	value.EnvelopeDigest = ""
	return value
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
