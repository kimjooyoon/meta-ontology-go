// Package bootstrap defines the small, deterministic evidence boundary used by
// the self-hosting bootstrap plan. It deliberately has no dependency on the
// compiler packages so that an independent verifier can consume its records.
package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const SchemaVersion = "gooo.bootstrap/v1"

// Artifact is a content-addressed build input or output.
type Artifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Manifest is the reproducible-build contract for one compiler stage.
type Manifest struct {
	Schema         string     `json:"schema"`
	Stage          string     `json:"stage"`
	SourceDigest   string     `json:"source_digest"`
	CompilerDigest string     `json:"compiler_digest"`
	SemanticDigest string     `json:"semantic_digest"`
	Inputs         []Artifact `json:"inputs"`
	Outputs        []Artifact `json:"outputs"`
}

// Evidence is an append-only, hash-chained claim about a manifest.
type Evidence struct {
	Sequence       uint64 `json:"sequence"`
	Kind           string `json:"kind"`
	Subject        string `json:"subject"`
	ClaimDigest    string `json:"claim_digest"`
	PreviousDigest string `json:"previous_digest,omitempty"`
}

// Attestation binds evidence to exactly one manifest and compiler identity.
type Attestation struct {
	Schema         string     `json:"schema"`
	ManifestDigest string     `json:"manifest_digest"`
	BuilderDigest  string     `json:"builder_digest"`
	Evidence       []Evidence `json:"evidence"`
}

// NewArtifact hashes content without reading from the filesystem. Callers can
// therefore construct the same record on different hosts.
func NewArtifact(filePath string, content []byte) (Artifact, error) {
	if err := validatePath(filePath); err != nil {
		return Artifact{}, err
	}
	return Artifact{Path: filePath, SHA256: DigestBytes(content), Size: int64(len(content))}, nil
}

