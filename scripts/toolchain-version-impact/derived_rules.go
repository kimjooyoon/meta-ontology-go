package main

import (
	"sort"
	"strings"
)

const (
	compatibilitySourcePath = "examples/self-improvement-observation/observation.gooo"
	compatibilityTargetPath = "internal/meta/compatibilitypolicy/generated/evaluator.go"
	publicTrustSourcePath   = "examples/public-trust-surface/main.gooo"
	publicTrustTargetPath   = "internal/meta/publictrust/generated/evaluator.go"
	readmeTargetPath        = "README.md"
)

func derivedOutputRules(occurrences []occurrence, targetHashes map[string]string) []derivedOutputRule {
	sources := map[string]string{}
	for _, site := range occurrences {
		if site.Classified && (site.Path == compatibilitySourcePath || site.Path == publicTrustSourcePath) {
			sources[site.Path] = site.SHA256
		}
	}
	var rules []derivedOutputRule
	add := func(source, target, id string) {
		if sha := sources[source]; sha != "" {
			rules = append(rules, derivedOutputRule{SourcePath: source, SourceSHA256: sha,
				TargetPath: target, TargetSHA256: targetHashes[target], RuleID: id, RuleVersion: 1})
		}
	}
	add(compatibilitySourcePath, compatibilityTargetPath, "urn:gooo:toolchain-projection:compatibility-evaluator")
	add(publicTrustSourcePath, publicTrustTargetPath, "urn:gooo:toolchain-projection:public-trust-evaluator")
	add(publicTrustSourcePath, readmeTargetPath, "urn:gooo:toolchain-projection:public-trust-badges")
	sort.Slice(rules, func(i, j int) bool { return rules[i].TargetPath < rules[j].TargetPath })
	return rules
}

func derivedSourceOrigins(rules []derivedOutputRule) ([]any, []any) {
	sources := map[string]string{}
	for _, rule := range rules {
		sources[rule.SourcePath] = rule.SourceSHA256
	}
	paths := make([]string, 0, len(sources))
	for path := range sources {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	nodes, references := make([]any, 0, len(paths)), make([]any, 0, len(paths))
	for _, path := range paths {
		id := "urn:gooo:toolchain-projection:source:" + strings.TrimPrefix(digest([]byte(path)), "sha256:") +
			":" + strings.TrimPrefix(sources[path], "sha256:")
		nodes = append(nodes, map[string]any{"@id": id, "@type": "prov:Entity", "gooo:path": path,
			"gooo:content_sha256": sources[path]})
		references = append(references, map[string]string{"@id": id})
	}
	return nodes, references
}
