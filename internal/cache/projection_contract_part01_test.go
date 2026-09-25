package cache

import (
	"testing"
)

func projectionSpec() ProjectionKeySpec {
	return ProjectionKeySpec{
		Domain: "billing", Version: ProjectionKeyVersion, ArtifactKind: "go-source",
		Projection: "source", HostStage: GoHostedStage,
		SemanticClosureDigest: HashBytes([]byte("semantic-closure")),
		DependencyRoot:        HashBytes([]byte("dependency-root")),
		PolicySchemaDigest:    HashBytes([]byte("policy-schema")),
		Toolchain:             "go1.26.5", Target: "darwin/arm64", BuildTags: []string{"linux"},
		OptionsDigest: mustOptionsDigest(map[string]any{"mode": "fast", "trim": true}),
	}
}
func mustOptionsDigest(value any) Digest {
	digest, err := DigestOptions(value)
	if err != nil {
		panic(err)
	}
	return digest
}
func TestProjectionKeyC1MutationMatrix(t *testing.T) {
	baseSpec := projectionSpec()
	base, err := NewProjectionKey(baseSpec)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(*ProjectionKeySpec)
	}{
		{"domain", func(s *ProjectionKeySpec) { s.Domain = "payments" }},
		{"version", func(s *ProjectionKeySpec) { s.Version = "v3" }},
		{"artifact kind", func(s *ProjectionKeySpec) { s.ArtifactKind = "docs" }},
		{"projection", func(s *ProjectionKeySpec) { s.Projection = "ir" }},
		{"host stage", func(s *ProjectionKeySpec) { s.HostStage = GoooHostedStage }},
		{"semantic closure", func(s *ProjectionKeySpec) { s.SemanticClosureDigest = HashBytes([]byte("changed")) }},
		{"dependency root", func(s *ProjectionKeySpec) { s.DependencyRoot = HashBytes([]byte("changed")) }},
		{"policy schema", func(s *ProjectionKeySpec) { s.PolicySchemaDigest = HashBytes([]byte("changed")) }},
		{"toolchain", func(s *ProjectionKeySpec) { s.Toolchain = "go1.26.6" }},
		{"target", func(s *ProjectionKeySpec) { s.Target = "linux/amd64" }},
		{"build tags", func(s *ProjectionKeySpec) { s.BuildTags = []string{"linux", "race"} }},
		{"options", func(s *ProjectionKeySpec) { s.OptionsDigest = mustOptionsDigest(map[string]any{"mode": "safe"}) }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			mutated := baseSpec
			test.mutate(&mutated)
			key, err := NewProjectionKey(mutated)
			if err != nil {
				t.Fatal(err)
			}
			if key == base {
				t.Fatalf("mutation %q retained key %s", test.name, key)
			}
		})
	}
}
