package main

import (
	metatoolchain "github.com/kimjooyoon/meta-ontology-go/internal/meta/toolchain"
)

func computeDigests(root, generated string) (digests, error) {
	source, err := hashGitFiles(root)
	if err != nil {
		return digests{}, err
	}
	ir, err := hashGitFiles(root, "internal/semantic", "internal/bidir")
	if err != nil {
		return digests{}, err
	}
	fixtures, err := hashGitFiles(root, "internal/generator", "examples")
	if err != nil {
		return digests{}, err
	}
	policy, err := hashGitFiles(root, ".github", "scripts", "internal/verify")
	if err != nil {
		return digests{}, err
	}
	sourceMap, err := hashGeneratedSourceMap(generated)
	if err != nil {
		return digests{}, err
	}
	generatedDigest, err := hashDirectory(generated)
	if err != nil {
		return digests{}, err
	}
	toolchain, err := toolchainIdentity()
	if err != nil {
		return digests{}, err
	}
	return digests{SourceSHA256: source, IRSHA256: ir, GeneratorFixtureSHA256: fixtures, GeneratedOutputSHA256: generatedDigest, SourceMapSHA256: sourceMap, PolicySHA256: policy, ToolchainSHA256: digestBytes([]byte(toolchain))}, nil
}
func toolchainIdentity() (string, error) {
	return metatoolchain.Identity(".")
}
