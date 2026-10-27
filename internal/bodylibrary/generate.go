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
	return GenerateWithStageExports(ctx, filename, source, entry, packageName, nil)
}

// GenerateWithStageExports gives selected pure stages public names. Evaluate
// still composes the complete entry graph; a stage receives its own body inputs.
func GenerateWithStageExports(ctx context.Context, filename string, source []byte, entry, packageName string, exports []StageExport) (string, error) {
	return GenerateWithOptions(ctx, filename, source, entry, packageName, GenerationOptions{StageExports: exports})
}

// GenerateWithOptions can retain the generation inputs as an immutable constant.
// Reading that optional view does not evaluate a Gooo activity.
func GenerateWithOptions(ctx context.Context, filename string, source []byte, entry, packageName string, options GenerationOptions) (string, error) {
	if !token.IsIdentifier(packageName) || packageName == "_" || packageName == "main" {
		return "", fmt.Errorf("library generation requires an importable Go package name")
	}
	projection, err := bodyexecution.GenerateComputedProjection(ctx, filename, source, entry)
	if err != nil {
		return "", err
	}
	names, origins, err := resolveStageExports(projection.Plan, options.StageExports)
	if err != nil {
		return "", err
	}
	wrapper, err := libraryEntry(projection.Plan, projection.Source)
	if err != nil {
		return "", err
	}
	generated, err := nameLibraryProjection(projection.Source+wrapper, packageName, projection.SourceSHA256, projection.Plan, names)
	if err != nil {
		return "", err
	}
	if options.Origin != nil {
		generated, err = appendGenerationOrigin(generated, projection.SourceSHA256, projection.Plan, options)
		if err != nil {
			return "", err
		}
	}
	if len(origins)+len(generated) > maxLibraryProjectionBytes {
		return "", fmt.Errorf("library projection exceeds 256 KiB")
	}
	return origins + generated, nil
}
