package extractor

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/moduleboundary"

// sourceModuleRoot chooses the nearest go.mod without changing repository-relative
// source and generated-file identities.
func sourceModuleRoot(root, logical string) (string, error) {
	path, err := safePath(root, logical)
	if err != nil {
		return "", err
	}
	return moduleboundary.Source(root, path)
}
