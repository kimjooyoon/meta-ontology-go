package bootstrap

import "testing"

func TestManifestCanonicalizesArtifactOrder(t *testing.T) {
	first, err := NewArtifact("src/main.gooo", []byte("package demo"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewArtifact("src/types.gooo", []byte("entity Demo"))
	if err != nil {
		t.Fatal(err)
	}
	m1 := newManifest(t, []Artifact{first, second})
	m2 := newManifest(t, []Artifact{second, first})
	left, err := m1.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	right, err := m2.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(left) != string(right) || !m1.Equivalent(m2) {
		t.Fatalf("artifact order changed canonical manifest:\n%s\n%s", left, right)
	}
	leftDigest, err := m1.Digest()
	if err != nil {
		t.Fatal(err)
	}
	rightDigest, err := m2.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if leftDigest != rightDigest {
		t.Fatalf("equivalent manifests have different digests: %s != %s", leftDigest, rightDigest)
	}
}

func TestManifestRejectsUnsafeOrDuplicateArtifacts(t *testing.T) {
	valid, err := NewArtifact("src/main.gooo", []byte("source"))
	if err != nil {
		t.Fatal(err)
	}
	for _, artifact := range []Artifact{
		{Path: "../escape", SHA256: valid.SHA256, Size: 1},
		{Path: "src//main.gooo", SHA256: valid.SHA256, Size: 1},
		{Path: "src/main.gooo", SHA256: "BAD", Size: 1},
		{Path: "src/negative", SHA256: valid.SHA256, Size: -1},
	} {
		if _, err := NewManifest("stage-0", digest("source"), digest("compiler"), digest("semantic"), []Artifact{artifact}, nil); err == nil {
			t.Fatalf("invalid artifact was accepted: %#v", artifact)
		}
	}
	if _, err := NewManifest("stage-0", digest("source"), digest("compiler"), digest("semantic"), []Artifact{valid, valid}, nil); err == nil {
		t.Fatal("duplicate artifact path was accepted")
	}
}

func TestAttestationBindsManifestAndChainsEvidence(t *testing.T) {
	manifest := newManifest(t, nil)
	attestation, err := NewAttestation(manifest, digest("go verifier"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := attestation.AddEvidence("parse", "src/main.gooo", digest("parse"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := first.AddEvidence("semantic-equivalence", "ir", digest("equivalent"))
	if err != nil {
		t.Fatal(err)
	}
	if len(attestation.Evidence) != 0 || len(first.Evidence) != 1 || len(second.Evidence) != 2 {
		t.Fatal("append-only evidence changed a prior attestation")
	}
	if err := second.Validate(manifest); err != nil {
		t.Fatal(err)
	}
	other, err := NewAttestation(manifest, digest("gooo verifier"))
	if err != nil {
		t.Fatal(err)
	}
	other, err = other.AddEvidence("parse", "src/main.gooo", digest("parse"))
	if err != nil {
		t.Fatal(err)
	}
	other, err = other.AddEvidence("semantic-equivalence", "ir", digest("equivalent"))
	if err != nil {
		t.Fatal(err)
	}
	if err := other.Validate(manifest); err != nil {
		t.Fatal(err)
	}
	leftEvidence, err := second.EvidenceDigest()
	if err != nil {
		t.Fatal(err)
	}
	rightEvidence, err := other.EvidenceDigest()
	if err != nil || leftEvidence != rightEvidence {
		t.Fatalf("equivalent verifier evidence differs: %v", err)
	}
	leftAttestation, err := second.Digest()
	if err != nil {
		t.Fatal(err)
	}
	rightAttestation, err := other.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if leftAttestation == rightAttestation {
		t.Fatal("different verifier identities produced the same full attestation")
	}
	jsonBefore, err := second.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	jsonAfter, err := second.CanonicalJSON()
	if err != nil || string(jsonBefore) != string(jsonAfter) {
		t.Fatalf("attestation serialization is not reproducible: %v", err)
	}
	second.Evidence[0].ClaimDigest = digest("tampered")
	if err := second.Validate(manifest); err == nil {
		t.Fatal("tampered evidence chain was accepted")
	}
}

func TestAttestationRejectsWrongPredecessorAndManifest(t *testing.T) {
	manifest := newManifest(t, nil)
	attestation, err := NewAttestation(manifest, digest("go verifier"))
	if err != nil {
		t.Fatal(err)
	}
	invalid := Evidence{Sequence: 2, Kind: "verify", Subject: "ir", ClaimDigest: digest("claim")}
	if _, err := attestation.AppendEvidence(invalid); err == nil {
		t.Fatal("non-contiguous evidence was accepted")
	}
	first, err := attestation.AddEvidence("verify", "ir", digest("claim"))
	if err != nil {
		t.Fatal(err)
	}
	invalid = Evidence{Sequence: 2, Kind: "verify", Subject: "go", ClaimDigest: digest("claim"), PreviousDigest: digest("wrong")}
	if _, err := first.AppendEvidence(invalid); err == nil {
		t.Fatal("wrong evidence predecessor was accepted")
	}
	changed := manifest
	changed.SemanticDigest = digest("different")
	if err := first.Validate(changed); err == nil {
		t.Fatal("attestation for a different manifest was accepted")
	}
}

func newManifest(t *testing.T, inputs []Artifact) Manifest {
	t.Helper()
	manifest, err := NewManifest("stage-0", digest("source"), digest("compiler"), digest("semantic"), inputs, nil)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func digest(value string) string {
	return DigestBytes([]byte(value))
}
