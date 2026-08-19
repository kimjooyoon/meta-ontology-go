package main

func isCanonicalFailureJob(name string) bool {
	for _, canonical := range []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"} {
		if name == canonical {
			return true
		}
	}
	return false
}
func containsCode(codes []string, target string) bool {
	for _, code := range codes {
		if code == target {
			return true
		}
	}
	return false
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
