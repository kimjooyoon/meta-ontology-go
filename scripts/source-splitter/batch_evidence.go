package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func summarizeSplitEvidence(expectedSHA, logical string, evidence splitEvidence) (splitBatchSubject, error) {
	empty := splitBatchSubject{}
	write := evidence.Write
	if evidence.OperationID != splitGoEvidenceOperationID || evidence.ExpectedHeadSHA != expectedSHA ||
		!evidence.EvidenceComplete || evidence.Source.Path != logical || !write.Complete ||
		!write.ExecutionSucceeded || write.WritesOutsideDeclaredTargets != 0 || write.TemporaryFilesRemaining != 0 {
		return empty, fmt.Errorf("split evidence for %s is incomplete", logical)
	}
	if len(write.DeclaredTargets) < 2 || len(evidence.Candidates) != len(write.DeclaredTargets) ||
		write.DeclaredTargets[0] != logical || len(write.Events) == 0 {
		return empty, fmt.Errorf("split evidence targets for %s are unknown", logical)
	}
	for index, target := range write.DeclaredTargets {
		if evidence.Candidates[index].Path != target || len(evidence.Candidates[index].Data) == 0 {
			return empty, fmt.Errorf("split candidate %d for %s is unknown", index, logical)
		}
	}
	for _, event := range write.Events {
		if !event.Success {
			return empty, fmt.Errorf("split event for %s failed", logical)
		}
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		return empty, err
	}
	digest := sha256.Sum256(encoded)
	changed := append([]string(nil), write.DeclaredTargets...)
	created := append([]string(nil), write.DeclaredTargets[1:]...)
	return splitBatchSubject{Logical: logical, Status: "applied", ChangedFiles: changed,
		CreatedFiles: created, ReceiptDigest: "sha256:" + hex.EncodeToString(digest[:])}, nil
}
