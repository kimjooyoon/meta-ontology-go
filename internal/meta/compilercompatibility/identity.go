package compilercompatibility

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

var implementationManifestPaths = append(generation.SemanticRetentionCompilerManifestPaths(),
	"cmd/gooo/generate_pipeline_part03.go",
	"cmd/gooo/public_continuity_generate_part01.go",
	"cmd/gooo/public_continuity_generate_part02.go",
	"cmd/gooo/public_continuity_generate_part03.go",
	"internal/meta/compilercompatibility/model.go",
	"internal/meta/compilercompatibility/identity.go",
	"internal/meta/compilercompatibility/certificate.go",
	"internal/meta/compilercompatibility/evaluate.go",
	"internal/meta/compatibilitypolicy/policy.go",
	"internal/meta/compatibilitypolicy/generated/evaluator.go",
	"cmd/gooo/compatibility_generate_part01.go",
	"cmd/gooo/compatibility_generate_part02.go")

func ImplementationManifestPaths() []string {
	return append([]string(nil), implementationManifestPaths...)
}

func CompilerImplementationDigest(readFile func(string) ([]byte, error)) (string, error) {
	return generation.SemanticRetentionManifestDigest(readFile, implementationManifestPaths)
}
