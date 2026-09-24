package analysisprovenance

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

// ObservationIdentitySchema identifies the non-authorizing identity envelope
// for one semantic observation. It binds the observed subject to the semantic
// IR and the analysis inputs that produced the observation.
const ObservationIdentitySchema = "gooo/analysis-observation-identity/v1"

// ObservationIdentity binds one source subject, its lowered semantic IR, and
// the analysis inputs. It is observation provenance, not an execution grant,
// an issuer credential, or a promotion decision.
type ObservationIdentity struct {
	Schema           string `json:"schema"`
	SubjectDigest    string `json:"subject_digest"`
	SemanticDigest   string `json:"semantic_digest"`
	ProfileDigest    string `json:"profile_digest"`
	ToolchainDigest  string `json:"toolchain_digest"`
	ContractDigest   string `json:"contract_digest"`
	ProvenanceDigest string `json:"provenance_digest"`
}

// NewObservationIdentity constructs a closed identity from exact digest
// inputs. The subject is the canonical source or artifact digest; it is not a
// path, an actor, or an authorization identity.
func NewObservationIdentity(subjectDigest, semanticDigest, profileDigest, toolchainDigest, contractDigest string) ObservationIdentity {
	return ObservationIdentity{
		Schema:           ObservationIdentitySchema,
		SubjectDigest:    subjectDigest,
		SemanticDigest:   semanticDigest,
		ProfileDigest:    profileDigest,
		ToolchainDigest:  toolchainDigest,
		ContractDigest:   contractDigest,
		ProvenanceDigest: DocumentDigest(subjectDigest, semanticDigest, profileDigest, toolchainDigest, contractDigest),
	}
}

// Validate rejects incomplete or tampered observation identities. A valid
// identity remains evidence only; callers must perform authorization through a
// separate explicit contract.
func (identity ObservationIdentity) Validate() error {
	if identity.Schema != ObservationIdentitySchema {
		return fmt.Errorf("analysis observation identity schema is %q", identity.Schema)
	}
	for name, value := range map[string]string{
		"subject_digest":    identity.SubjectDigest,
		"semantic_digest":   identity.SemanticDigest,
		"profile_digest":    identity.ProfileDigest,
		"toolchain_digest":  identity.ToolchainDigest,
		"contract_digest":   identity.ContractDigest,
		"provenance_digest": identity.ProvenanceDigest,
	} {
		if !cache.Digest(value).Known() {
			return fmt.Errorf("analysis observation identity %s is not a known digest", name)
		}
	}
	want := DocumentDigest(identity.SubjectDigest, identity.SemanticDigest, identity.ProfileDigest, identity.ToolchainDigest, identity.ContractDigest)
	if identity.ProvenanceDigest != want {
		return fmt.Errorf("analysis observation identity provenance digest mismatch")
	}
	return nil
}

// Valid reports whether Validate accepts the identity.
func (identity ObservationIdentity) Valid() bool {
	return identity.Validate() == nil
}
