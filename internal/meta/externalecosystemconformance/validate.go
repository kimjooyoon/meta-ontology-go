package externalecosystemconformance

import (
	"crypto/sha256"
	"encoding/hex"
)

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validateEvidence(e Evidence) (string, string, int) {
	if len(e.Readme) == 0 || len(e.GoMod) == 0 {
		return ResolutionUnknown, ReasonUnavailable, 0
	}
	readmeOK := len(e.Readme) == ExpectedReadmeBytes && digest(e.Readme) == ExpectedReadmeHash
	goModOK := len(e.GoMod) == ExpectedGoModBytes && digest(e.GoMod) == ExpectedGoModHash
	if !readmeOK || !goModOK {
		return ResolutionInvariant, ReasonDigestMismatch, 0
	}
	return "", "", 2
}

func validateCapsule(c Capsule) (string, string) {
	if c.CommitSHA != ExpectedCommit {
		return ResolutionInvariant, ReasonCommitMismatch
	}
	if c.LicenseSPDX != ExpectedLicense {
		return ResolutionInvariant, ReasonLicenseMismatch
	}
	staticOK := c.Schema == ReferenceSchema && c.ReferenceID == ExpectedReferenceID
	staticOK = staticOK && c.RepositoryURL == ExpectedRepository && c.TreeSHA == ExpectedTree
	staticOK = staticOK && c.ModulePath == ExpectedModule && c.ModuleGoVersion == ExpectedGoVersion
	if !staticOK || len(c.Documents) != 2 || len(c.Capabilities) != len(capabilityRules) {
		return ResolutionInvariant, ReasonCapsuleMismatch
	}
	expectedDocs := map[string]Document{
		"README": {"README", ExpectedRepository + "/blob/" + ExpectedCommit + "/README.md", ExpectedReadmeHash, ExpectedReadmeBytes},
		"GO_MOD": {"GO_MOD", ExpectedRepository + "/blob/" + ExpectedCommit + "/go.mod", ExpectedGoModHash, ExpectedGoModBytes},
	}
	seenDocs := map[string]bool{}
	for _, document := range c.Documents {
		expected, ok := expectedDocs[document.ID]
		if !ok || document != expected || seenDocs[document.ID] {
			return ResolutionInvariant, ReasonCapsuleMismatch
		}
		seenDocs[document.ID] = true
	}
	seenCapabilities := map[string]bool{}
	for _, capability := range c.Capabilities {
		rule, ok := capabilityRules[capability.ID]
		if !ok || !knownRelation(capability.Relation) {
			return ResolutionUnknown, ReasonRelationUnknown
		}
		valid := capability.Relation == rule.Relation && capability.MetaOperation == rule.MetaOperation
		if !valid || capability.Status != "REFERENCE_ONLY" || seenCapabilities[capability.ID] {
			return ResolutionInvariant, ReasonCapsuleMismatch
		}
		seenCapabilities[capability.ID] = true
	}
	return "", ""
}

func knownRelation(value string) bool {
	return value == "STRUCTURAL_HINT" || value == "DOCUMENTED_LIMITATION" ||
		value == "GUARDRAIL_CONTRAST"
}