// DigestBytes returns the lowercase SHA-256 digest used by all records.
func DigestBytes(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

// NewManifest constructs and normalizes a manifest. Input and output order is
// not semantic; canonical serialization sorts both lists by path.
func NewManifest(stage, sourceDigest, compilerDigest, semanticDigest string, inputs, outputs []Artifact) (Manifest, error) {
	manifest := Manifest{
		Schema:         SchemaVersion,
		Stage:          stage,
		SourceDigest:   sourceDigest,
		CompilerDigest: compilerDigest,
		SemanticDigest: semanticDigest,
		Inputs:         inputs,
		Outputs:        outputs,
	}
	return manifest.normalized()
}

// Validate checks the manifest without changing its in-memory representation.
func (m Manifest) Validate() error {
	_, err := m.normalized()
	return err
}

// CanonicalJSON returns deterministic JSON with no host-specific fields.
func (m Manifest) CanonicalJSON() ([]byte, error) {
	normalized, err := m.normalized()
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

// Digest returns the content identity of the canonical manifest.
func (m Manifest) Digest() (string, error) {
	data, err := m.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

// Equivalent compares manifests by their canonical semantic content.
func (m Manifest) Equivalent(other Manifest) bool {
	left, leftErr := m.CanonicalJSON()
	right, rightErr := other.CanonicalJSON()
	return leftErr == nil && rightErr == nil && string(left) == string(right)
}

// NewAttestation starts an empty evidence chain for a valid manifest and
// independently identifies the verifier that produced the claims.
func NewAttestation(manifest Manifest, builderDigest string) (Attestation, error) {
	manifestDigest, err := manifest.Digest()
	if err != nil {
		return Attestation{}, err
	}
	if !validDigest(builderDigest) {
		return Attestation{}, fmt.Errorf("attestation builder digest is not a lowercase SHA-256 digest")
	}
	return Attestation{
		Schema:         SchemaVersion,
		ManifestDigest: manifestDigest,
		BuilderDigest:  builderDigest,
		Evidence:       []Evidence{},
	}, nil
}

// AddEvidence appends one claim and derives its sequence and chain link.
func (a Attestation) AddEvidence(kind, subject, claimDigest string) (Attestation, error) {
	evidence := Evidence{Sequence: uint64(len(a.Evidence) + 1), Kind: kind, Subject: subject, ClaimDigest: claimDigest}
	if len(a.Evidence) > 0 {
		previous, err := a.Evidence[len(a.Evidence)-1].Digest()
		if err != nil {
			return Attestation{}, err
		}
		evidence.PreviousDigest = previous
	}
	return a.AppendEvidence(evidence)
}

// AppendEvidence accepts a pre-built record only when it extends the chain.
func (a Attestation) AppendEvidence(evidence Evidence) (Attestation, error) {
	if err := validateEvidence(evidence); err != nil {
		return Attestation{}, err
	}
	if evidence.Sequence != uint64(len(a.Evidence)+1) {
		return Attestation{}, fmt.Errorf("evidence sequence must be %d", len(a.Evidence)+1)
	}
	if len(a.Evidence) == 0 {
		if evidence.PreviousDigest != "" {
			return Attestation{}, fmt.Errorf("first evidence record cannot have a predecessor")
		}
	} else {
		previous, err := a.Evidence[len(a.Evidence)-1].Digest()
		if err != nil {
			return Attestation{}, err
		}
		if evidence.PreviousDigest != previous {
			return Attestation{}, fmt.Errorf("evidence predecessor does not match record %d", len(a.Evidence))
		}
	}
	result := a
	result.Evidence = append(append([]Evidence(nil), a.Evidence...), evidence)
	return result, nil
}

// Validate checks the attestation's binding and every evidence chain link.
func (a Attestation) Validate(manifest Manifest) error {
	if a.Schema != SchemaVersion {
		return fmt.Errorf("unsupported attestation schema %q", a.Schema)
	}
	manifestDigest, err := manifest.Digest()
	if err != nil {
		return err
	}
	if a.ManifestDigest != manifestDigest {
		return fmt.Errorf("attestation manifest digest does not match")
	}
	if !validDigest(a.BuilderDigest) {
		return fmt.Errorf("attestation builder digest is not a lowercase SHA-256 digest")
	}
	for index, evidence := range a.Evidence {
		if evidence.Sequence != uint64(index+1) {
			return fmt.Errorf("evidence sequence at index %d is invalid", index)
		}
		if err := validateEvidence(evidence); err != nil {
			return err
		}
		if index == 0 {
			if evidence.PreviousDigest != "" {
				return fmt.Errorf("first evidence record cannot have a predecessor")
			}
			continue
		}
		previous, err := a.Evidence[index-1].Digest()
		if err != nil {
			return err
		}
		if evidence.PreviousDigest != previous {
			return fmt.Errorf("evidence predecessor at index %d is invalid", index)
		}
	}
	return nil
}

// CanonicalJSON returns deterministic attestation JSON.
func (a Attestation) CanonicalJSON() ([]byte, error) {
	if a.Schema != SchemaVersion || !validDigest(a.ManifestDigest) || !validDigest(a.BuilderDigest) {
		return nil, fmt.Errorf("invalid attestation header")
	}
	if err := validateEvidenceChain(a.Evidence); err != nil {
		return nil, err
	}
	normalized := a
	normalized.Evidence = append([]Evidence{}, a.Evidence...)
	return json.Marshal(normalized)
}

// Digest returns the identity of the complete attestation.
func (a Attestation) Digest() (string, error) {
	data, err := a.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

// EvidenceDigest compares verifier claims while ignoring builder metadata.
func (a Attestation) EvidenceDigest() (string, error) {
	if err := validateEvidenceChain(a.Evidence); err != nil {
		return "", err
	}
	data, err := json.Marshal(append([]Evidence{}, a.Evidence...))
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

// Digest returns the identity of one evidence record, including its chain link.
func (e Evidence) Digest() (string, error) {
	if err := validateEvidence(e); err != nil {
		return "", err
	}
	data, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}
