package main

func (cfg config) artifactPaths() map[string]string {
	return map[string]string{
		"language-syntax-roundtrip": cfg.syntax,
		"language-semantic-model": cfg.semantic,
		"language-deterministic-query": cfg.query,
		"language-go-interoperation": cfg.interop,
		"language-diagnostic-provenance": cfg.diagnostic,
		"language-package-runtime": cfg.runtime,
		"toolchain-cli": cfg.cli,
		"toolchain-format-fix": cfg.formatFix,
		"toolchain-executable-use-cases": cfg.useCases,
	}
}

func readArtifacts(paths map[string]string) (map[string][]byte, error) {
	artifacts := make(map[string][]byte, len(paths))
	for id, path := range paths {
		raw, err := readFile(path)
		if err != nil {
			return nil, err
		}
		artifacts[id] = raw
	}
	return artifacts, nil
}
