package main

import (
	"fmt"
	"strings"
)

func computeProofDigests(root, generated string, evidence evidenceInput) (proofDigests, error) {
	source, err := hashGitFiles(root)
	if err != nil {
		return proofDigests{}, err
	}
	semantic, err := hashGitFiles(root, "internal/semantic", "internal/bidir")
	if err != nil {
		return proofDigests{}, err
	}
	provenance, err := hashGitFiles(root, "internal/provenance", "internal/semantic/prov.go")
	if err != nil {
		return proofDigests{}, err
	}
	projection, err := hashDirectory(generated)
	if err != nil {
		return proofDigests{}, err
	}
	build, err := hashGitFiles(root, "go.mod", "go.sum")
	if err != nil {
		return proofDigests{}, err
	}
	policy, err := hashGitFiles(root, ".github", "scripts", "internal/verify")
	if err != nil {
		return proofDigests{}, err
	}
	schema, err := hashGitFiles(root, ".github/ci-governance.json")
	if err != nil {
		return proofDigests{}, err
	}
	toolchain, err := toolchainIdentity()
	if err != nil {
		return proofDigests{}, err
	}
	toolchainDigest := digestBytes([]byte(toolchain))
	if mismatches := digestMismatchFields(source, semantic, projection, policy, toolchainDigest, evidence); len(mismatches) > 0 {
		return proofDigests{}, fmt.Errorf("independently recomputed CI evidence digest mismatch: %s", strings.Join(mismatches, ", "))
	}
	target := digestBytes([]byte(evidence.BaseRef + "\x00" + evidence.BaseSHA))
	return proofDigests{Source: source, Semantic: semantic, Provenance: provenance, Projection: projection, Build: build, Policy: policy, Schema: schema, Toolchain: toolchainDigest, Target: target}, nil
}
func digestMismatchFields(source, semantic, projection, policy, toolchain string, evidence evidenceInput) []string {
	checks := []struct {
		name, actual, expected string
	}{
		{"source", source, evidence.Digests.Source},
		{"semantic", semantic, evidence.Digests.IR},
		{"projection", projection, evidence.Digests.Generated},
		{"policy", policy, evidence.Digests.Policy},
		{"toolchain", toolchain, evidence.Digests.Toolchain},
	}
	mismatches := make([]string, 0, len(checks))
	for _, check := range checks {
		if check.actual != check.expected {
			mismatches = append(mismatches, check.name)
		}
	}
	return mismatches
}
