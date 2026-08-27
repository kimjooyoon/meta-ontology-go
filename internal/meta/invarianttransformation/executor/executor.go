package executor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

// Emit is the only effect writer. It accepts an exact independent
// authorization for the same provisional receipt and writes only a temporary
// artifact. Repository promotion remains outside this authority scope.
func Emit(receipt model.Receipt, judgment model.Judgment, subjectSHA, path string) (model.Effect, error) {
	if !judgment.Independent || judgment.Decision != model.DecisionAllowed || judgment.Resolution != model.ResolutionExact ||
		judgment.AuthorizationDigest == "" || receipt.AuthorizationDigest != judgment.AuthorizationDigest ||
		model.AuthorizationDigest(receipt) != receipt.AuthorizationDigest || receipt.Phase != model.ReceiptProvisional ||
		receipt.HeadSHA != subjectSHA || receipt.Evidence.EffectIntent != "approved-artifact" || len(receipt.Effects) != 0 ||
		!model.ValidHead(subjectSHA) || !filepath.IsAbs(path) {
		return model.Effect{}, fmt.Errorf("effect authorization is not exact or is stale")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return model.Effect{}, fmt.Errorf("prepare temporary artifact directory: %w", err)
	}
	data := ArtifactBytes(receipt)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return model.Effect{}, fmt.Errorf("write temporary artifact: %w", err)
	}
	artifact := model.ArtifactEvidence{
		Path: path, ContentDigest: model.DigestBytes(data), Size: len(data), CaseID: receipt.CaseID, SubjectSHA: subjectSHA,
		AuthorizationDigest: receipt.AuthorizationDigest, Producer: receipt.Producer, Executor: model.ExecutorID,
		Consumer: receipt.Consumer, RepositoryNetStatusUnchanged: true,
	}
	effect := model.Effect{
		Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved", ArtifactDigest: artifact.ContentDigest,
		ArtifactPath: artifact.Path, ArtifactSize: artifact.Size, Artifact: artifact, CaseID: receipt.CaseID, SubjectSHA: subjectSHA,
		Intent: receipt.Evidence.EffectIntent, AuthorizationDigest: receipt.AuthorizationDigest, Producer: receipt.Producer,
		Executor: model.ExecutorID, Consumer: receipt.Consumer, MetaOperation: "execute-authorized-temp-artifact",
		TempArtifactWriteAuthorized: true, RepositoryNetStatusUnchanged: true, RepositoryActualOrTransientWrites: model.UnknownEffectScope,
	}
	executionDigest := model.EffectExecutionDigest(effect)
	effect.ExecutionReceiptDigest = executionDigest
	effect.Artifact.EffectReceiptDigest = executionDigest
	return effect, nil
}

func ArtifactBytes(receipt model.Receipt) []byte {
	return []byte(fmt.Sprintf("gooo bounded transformation artifact\ncase=%s\ninput=%d\noperation=%s\noutput=%d\nsource=%s\nsemantic-source=%s\nauthorization=%s\nsubject=%s\n",
		receipt.CaseID, receipt.Evidence.InputValue, receipt.Evidence.CandidateOperation, receipt.Evidence.CandidateResult,
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
