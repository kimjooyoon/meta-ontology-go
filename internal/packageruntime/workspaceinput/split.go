package workspaceinput

import (
	"fmt"
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

type SplitRequest struct {
	PackagePath, Source, BeforeID, Into string
}

// SplitProposal couples the exact source partition with its authored file list.
// Definition retains authored names/imports; Result carries the normalized image.
// Preparation reads only. The caller owns installation and current-input checks.
type SplitProposal struct {
	Definition Definition
	Result     packageruntime.SourceSplit
}

func PrepareSplit(reader Reader, manifestPath string, definition Definition, request SplitRequest) (SplitProposal, error) {
	loaded, err := Load(reader, manifestPath, definition)
	if err != nil {
		return SplitProposal{}, err
	}
	result, err := packageruntime.SplitSource(loaded, request.PackagePath, request.Source, request.BeforeID, request.Into)
	if err != nil {
		return SplitProposal{}, err
	}
	if err := checkSplitLimits(result.Manifest); err != nil {
		return SplitProposal{}, err
	}
	after := definition
	after.Packages = slices.Clone(definition.Packages)
	for i := range after.Packages {
		pkg := &after.Packages[i]
		pkg.Imports = slices.Clone(pkg.Imports)
		pkg.Sources = slices.Clone(pkg.Sources)
		if pkg.Path == request.PackagePath {
			pkg.Sources = append(pkg.Sources, request.Into)
		}
	}
	return SplitProposal{Definition: after, Result: result}, nil
}

func checkSplitLimits(manifest packageruntime.Manifest) error {
	count, size := 0, 0
	for _, pkg := range manifest.Packages {
		for _, source := range pkg.Sources {
			count++
			size += len(source.Content)
			if count > MaxSourceCount || len(source.Content) > MaxSourceBytes || size > MaxSourceSetBytes {
				return fmt.Errorf("split output exceeds the workspace source count or byte limit")
			}
		}
	}
	return nil
}
