package main

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodylibrary"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

func generatePackageLibrary(program workspaceexecution.Program, manifestDigest string, options packageGenerateOptions) (string, error) {
	var header strings.Builder
	info, _ := debug.ReadBuildInfo()
	compiler := versionBuildMetadata(info, runtime.Version())
	generation := bodylibrary.GenerationOptions{StageExports: options.exports}
	if options.embedOrigin {
		generation.Origin = &bodylibrary.GenerationContext{
			CompilerVersion: compiler.Version, CompilerSourceSHA: compiler.CompilerSourceSHA,
			CompilerSourceStatus: compiler.SourceStatus, GoVersion: compiler.GoVersion,
			WorkspaceManifestSHA: manifestDigest, WorkspaceImageSHA: program.Workspace.Digest,
		}
	}
	fmt.Fprintf(&header, "// Compiler version=%q source=%q source_status=%q go=%q\n",
		compiler.Version, compiler.CompilerSourceSHA, compiler.SourceStatus, compiler.GoVersion)
	fmt.Fprintf(&header, "// Workspace manifest: %s\n// Workspace image: %s\n", manifestDigest, program.Workspace.Digest)
	for _, pkg := range program.Workspace.Packages {
		for _, source := range pkg.Sources {
			fmt.Fprintf(&header, "// Gooo source package=%q file=%q source_json_string_sha256=%q\n",
				pkg.Path, source.Filename, source.SourceDigest)
			if generation.Origin != nil {
				generation.Origin.Sources = append(generation.Origin.Sources,
					bodylibrary.GenerationSource{PackagePath: pkg.Path, Filename: source.Filename, SourceSHA: source.SourceDigest})
			}
		}
	}
	source, err := bodylibrary.GenerateWithOptions(context.Background(), "workspace.gooo", []byte(program.Source),
		program.Entry.LoweredName, options.packageName, generation)
	if err != nil {
		return "", err
	}
	header.WriteString(source)
	return header.String(), nil
}
