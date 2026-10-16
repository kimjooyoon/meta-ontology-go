package conceptoperation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

// VerifyEvidence keeps legacy source replay and observed-data assessment behind
// the same readiness entry point. The two receipt schemas remain distinct.
func VerifyEvidence(payload []byte, inputs SourceInputs, repository fs.FS, repositoryRoot,
	expectedRepository, expectedSubjectSHA string) (EvidenceState, error) {
	var header struct {
		Schema string `json:"schema"`
	}
	if json.Unmarshal(payload, &header) == nil && header.Schema == ObservedReceiptSchema {
		return verifyObservedSource(payload, inputs, repository, repositoryRoot, expectedRepository, expectedSubjectSHA)
	}
	if inputs.Observation != nil {
		return EvidenceState{}, fmt.Errorf("verify concept-operation evidence: observations require their own receipt")
	}
	evidence, err := Verify(payload, expectedRepository, expectedSubjectSHA)
	if err != nil {
		return EvidenceState{}, fmt.Errorf("verify concept-operation evidence: %w", err)
	}
	if err := VerifySource(evidence, inputs, repository, repositoryRoot, expectedRepository, expectedSubjectSHA); err != nil {
		return EvidenceState{}, fmt.Errorf("verify concept-operation source: %w", err)
	}
	return EvidenceState{Digest: evidence.Digest, Status: evidence.Status}, nil
}

func verifyObservedSource(payload []byte, inputs SourceInputs, repository fs.FS, repositoryRoot,
	expectedRepository, expectedSubjectSHA string) (EvidenceState, error) {
	if err := completeness.CheckJSON(payload, 1<<20); err != nil {
		return EvidenceState{}, err
	}
	var receipt ObservedReceipt
	if err := decodeExact(payload, &receipt); err != nil {
		return EvidenceState{}, err
	}
	if receipt.Schema != ObservedReceiptSchema || receipt.MetricID != MetricID ||
		receipt.Producer != "metricprogram/conceptoperation.BuildObserved" || receipt.Consumer != "language-readiness" ||
		receipt.RepositoryWorkspaceWrites || receipt.PromotionAuthorized || inputs.Observation == nil {
		return EvidenceState{}, fmt.Errorf("observed concept-operation receipt identity is invalid")
	}
	if err := VerifySource(receipt.Baseline, inputs, repository, repositoryRoot, expectedRepository, expectedSubjectSHA); err != nil {
		return EvidenceState{}, fmt.Errorf("verify observed baseline source: %w", err)
	}
	source, err := fs.ReadFile(repository, ObservationPolicyPath)
	if err != nil || !bytes.Equal(source, inputs.Observation.PolicySource) {
		return EvidenceState{}, fmt.Errorf("observation policy differs from repository source")
	}
	rebuilt, err := BuildObserved(context.Background(), inputs)
	if err != nil {
		return EvidenceState{}, fmt.Errorf("rebuild observed concept-operation receipt: %w", err)
	}
	if !artifact.Equal(rebuilt, receipt) {
		return EvidenceState{}, fmt.Errorf("observed concept-operation receipt does not match source replay")
	}
	assessment, err := observedAssessment(rebuilt)
	if err != nil {
		return EvidenceState{}, err
	}
	return EvidenceState{Digest: rebuilt.Digest, Status: assessment.Status}, nil
}
