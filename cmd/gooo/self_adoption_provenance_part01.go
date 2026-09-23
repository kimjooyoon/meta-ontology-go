package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func semanticAdoptionProvenance(sourcePath, sourceDigest, contractPath, contractDigest string) *generation.SemanticAdoptionProvenance {
	profile := syntax.CurrentEntityFieldsSupport().Profile
	profileDigest := cache.HashBytes([]byte(fmt.Sprintf("%s|%d|%s", profile.ID, profile.Version, profile.Digest))).String()
	return &generation.SemanticAdoptionProvenance{
		SourcePath: sourcePath, ContractPath: contractPath,
		SourceDigest: sourceDigest, ProfileDigest: profileDigest,
		ToolchainDigest: generation.SemanticRetentionToolchainDigest(), ContractDigest: contractDigest,
	}
}
