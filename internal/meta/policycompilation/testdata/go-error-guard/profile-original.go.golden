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
	if options.profile == policycompilation.PublicPolicyRevisionProfileID {
		return runMetaPolicyRevisionProfile(options, input, jsonMode, stdout, stderr)
	}
	prepared, err := preparePublicProfile(options, input)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	artifacts, err := buildPublicProfileArtifacts(prepared)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	if err := writePublicProfileArtifacts(artifacts); err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	if jsonMode {
		if _, err := stdout.Write(artifacts.manifestBytes); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "generated profile: %s\npolicy: %s\nartifact: %s\njudge: %s\nmanifest: %s\n", policycompilation.PublicProfileID, artifacts.paths[0], artifacts.paths[1], artifacts.paths[2], artifacts.paths[3])
	return exitOK
}

type publicProfilePreparation struct {
	projectRoot string
	sourcePath  string
	outputRoot  string
	policy      policycompilation.CompiledPolicy
}

func preparePublicProfile(options generateOptions, input generateInput) (publicProfilePreparation, error) {
	projectRoot, err := profileCanonicalDirectory(options.profileProjectRoot)
	if err != nil {
		return publicProfilePreparation{}, err
	}
	sourcePath, err := profileCanonicalInput(options.filename)
	if err != nil {
		return publicProfilePreparation{}, err
	}
	if !profilePathWithin(projectRoot, sourcePath) {
		return publicProfilePreparation{}, fmt.Errorf("input source is outside the declared project boundary")
	}
	policy, err := policycompilation.CompileForIdentity(options.filename, input.source, options.profilePackage, options.profileNamespace)
	if err != nil {
		return publicProfilePreparation{}, fmt.Errorf("compile source policy: %w", err)
	}
	outputCandidate, err := profileCanonicalCandidate(options.outputDir)
	if err != nil {
		return publicProfilePreparation{}, err
	}
	if profilePathsOverlap(projectRoot, outputCandidate) {
		return publicProfilePreparation{}, errors.New("profile output must be outside the declared project boundary")
	}
	outputRoot, err := canonicalOutputRoot(options.outputDir)
	if err != nil {
		return publicProfilePreparation{}, err
	}
	if profilePathsOverlap(projectRoot, outputRoot) {
		return publicProfilePreparation{}, errors.New("profile output must be outside the declared project boundary")
	}
	if err := ensurePublicOutputDirectory(outputRoot); err != nil {
		return publicProfilePreparation{}, err
	}
	return publicProfilePreparation{projectRoot: projectRoot, sourcePath: sourcePath, outputRoot: outputRoot, policy: policy}, nil
}

type publicProfileArtifacts struct {
	policyBytes   []byte
	artifactBytes []byte
	judge         []byte
	manifestBytes []byte
	paths         []string
}

func buildPublicProfileArtifacts(prepared publicProfilePreparation) (publicProfileArtifacts, error) {
	judge := policycompilation.GenerateJudge(prepared.policy)
	policyBytes, err := profileJSONBytes(prepared.policy)
	if err != nil {
		return publicProfileArtifacts{}, fmt.Errorf("encode policy artifact: %w", err)
	}
	artifact := policycompilation.PolicyArtifact{Schema: policycompilation.ArtifactSchema, Policy: prepared.policy, GeneratedJudgeHash: policycompilation.DigestBytes(judge)}
	artifactBytes, err := profileJSONBytes(artifact)
	if err != nil {
		return publicProfileArtifacts{}, fmt.Errorf("encode compiled artifact: %w", err)
	}
	manifest := publicProfileManifest(prepared, policyBytes, artifactBytes, judge)
	manifestBytes, err := profileJSONBytes(manifest)
	if err != nil {
		return publicProfileArtifacts{}, fmt.Errorf("encode generation manifest: %w", err)
	}
	paths, err := publicProfilePaths(prepared.outputRoot)
	if err != nil {
		return publicProfileArtifacts{}, err
	}
	return publicProfileArtifacts{policyBytes: policyBytes, artifactBytes: artifactBytes, judge: judge, manifestBytes: manifestBytes, paths: paths}, nil
}

