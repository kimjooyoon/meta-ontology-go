package packageruntime

import (
	"fmt"
	"reflect"
	"slices"
)

// SplitSharedSource is for a manifest whose filenames address shared source
// occurrences, such as one external workspace definition. Every package that
// names the selected file consumes the same partition. Virtual package-local
// filenames retain the separate SplitSource operation.
func SplitSharedSource(manifest Manifest, packagePath, filename, beforeID, into string) (SourceSplit, error) {
	return splitSource(manifest, packagePath, filename, beforeID, into, true)
}

func verifySourcePartition(before Interface, result SourceSplit) (SourceSplit, error) {
	image, packages, err := build(result.Manifest, true)
	if err != nil {
		return SourceSplit{}, fmt.Errorf("split package: %w", err)
	}
	after := Interface{Entry: result.Manifest.Entry, Packages: packages, ImageDigest: image.Digest}
	if !reflect.DeepEqual(splitInterface(before), splitInterface(after)) {
		return SourceSplit{}, fmt.Errorf("split changed a declaration identity, type, field or entry")
	}
	result.OutputImageDigest = after.ImageDigest
	return result, nil
}

func shareSourcePartition(original Manifest, result SourceSplit) (SourceSplit, error) {
	consumers := []string{result.PackagePath}
	for i, pkg := range original.Packages {
		if pkg.Path == result.PackagePath {
			continue
		}
		for j, source := range pkg.Sources {
			if source.Filename != result.InputFilename {
				continue
			}
			if splitSourceDigest(source.Content) != result.InputSHA256 {
				return SourceSplit{}, fmt.Errorf("shared source %q has different bytes in package %q", source.Filename, pkg.Path)
			}
			sources := slices.Clone(pkg.Sources)
			sources[j].Content = result.Files[0].Source
			part := result.Files[1]
			result.Manifest.Packages[i].Sources = append(sources, Source{Filename: part.Filename, Content: part.Source})
			consumers = append(consumers, pkg.Path)
		}
	}
	if len(consumers) > 1 {
		slices.Sort(consumers)
		result.Schema = "gooo/package-source-split/v2"
		result.AffectedPackages = consumers
	}
	return result, nil
}
