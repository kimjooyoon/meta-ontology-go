package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

// Emit is the only effect writer. It accepts an exact independent
// authorization for the same provisional receipt and writes only a temporary
// artifact. Repository promotion remains outside this authority scope.
func Emit(receipt model.Receipt, judgment model.Judgment, subjectSHA, path string) (model.Effect, error) {
	tempRoot, relativePath, err := rootedTempTarget(path)
	if err != nil {
		return model.Effect{}, err
	}
	if !judgment.Independent || judgment.Decision != model.DecisionAllowed || judgment.Resolution != model.ResolutionExact || judgment.Status != model.StatusDischarged || judgment.CheckedClaims != len(receipt.Claims) || judgment.DischargedClaims != len(receipt.Claims) || judgment.OpenClaims != 0 || judgment.RefutedClaims != 0 ||
		judgment.AuthorizationDigest == "" || receipt.AuthorizationDigest != judgment.AuthorizationDigest ||
		model.AuthorizationDigest(receipt) != receipt.AuthorizationDigest || receipt.Phase != model.ReceiptProvisional ||
		receipt.HeadSHA != subjectSHA || receipt.Evidence.EffectIntent != "approved-artifact" || len(receipt.Effects) != 0 ||
		!model.ValidHead(subjectSHA) || receipt.Digest == "" || receipt.Digest != model.SealReceipt(receipt).Digest || tempRoot == "" || relativePath == "" {
		return model.Effect{}, fmt.Errorf("effect authorization is not exact or is stale")
	}
	data := ArtifactBytes(receipt)
	root, err := os.OpenRoot(tempRoot)
	if err != nil {
		return model.Effect{}, fmt.Errorf("open temporary artifact root: %w", err)
	}
	defer root.Close()
	file, err := root.OpenFile(relativePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return model.Effect{}, fmt.Errorf("write target is not a safe temporary descendant: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return model.Effect{}, fmt.Errorf("write temporary artifact: %w", err)
	}
	if err := file.Close(); err != nil {
		return model.Effect{}, fmt.Errorf("close temporary artifact: %w", err)
	}
	artifact := model.ArtifactEvidence{
		Path: path, ContentDigest: model.DigestBytes(data), Size: len(data), CaseID: receipt.CaseID, ExecutionID: receipt.ExecutionID, SubjectSHA: subjectSHA,
		AuthorizationDigest: receipt.AuthorizationDigest, Producer: receipt.Producer, Executor: model.ExecutorID,
		Consumer: receipt.Consumer, RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false, RepositoryNetState: model.RepositoryNetStateUnknown,
	}
	effect := model.Effect{
		Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved", ArtifactDigest: artifact.ContentDigest,
		ArtifactPath: artifact.Path, ArtifactSize: artifact.Size, Artifact: artifact, CaseID: receipt.CaseID, ExecutionID: receipt.ExecutionID, SubjectSHA: subjectSHA,
		Intent: receipt.Evidence.EffectIntent, AuthorizationDigest: receipt.AuthorizationDigest, Producer: receipt.Producer,
		Executor: model.ExecutorID, Consumer: receipt.Consumer, MetaOperation: "execute-authorized-temp-artifact",
		TempArtifactWriteAuthorized: true, RepositoryNetStatusObserved: false, RepositoryNetStatusUnchanged: false, RepositoryNetState: model.RepositoryNetStateUnknown, RepositoryActualOrTransientWrites: model.UnknownEffectScope,
		RepositoryPathAuthorization: false, AmbientProcessAuthority: model.UnknownEffectScope,
	}
	executionDigest := model.EffectExecutionDigest(effect)
	effect.ExecutionReceiptDigest = executionDigest
	effect.Artifact.EffectReceiptDigest = executionDigest
	return effect, nil
}

func ArtifactBytes(receipt model.Receipt) []byte {
	return []byte(fmt.Sprintf("gooo bounded transformation artifact\ncase=%s\nexecution=%s\ninput=%d\noperation=%s\noutput=%d\nsource=%s\nsemantic-source=%s\nauthorization=%s\nsubject=%s\n",
		receipt.CaseID, receipt.ExecutionID, receipt.Evidence.InputValue, receipt.Evidence.CandidateOperation, receipt.Evidence.CandidateResult,
		receipt.SourceDigest, receipt.SemanticSourceDigest, receipt.AuthorizationDigest, receipt.HeadSHA))
}

func Observe(effect model.Effect) (model.ArtifactEvidence, error) {
	if effect.Kind != model.EffectApproved || effect.Artifact.Path == "" || effect.Artifact.Path != effect.ArtifactPath ||
		effect.Artifact.ContentDigest != effect.ArtifactDigest || effect.Artifact.Size != effect.ArtifactSize {
		return model.ArtifactEvidence{}, fmt.Errorf("artifact evidence metadata is inconsistent")
	}
	data, err := os.ReadFile(effect.Artifact.Path)
	if err != nil {
		return model.ArtifactEvidence{}, fmt.Errorf("observe artifact: %w", err)
	}
	if len(data) != effect.Artifact.Size || model.DigestBytes(data) != effect.Artifact.ContentDigest {
		return model.ArtifactEvidence{}, fmt.Errorf("artifact bytes do not match receipt")
	}
	if model.EffectExecutionDigest(effect) != effect.ExecutionReceiptDigest || effect.Artifact.EffectReceiptDigest != effect.ExecutionReceiptDigest {
		return model.ArtifactEvidence{}, fmt.Errorf("artifact execution receipt digest is invalid")
	}
	return effect.Artifact, nil
}

func Path(root, label string) string {
	return filepath.Join(root, "gooo-invariant-transformation-"+label+".bin")
}

func rootedTempTarget(path string) (string, string, error) {
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	roots := []string{root}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("effect target path is not absolute: %w", err)
	}
	for _, candidate := range roots {
		canonicalRoot, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			continue
		}
		canonicalRoot, err = filepath.Abs(canonicalRoot)
		if err != nil {
			continue
		}
		if repositoryRoot, ok := canonicalRepositoryRoot(); ok {
			candidateAbsolute, candidateErr := filepath.Abs(candidate)
			if pathsOverlap(canonicalRoot, repositoryRoot) || (candidateErr == nil && within(repositoryRoot, candidateAbsolute)) || within(repositoryRoot, absolutePath) {
				continue
			}
		}
		resolvedPath := absolutePath
		if resolved, resolveErr := filepath.EvalSymlinks(absolutePath); resolveErr == nil {
			resolvedPath, err = filepath.Abs(resolved)
			if err != nil {
				continue
			}
		} else {
			resolvedParent, parentErr := filepath.EvalSymlinks(filepath.Dir(absolutePath))
			if parentErr != nil {
				continue
			}
			resolvedParent, err = filepath.Abs(resolvedParent)
			if err != nil {
				continue
			}
			resolvedPath = filepath.Join(resolvedParent, filepath.Base(absolutePath))
		}
		relative, err := filepath.Rel(canonicalRoot, resolvedPath)
		if err != nil || relative == "." || relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
			continue
		}
		if !within(canonicalRoot, resolvedPath) {
			continue
		}
		return canonicalRoot, relative, nil
	}
	return "", "", fmt.Errorf("EFFECT_AUTHORIZATION/validate-rooted-temp-target/REPOSITORY_OR_SYMLINK_ESCAPE_REJECTED")
}

func canonicalRepositoryRoot() (string, bool) {
	root := os.Getenv("GITHUB_WORKSPACE")
	if root == "" {
		root, _ = os.Getwd()
		for {
			if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
				break
			}
			parent := filepath.Dir(root)
			if parent == root {
				return "", false
			}
			root = parent
		}
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", false
	}
	resolved, err = filepath.Abs(resolved)
	return resolved, err == nil
}

func within(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != "." && relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func pathsOverlap(left, right string) bool {
	return left == right || within(left, right) || within(right, left)
}
