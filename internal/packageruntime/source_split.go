package packageruntime

import (
	"crypto/sha256"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// SourceSlice records exact bytes copied from the original source. Header
// duplication is explicit; declaration bodies and comments are never formatted.
type SourceSlice struct {
	InputStart  int `json:"input_start_byte"`
	InputEnd    int `json:"input_end_byte"`
	OutputStart int `json:"output_start_byte"`
	OutputEnd   int `json:"output_end_byte"`
}

type SplitSourceFile struct {
	Filename string        `json:"filename"`
	SHA256   string        `json:"sha256"`
	Source   string        `json:"source"`
	Origins  []SourceSlice `json:"origins"`
}

type SourceSplit struct {
	Schema            string            `json:"schema"`
	PackagePath       string            `json:"package_path"`
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
	afterImage, afterPackages, err := build(result.Manifest, true)
	if err != nil {
		return SourceSplit{}, fmt.Errorf("split package: %w", err)
	}
	after := Interface{Entry: result.Manifest.Entry, Packages: afterPackages, ImageDigest: afterImage.Digest}
	if !reflect.DeepEqual(splitInterface(before), splitInterface(after)) {
		return SourceSplit{}, fmt.Errorf("split changed a declaration identity, type, field or entry")
	}
	result.OutputImageDigest = after.ImageDigest
	return result, nil
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
	if file == nil || diagnostics.HasErrors() || len(file.Decls) < 2 || name == "" {
		return nil, fmt.Errorf("split needs a resolved declaration ID and at least two declarations")
	}
	header, cut := file.Decls[0].SourceSpan().Start.Offset, -1
	for i, declaration := range file.Decls {
		declName := ""
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			declName = value.Name
		case *syntax.ActivityDecl:
			declName = value.Name
		}
		if i > 0 && declName == name {
			cut = declaration.SourceSpan().Start.Offset
		}
	}
	if err := validateSplitHeader(file, header, cut, len(source.Content)); err != nil {
		return nil, err
	}
	left, right := source.Content[:cut], source.Content[:header]+source.Content[cut:]
	return []SplitSourceFile{
		{Filename: source.Filename, SHA256: splitSourceDigest(left), Source: left,
			Origins: []SourceSlice{{0, cut, 0, cut}}},
		{Filename: into, SHA256: splitSourceDigest(right), Source: right,
			Origins: []SourceSlice{{0, header, 0, header}, {cut, len(source.Content), header, len(right)}}},
	}, nil
}

func validateSplitHeader(file *syntax.File, header, cut, length int) error {
	if header <= 0 || cut <= header || cut >= length {
		return fmt.Errorf("split boundary must precede a non-first declaration")
	}
	if file.Package == nil || file.Namespace == nil || file.Package.Span.End.Offset > header || file.Namespace.Span.End.Offset > header {
		return fmt.Errorf("split requires a leading package and namespace header")
	}
	for _, imp := range file.Imports {
		if imp.Span.End.Offset > header {
			return fmt.Errorf("split requires imports before declarations to retain file-local aliases")
		}
	}
	for _, binding := range file.Bindings {
		if binding.Span.Start.Offset < header {
			return fmt.Errorf("split cannot duplicate a binding in the source header")
		}
	}
	return nil
}
