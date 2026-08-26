package transformationeffect

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

const splitGoContractPath = "examples/source-splitter-conformance/contract.json"

func operationObservations(root string, action generation.Action, applied []byte, fallbackEvidence string) ([]generation.IndicatorReceipt, *SplitGoEvaluationArtifact, string, error) {
	if action.Operation != sourcepolicy.OperationSplitGo {
		observations := make([]generation.IndicatorReceipt, 0, len(action.RequiredIndicatorIDs))
		for _, id := range action.RequiredIndicatorIDs {
			observations = append(observations, generation.IndicatorReceipt{ID: id,
				Verdict: generation.IndicatorVerdictUnknown, EvidenceDigest: hashJSON([]string{fallbackEvidence, id}), ProofChoice: action.ProofChoice})
		}
		return observations, nil, fallbackEvidence, nil
	}

	contractRaw, err := os.ReadFile(filepath.Join(root, splitGoContractPath))
	if err != nil {
		return nil, nil, "", fmt.Errorf("read SplitGo conformance contract: %w", err)
	}
	artifact, err := EvaluateSplitGo(contractRaw, applied, action.RequiredIndicatorIDs, string(action.ProofChoice))
	if err != nil {
		return nil, nil, "", fmt.Errorf("evaluate SplitGo raw evidence: %w", err)
	}
	return artifact.Receipts, &artifact, hashJSON(artifact), nil
}
