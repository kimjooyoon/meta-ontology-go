package packageruntime

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourcepartition"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type SourceSplitBoundary struct {
	BeforeID  string               `json:"before_id"`
	Name      string               `json:"name"`
	StartByte int                  `json:"start_byte"`
	StartLine int                  `json:"start_line"`
	Left      sourcepartition.Size `json:"left"`
	Right     sourcepartition.Size `json:"right"`
}

type SourceSplitBoundaries struct {
	PackagePath       string                `json:"package_path"`
	InputFilename     string                `json:"input_filename"`
	InputSHA256       string                `json:"input_sha256"`
	InputImageDigest  string                `json:"input_image_digest"`
	Input             sourcepartition.Size  `json:"input"`
	HeaderBytes       int                   `json:"duplicated_header_bytes"`
	SemanticAdmission string                `json:"semantic_admission"`
	Candidates        []SourceSplitBoundary `json:"candidates"`
}

// InspectSourceSplit validates the original workspace once, then describes its
// physical declaration cuts. It does not select a cut or validate hypothetical
// consumers, destinations, output limits or registry updates. Use SplitSharedSource
// (and host preparation) for the selected candidate before installing anything.
func InspectSourceSplit(manifest Manifest, packagePath, filename string) (SourceSplitBoundaries, error) {
	view, err := Describe(manifest)
	if err != nil {
		return SourceSplitBoundaries{}, err
	}
	packages := make(map[string]InterfacePackage, len(view.Packages))
	for _, pkg := range view.Packages {
		packages[pkg.Path] = pkg
	}
	for _, pkg := range manifest.Packages {
		if pkg.Path != packagePath {
			continue
		}
		for _, source := range pkg.Sources {
			if source.Filename == filename {
				return describeSplitBoundaries(source, packagePath, view.ImageDigest, packages[pkg.Path])
			}
		}
	}
	return SourceSplitBoundaries{}, fmt.Errorf("split source %q is absent from package %q", filename, packagePath)
}

func describeSplitBoundaries(source Source, packagePath, image string, pkg InterfacePackage) (SourceSplitBoundaries, error) {
	layout, err := sourcepartition.InspectBoundaries(source.Filename, source.Content)
	if err != nil {
		return SourceSplitBoundaries{}, err
	}
	ids := make(map[string]string)
	for _, declaration := range pkg.Declarations {
		if declaration.Source == source.Filename {
			ids[declaration.Name] = declaration.ID
		}
	}
	result := SourceSplitBoundaries{PackagePath: packagePath, InputFilename: source.Filename,
		InputSHA256: splitSourceDigest(source.Content), InputImageDigest: image, Input: layout.Input,
		HeaderBytes: layout.HeaderBytes, SemanticAdmission: "UNASSESSED", Candidates: []SourceSplitBoundary{}}
	for _, boundary := range layout.Boundaries {
		name, span := splitDeclarationName(boundary.Declaration), boundary.Declaration.SourceSpan()
		if ids[name] == "" {
			return SourceSplitBoundaries{}, fmt.Errorf("split boundary has no resolved declaration identity")
		}
		result.Candidates = append(result.Candidates, SourceSplitBoundary{BeforeID: ids[name], Name: name,
			StartByte: span.Start.Offset, StartLine: span.Start.Line, Left: boundary.Left, Right: boundary.Right})
	}
	return result, nil
}

func splitDeclarationName(declaration syntax.Declaration) string {
	switch value := declaration.(type) {
	case *syntax.EntityDecl:
		return value.Name
	case *syntax.ActivityDecl:
		return value.Name
	}
	return ""
}
