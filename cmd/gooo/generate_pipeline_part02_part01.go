package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"path/filepath"
	"time"
)

func buildGenerateArtifacts(options generateOptions, input generateInput, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) (generateArtifacts, int) {
	generation, err := generateWithDeadline(input.file, input.previousGo, remainingDeadline(deadline))
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.filename, "generator.generate", "generation failed", err, input.file)
	}
	output := filepath.Join(options.outputDir, generatedFileName)
	manifestPath := options.manifestPath
	if manifestPath == "" {
		manifestPath = filepath.Join(options.outputDir, generatedManifestFileName)
	}
	cleanupDirs := append(missingDirectoryChain(options.outputDir), missingDirectoryChain(filepath.Dir(manifestPath))...)
	cleanupDirs = uniquePaths(cleanupDirs)
	buildSucceeded := false
	defer func() {
		if !buildSucceeded {
			_ = cleanupGenerateDirectories(cleanupDirs)
		}
	}()
	manifest, err := buildProjectionManifest(options.filename, output, input.source, input.previousGo, generation.ir, generation.result)
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.filename, "manifest.build", "manifest failed", err, input.file)
	}
	root, err := canonicalOutputRoot(options.outputDir)
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.outputDir, "io.mkdir", "output root", err, input.file)
	}
	output, err = resolveOutputPath(root, generatedFileName)
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.outputDir, "io.output-path", "output path", err, input.file)
	}
	manifestPath, err = resolveManifestPath(root, manifestPath)
	if err != nil {
		return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.outputDir, "io.manifest-path", "manifest path", err, input.file)
	}
	manifest.GeneratedFile = output
	runtimePlanPath := ""
	var runtimePlanData []byte
	if len(generation.ir.RuntimeBindings) > 0 {
		if options.runtimePlanFilename == "" {
			return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.filename, "runtime-plan.required", "runtime plan required for runtime bindings", errRuntimePlanRequired, input.file)
		}
		runtimePlanPath, err = resolveOutputPath(root, options.runtimePlanFilename)
		if err != nil {
			return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.runtimePlanFilename, "io.runtime-plan-path", "runtime plan path", err, input.file)
		}
		document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(input.file, syntax.EntityFieldsV1Support())
		if err != nil {
			return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.filename, "runtime-plan.lower", "runtime plan lowering", err, input.file)
		}
		var typedPlan bidir.TypedPlan
		if len(document.BindingEdges) > 0 {
			typedPlan, err = bidir.CompileTypedPlan(document)
			if err != nil {
				return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.filename, "runtime-plan.validate", "runtime plan validation", err, input.file)
			}
		}
		runtimePlanData, err = buildRuntimePlanDataWithTypedPlan(input.source, generation.ir, typedPlan)
		if err != nil {
			return generateArtifacts{}, reportGenerateError(jsonMode, stdout, stderr, options.filename, "runtime-plan.build", "runtime plan failed", err, input.file)
		}
		cleanupDirs = append(cleanupDirs, missingDirectoryChain(filepath.Dir(runtimePlanPath))...)
		cleanupDirs = uniquePaths(cleanupDirs)
	}
	buildSucceeded = true
	return generateArtifacts{ir: generation.ir, result: generation.result, output: output, manifestPath: manifestPath, manifest: manifest, runtimePlanPath: runtimePlanPath, runtimePlanData: runtimePlanData, cleanupDirs: cleanupDirs}, exitOK
}
func reportGenerateError(jsonMode bool, stdout, stderr io.Writer, filename, code, prefix string, err error, file *syntax.File) int {
	if jsonMode {
		return reportFailure(true, stdout, stderr, "generate", filename, code, err.Error(), syntaxFileSpan(file))
	}
	fmt.Fprintf(stderr, "gooo: %s: %s: %v\n", filename, prefix, err)
	return exitFailure
}
func writeGenerateArtifacts(artifacts generateArtifacts, jsonMode bool, stdout, stderr io.Writer) int {
	manifestBytes, err := jsonManifestBytes(artifacts.manifest)
	if err != nil {
		_ = cleanupGenerateDirectories(artifacts.cleanupDirs)
		return reportGenerateIOError(jsonMode, stdout, stderr, artifacts.manifestPath, "io.write-manifest", "write manifest", err)
	}
	writes := []atomicWrite{
		{path: artifacts.output, data: artifacts.result.Source},
		{path: artifacts.manifestPath, data: manifestBytes},
	}
	if artifacts.runtimePlanPath != "" {
		writes = append(writes, atomicWrite{path: artifacts.runtimePlanPath, data: artifacts.runtimePlanData})
	}
	if err := writeAtomicFiles(writes); err != nil {
		_ = cleanupGenerateDirectories(artifacts.cleanupDirs)
		return reportGenerateIOError(jsonMode, stdout, stderr, artifacts.output, "io.write", "write generated source and manifest", err)
	}
	return exitOK
}