func publicProfileManifest(prepared publicProfilePreparation, policyBytes, artifactBytes, judge []byte) policycompilation.PublicGenerationManifest {
	policy := prepared.policy
	return policycompilation.PublicGenerationManifest{
		Schema: policycompilation.PublicGenerationManifestSchema, Profile: policycompilation.PublicProfileID,
		SourceFile: profileRelativePath(prepared.projectRoot, prepared.sourcePath), SourceDigest: policy.SourceDigest,
		SemanticDigest: policy.SemanticDigest, Package: policy.Package, Namespace: policy.Namespace,
		PolicyBytesDigest: policycompilation.DigestBytes(policyBytes), ArtifactBytesDigest: policycompilation.DigestBytes(artifactBytes),
		GeneratedJudgeDigest: policycompilation.DigestBytes(judge), GeneratedFiles: []string{"policy.json", "artifact.json", "judge.go", "generation-manifest.json"},
		OutputRootClass: policycompilation.PublicGenerationOutputRootClass, ExecutionObserved: false,
		CurrentConformance: policycompilation.PublicGenerationConformanceUnknown, RepositoryWrites: 0,
		MutationAuthority: 0, PromotionAuthority: 0,
	}
}

func publicProfilePaths(outputRoot string) ([]string, error) {
	paths := make([]string, 0, 4)
	for _, name := range []string{"policy.json", "artifact.json", "judge.go", "generation-manifest.json"} {
		path, err := resolveOutputPath(outputRoot, name)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func writePublicProfileArtifacts(artifacts publicProfileArtifacts) error {
	if err := writeAtomicFiles([]atomicWrite{
		{path: artifacts.paths[0], data: artifacts.policyBytes},
		{path: artifacts.paths[1], data: artifacts.artifactBytes},
		{path: artifacts.paths[2], data: artifacts.judge},
		{path: artifacts.paths[3], data: artifacts.manifestBytes},
	}); err != nil {
		return fmt.Errorf("write profile artifacts: %w", err)
	}
	return nil
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

func runMetaPolicyRevisionProfile(options generateOptions, input generateInput, jsonMode bool, stdout, stderr interfaceWriter) int {
	revision := policycompilation.PolicyDecisionRevision{
		ExpectedSourceDigest: options.profileSourceDigest,
		Condition:            options.profileCondition,
		FromDecision:         options.profileFromDecision,
		ToDecision:           options.profileToDecision,
	}
	// Reject an unbound request before preparing or creating an output directory.
	proposal, err := policycompilation.ProposePolicyDecisionRevision(options.filename, input.source, options.profilePackage, options.profileNamespace, revision)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	prepared, err := preparePublicProfile(options, input)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	candidatePath, err := resolveOutputPath(prepared.outputRoot, "candidate.gooo")
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	reportPath, err := resolveOutputPath(prepared.outputRoot, "proposal.json")
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	report := publicPolicyRevisionReport(prepared, revision, proposal)
	reportBytes, err := profileJSONBytes(report)
	if err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	if err := writeAtomicFiles([]atomicWrite{
		{path: candidatePath, data: []byte(proposal.CandidateSource)},
		{path: reportPath, data: reportBytes},
	}); err != nil {
		return reportMetaPolicyProfileError(stderr, err)
	}
	if jsonMode {
		if _, err := stdout.Write(reportBytes); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "generated profile: %s\ncandidate: %s\nproposal: %s\n", policycompilation.PublicPolicyRevisionProfileID, candidatePath, reportPath)
	return exitOK
}

func publicPolicyRevisionReport(prepared publicProfilePreparation, revision policycompilation.PolicyDecisionRevision, proposal policycompilation.PolicyDecisionProposal) policycompilation.PublicPolicyRevisionReport {
	return policycompilation.PublicPolicyRevisionReport{
		Schema: policycompilation.PublicPolicyRevisionReportSchema, Profile: policycompilation.PublicPolicyRevisionProfileID,
		SourceFile: profileRelativePath(prepared.projectRoot, prepared.sourcePath),
		Condition:  revision.Condition, FromDecision: revision.FromDecision, ToDecision: revision.ToDecision,
		Original: proposal.Original, Candidate: proposal.Candidate,
		ChangedCoordinates: append([]string(nil), proposal.ChangedCoordinates...),
		GeneratedFiles:     []string{"candidate.gooo", "proposal.json"},
		CandidateFormat:    policycompilation.PublicPolicyRevisionFormat,
		OutputRootClass:    policycompilation.PublicGenerationOutputRootClass,
		ExecutionObserved:  false, CurrentConformance: policycompilation.PublicGenerationConformanceUnknown,
		RepositoryWrites: 0, MutationAuthority: 0, PromotionAuthority: 0,
	}
}
