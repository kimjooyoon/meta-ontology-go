package cache

import (
	"testing"
)

func TestProjectionIdentityRequiresTypedFreshness(t *testing.T) {
	key, err := NewProjectionKey(projectionSpec())
	if err != nil {
		t.Fatal(err)
	}
	evidence := evidenceFixture("projection-hit")
	identity := ProjectionIdentity{
		SemanticClosureDigest: key.SemanticClosureDigest, SourceDigest: evidence.SourceDigest,
		IRDigest: evidence.IRDigest, OptionsDigest: key.OptionsDigest,
		Toolchain: key.Toolchain, ToolchainDigest: evidence.ToolchainDigest,
	}
	if err := identity.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*ProjectionIdentity){
		"source":    func(i *ProjectionIdentity) { i.SourceDigest = HashBytes([]byte("other-source")) },
		"IR":        func(i *ProjectionIdentity) { i.IRDigest = HashBytes([]byte("other-ir")) },
		"options":   func(i *ProjectionIdentity) { i.OptionsDigest = HashBytes([]byte("other-options")) },
		"toolchain": func(i *ProjectionIdentity) { i.Toolchain = "go1.26.6" },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := identity
			mutate(&mutated)
			if mutated.matchesKey(key) && mutated.matchesEvidence(evidence) {
				t.Fatal("identity mutation retained both key and evidence identity")
			}
		})
	}
}
