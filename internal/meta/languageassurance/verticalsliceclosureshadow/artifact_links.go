package verticalsliceclosureshadow

import "strings"

func applyBoundaryLinks(result *BoundaryResult, artifact artifactEnvelope,
	artifacts map[string]artifactEnvelope, raw map[string][]byte,
	results map[string]*BoundaryResult, head string) {
	if result.Status != StatusSatisfied {
		return
	}
	switch result.ID {
	case "syntax":
		result.LinksSatisfied = 1
	case "semantics":
		if dependencyReady(result, results["syntax"]) {
			result.LinksSatisfied = semanticLinks(artifact, artifacts["syntax"], raw["syntax"])
		}
	case "binding":
		if dependencyReady(result, results["semantics"]) {
			result.LinksSatisfied = bindingLinks(artifact, artifacts["semantics"], raw["semantics"])
		}
	case "use-cases":
		if dependencyReady(result, results["syntax"]) &&
			artifact.Source.ConceptArtifactDigest ==
				artifacts["syntax"].Source.ConceptArtifactDigest {
			result.LinksSatisfied = 1
		}
	case "toolchain":
		if dependencyReady(result, results["syntax"]) &&
			dependencyReady(result, results["semantics"]) &&
			dependencyReady(result, results["use-cases"]) {
			result.LinksSatisfied = toolchainLinks(artifact.Surfaces, head)
		}
	case "release":
		result.LinksSatisfied = releaseLinks(artifact)
	}
	if result.Status == StatusSatisfied && result.LinksSatisfied != result.LinksTotal {
		setBlocked(result, boundaryReasonLink)
	}
}

func dependencyReady(target, dependency *BoundaryResult) bool {
	if dependency != nil && dependency.Status == StatusSatisfied {
		return true
	}
	if dependency == nil || dependency.Status == StatusUnknown {
		setUnknown(target, boundaryReasonDependency, false)
	} else {
		setBlocked(target, boundaryReasonLink)
	}
	return false
}

func semanticLinks(semantic, syntax artifactEnvelope, syntaxRaw []byte) int {
	links := 0
	if semantic.Source.SyntaxArtifactDigest == digestBytes(syntaxRaw) {
		links++
	}
	if semantic.Source.SyntaxReportDigest == syntax.ReportDigest {
		links++
	}
	return links
}

func bindingLinks(binding, semantic artifactEnvelope, semanticRaw []byte) int {
	links := 0
	if binding.Source.SemanticFileDigest == digestBytes(semanticRaw) {
		links++
	}
	if binding.Source.SemanticReportDigest == semantic.ReportDigest {
		links++
	}
	return links
}

func toolchainLinks(surfaces []artifactSurface, head string) int {
	expected := map[string]struct {
		schema string
		cases  int
	}{
		"language-syntax-roundtrip":       {"gooo/language-syntax-roundtrip/v1", 17},
		"language-semantic-model":         {"gooo/language-semantic-model/v1", 20},
		"toolchain-executable-use-cases": {"gooo/toolchain-executable-use-cases/v1", 3},
	}
	links := 0
	for _, surface := range surfaces {
		want, ok := expected[surface.ID]
		if ok && surface.Schema == want.schema && surface.Cases == want.cases &&
			surface.Status == StatusSatisfied && surface.HeadSHA == head {
			links++
		}
	}
	return links
}

func releaseLinks(artifact artifactEnvelope) int {
	if artifact.Summary.ToolchainBindings != 3 {
		return 0
	}
	links := 0
	for _, item := range artifact.Cases {
		if strings.HasSuffix(item.ID, "-go127-toolchain") &&
			item.Observed == "go1.27.0" && item.Expected == "go1.27.0" {
			links++
		}
	}
	return links
}
