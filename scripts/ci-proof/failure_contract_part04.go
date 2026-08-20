package main

import "slices"

func isCanonicalFailureJob(name string) bool {
	return slices.Contains([]string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}, name)
}
func containsCode(codes []string, target string) bool {
	return slices.Contains(codes, target)
}
func sameArtifactInputs(left, right []artifactInput) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
