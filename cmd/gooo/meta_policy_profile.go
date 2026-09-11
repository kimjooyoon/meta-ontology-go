package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func runMetaPolicyGenerationProfile(options generateOptions, input generateInput, jsonMode bool, stdout, stderr interfaceWriter) int {
	projectRoot, err := profileCanonicalDirectory(options.profileProjectRoot)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	sourcePath, err := profileCanonicalInput(options.filename)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	if !profilePathWithin(projectRoot, sourcePath) {
		return reportMetaPolicyProfileError(stderr, fmt.Errorf("input source is outside the declared project boundary"))
	}
	outputCandidate, err := profileCanonicalCandidate(options.outputDir)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	if profilePathsOverlap(projectRoot, outputCandidate) {
		return reportMetaPolicyProfileError(stderr, errors.New("profile output must be outside the declared project boundary"))
	}
	outputRoot, err := canonicalOutputRoot(options.outputDir)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	if profilePathsOverlap(projectRoot, outputRoot) {
		return reportMetaPolicyProfileError(stderr, errors.New("profile output must be outside the declared project boundary"))
	}
	if err := ensurePublicOutputDirectory(outputRoot); err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}

	policy, err := policycompilation.CompileForIdentity(options.filename, input.source, options.profilePackage, options.profileNamespace)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, fmt.Errorf("compile source policy: %w", err))
	}
	judge := policycompilation.GenerateJudge(policy)
	policyBytes, err := profileJSONBytes(policy)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, fmt.Errorf("encode policy artifact: %w", err))
	}
	artifact := policycompilation.PolicyArtifact{Schema: policycompilation.ArtifactSchema, Policy: policy, GeneratedJudgeHash: policycompilation.DigestBytes(judge)}
	artifactBytes, err := profileJSONBytes(artifact)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, fmt.Errorf("encode compiled artifact: %w", err))
	}
	manifest := policycompilation.PublicGenerationManifest{
		Schema:               policycompilation.PublicGenerationManifestSchema,
		Profile:              policycompilation.PublicProfileID,
		SourceFile:           profileRelativePath(projectRoot, sourcePath),
		SourceDigest:         policy.SourceDigest,
		SemanticDigest:       policy.SemanticDigest,
		Package:              policy.Package,
		Namespace:            policy.Namespace,
		PolicyBytesDigest:    policycompilation.DigestBytes(policyBytes),
		ArtifactBytesDigest:  policycompilation.DigestBytes(artifactBytes),
		GeneratedJudgeDigest: policycompilation.DigestBytes(judge),
		GeneratedFiles:       []string{"policy.json", "artifact.json", "judge.go", "generation-manifest.json"},
		OutputRootClass:      policycompilation.PublicGenerationOutputRootClass,
		ExecutionObserved:    false,
		CurrentConformance:   policycompilation.PublicGenerationConformanceUnknown,
		RepositoryWrites:     0,
		MutationAuthority:    0,
		PromotionAuthority:   0,
	}
	manifestBytes, err := profileJSONBytes(manifest)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, fmt.Errorf("encode generation manifest: %w", err))
	}
	policyPath, err := resolveOutputPath(outputRoot, "policy.json")
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	artifactPath, err := resolveOutputPath(outputRoot, "artifact.json")
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	judgePath, err := resolveOutputPath(outputRoot, "judge.go")
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	manifestPath, err := resolveOutputPath(outputRoot, "generation-manifest.json")
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	if err := writeAtomicFiles([]atomicWrite{
		{path: policyPath, data: policyBytes},
		{path: artifactPath, data: artifactBytes},
		{path: judgePath, data: judge},
		{path: manifestPath, data: manifestBytes},
	}); err != nil {
		return reportMetaPolicyProfileError(stderr, fmt.Errorf("write profile artifacts: %w", err))
	}
	if jsonMode {
		if _, err := stdout.Write(manifestBytes); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "generated profile: %s\npolicy: %s\nartifact: %s\njudge: %s\nmanifest: %s\n", policycompilation.PublicProfileID, policyPath, artifactPath, judgePath, manifestPath)
	return exitOK
}

type interfaceWriter interface {
	Write([]byte) (int, error)
}

func reportMetaPolicyProfileError(stderr interfaceWriter, err error) int {
	fmt.Fprintf(stderr, "gooo: generate profile: %v\n", err)
	return exitFailure
}

func profileJSONBytes(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func profileCanonicalDirectory(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("project boundary is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("project boundary absolute path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("inspect project boundary: %w", err)
	}
	if !info.IsDir() {
		return "", errors.New("project boundary is not a directory")
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("canonical project boundary: %w", err)
	}
	return filepath.Clean(canonical), nil
}

func profileCanonicalInput(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("input source is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("input source absolute path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("inspect input source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("input source is not a regular file")
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("canonical input source: %w", err)
	}
	return filepath.Clean(canonical), nil
}

func profileCanonicalCandidate(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("output root is empty")
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("output root absolute path: %w", err)
	}
	candidate := filepath.Clean(abs)
	suffix := []string{}
	for {
		info, err := os.Lstat(candidate)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return "", errors.New("output root is a symbolic link")
			}
			if !info.IsDir() {
				return "", errors.New("output root is not a directory")
			}
			canonical, err := filepath.EvalSymlinks(candidate)
			if err != nil {
				return "", fmt.Errorf("canonical output root: %w", err)
			}
			root := filepath.Clean(canonical)
			for _, part := range suffix {
				root = filepath.Join(root, part)
			}
			return filepath.Clean(root), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("inspect output root: %w", err)
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			return "", errors.New("output root has no existing ancestor")
		}
		suffix = append([]string{filepath.Base(candidate)}, suffix...)
		candidate = parent
	}
}

func profilePathWithin(root, target string) bool {
	if root == target {
		return true
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || filepath.IsAbs(rel) || rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func profilePathsOverlap(left, right string) bool {
	return left == right || profilePathWithin(left, right) || profilePathWithin(right, left)
}

func profileRelativePath(root, target string) string {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return filepath.Base(target)
	}
	return filepath.ToSlash(rel)
}
