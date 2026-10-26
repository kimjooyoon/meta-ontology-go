// Package bodylibrary exposes a computed Gooo graph as an importable Go library.
// Naming and host API presentation are separate from execution observations.
package bodylibrary

import (
	"context"
	"fmt"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func Generate(ctx context.Context, filename string, source []byte, entry, packageName string) (string, error) {
	if !token.IsIdentifier(packageName) || packageName == "_" || packageName == "main" {
		return "", fmt.Errorf("library generation requires an importable Go package name")
	}
	projection, err := bodyexecution.GenerateComputedProjection(ctx, filename, source, entry)
	if err != nil {
		return "", err
	}
	wrapper, err := libraryEntry(projection.Plan, projection.Source)
	if err != nil {
		return "", err
	}
	return nameLibraryProjection(projection.Source+wrapper, packageName, projection.SourceSHA256, projection.Plan)
}
