package verticalsliceclosureshadow

import "strings"

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
		"language-syntax-roundtrip":      {"gooo/language-syntax-roundtrip/v1", 20},
		"language-semantic-model":        {"gooo/language-semantic-model/v1", 22},
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
