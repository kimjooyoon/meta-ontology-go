package packageruntime

import (
	"crypto/sha256"
	"fmt"
	"path"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcepartition"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// SourceSlice and SplitSourceFile retain the source partition wire contract.
type SourceSlice = sourcepartition.Slice
type SplitSourceFile = sourcepartition.File

type SourceSplit struct {
	Schema            string            `json:"schema"`
	PackagePath       string            `json:"package_path"`
	AffectedPackages  []string          `json:"affected_packages,omitempty"`
	InputFilename     string            `json:"input_filename"`
	InputSHA256       string            `json:"input_sha256"`
	BeforeID          string            `json:"before_id"`
	InputImageDigest  string            `json:"input_image_digest"`
	OutputImageDigest string            `json:"output_image_digest"`
	Files             []SplitSourceFile `json:"files"`
	Manifest          Manifest          `json:"manifest,omitzero"`
}

// SplitSource proposes a package-aware source partition, without writing files.
// The caller must own the supplied workspace manifest and install its updated
// source list with both outputs. Single-file or embedded consumers need their
// own migration; this function does not discover or modify those consumers.
func SplitSource(manifest Manifest, packagePath, filename, beforeID, into string) (SourceSplit, error) {
	return splitSource(manifest, packagePath, filename, beforeID, into, false)
}

func splitSource(manifest Manifest, packagePath, filename, beforeID, into string, shared bool) (SourceSplit, error) {
	beforeImage, beforePackages, err := build(manifest, true)
	if err != nil {
		return SourceSplit{}, err
	}
	before := Interface{Entry: manifest.Entry, Packages: beforePackages, ImageDigest: beforeImage.Digest}
	if !splitFilename(into) {
		return SourceSplit{}, fmt.Errorf("split destination must be a canonical relative .gooo filename")
	}
	for _, pkg := range manifest.Packages {
		for _, source := range pkg.Sources {
			if strings.EqualFold(path.Clean(source.Filename), into) {
				return SourceSplit{}, fmt.Errorf("split destination %q already belongs to the workspace", into)
			}
		}
	}
	result, err := partitionSource(manifest, before, packagePath, filename, beforeID, into)
	if err != nil {
		return SourceSplit{}, err
	}
	if shared {
		result, err = shareSourcePartition(manifest, result)
		if err != nil {
			return SourceSplit{}, err
		}
	}
	return verifySourcePartition(before, result)
}

func splitFilename(value string) bool {
	return value != "" && path.Clean(value) == value && !path.IsAbs(value) &&
		!strings.ContainsAny(value, "\\\x00:") && !strings.HasPrefix(value, "../") && path.Ext(value) == ".gooo"
}

func partitionSource(manifest Manifest, before Interface, packagePath, filename, beforeID, into string) (SourceSplit, error) {
	updated := manifest
	updated.Packages = append([]PackageSpec(nil), manifest.Packages...)
	for i, pkg := range manifest.Packages {
		if pkg.Path != packagePath {
			continue
		}
		for j, source := range pkg.Sources {
			if source.Filename != filename {
				continue
			}
			files, err := splitSourceBytes(source, before, packagePath, beforeID, into)
			if err != nil {
				return SourceSplit{}, err
			}
			updated.Packages[i].Sources = append([]Source(nil), pkg.Sources...)
			updated.Packages[i].Sources[j].Content = files[0].Source
			updated.Packages[i].Sources = append(updated.Packages[i].Sources, Source{Filename: into, Content: files[1].Source})
			return SourceSplit{Schema: "gooo/package-source-split/v1", PackagePath: packagePath,
				InputFilename: filename, InputSHA256: splitSourceDigest(source.Content), BeforeID: beforeID,
				InputImageDigest: before.ImageDigest, Files: files, Manifest: updated}, nil
		}
	}
	return SourceSplit{}, fmt.Errorf("split source %q is absent from package %q", filename, packagePath)
}

func splitSourceDigest(source string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(source)))
}

func splitInterface(view Interface) Interface {
	view.ImageDigest, view.Digest = "", ""
	for i := range view.Packages {
		view.Packages[i].Sources = nil
		for j := range view.Packages[i].Declarations {
			view.Packages[i].Declarations[j].Source = ""
		}
	}
	return view
}

func splitSourceBytes(source Source, view Interface, packagePath, beforeID, into string) ([]SplitSourceFile, error) {
	name := ""
	for _, pkg := range view.Packages {
		for _, declaration := range pkg.Declarations {
			if pkg.Path == packagePath && declaration.Source == source.Filename && declaration.ID == beforeID {
				name = declaration.Name
			}
		}
	}
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(source.Filename, source.Content, syntax.EntityFieldsV4Support())
	if file == nil || diagnostics.HasErrors() || name == "" {
		return nil, fmt.Errorf("split needs a resolved declaration ID and valid source")
	}
	for _, declaration := range file.Decls {
		declName := ""
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			declName = value.Name
		case *syntax.ActivityDecl:
			declName = value.Name
		}
		if declName == name {
			return sourcepartition.SplitAt(source.Filename, source.Content, into, declaration.SourceSpan().Start.Offset)
		}
	}
	return nil, fmt.Errorf("resolved split declaration is absent from source")
}
