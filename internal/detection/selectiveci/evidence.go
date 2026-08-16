package selectiveci

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateSelectedEvidence(input Input, selected []selectedPath) ([]string, []string, error) {
	receiptDigests := make([]string, 0, len(selected))
	pathIDs := make([]string, 0, len(selected))
	receipts := map[string]Receipt{}
	for _, receipt := range input.Receipts {
		receipts[receipt.CommandID] = receipt
	}
	paths := map[string]ProvenancePath{}
	for _, path := range input.ProvenancePaths {
		paths[path.CommandID] = path
	}
	for _, entry := range selected {
		receipt, ok := receipts[entry.command.ID]
		if !ok {
			return nil, nil, failure(ReasonResourceReceipt, "selected command has no resource receipt")
		}
		if receipt.SnapshotDigest != input.Head.Digest {
			return nil, nil, failure(ReasonMismatchedDigest, "resource receipt snapshot does not match head")
		}
		resource := resourceenvelope.Evaluate(receipt.Envelope)
		if resource.Status != resourceenvelope.PASS {
			return nil, nil, resourceFailure(resource.ReasonCode)
		}
		if resource.CPUCoreNS > entry.command.CPUWorkUnits || resource.PeakRSSBytes > entry.command.MemoryBytes {
			return nil, nil, failure(ReasonResourceLimit, "resource receipt exceeds command ceiling")
		}
		path, ok := paths[entry.command.ID]
		if !ok {
			return nil, nil, failure(ReasonAmbiguousPath, "selected command has no provenance path")
		}
		if err := evaluatePath(path); err != nil {
			return nil, nil, err
		}
		receiptDigests = append(receiptDigests, digestBytes([]byte(entry.command.ID+"\x00"+receipt.SnapshotDigest+"\x00"+resource.Canonical())))
		pathIDs = append(pathIDs, path.Requirement.PathID)
	}
	return sortedUnique(receiptDigests), sortedUnique(pathIDs), nil
}

func resourceFailure(reason string) error {
	if reason == "cpu-arithmetic" {
		return failure(ReasonResourceArithmetic, reason)
	}
	return failure(ReasonResourceReceipt, reason)
}

func evaluatePath(path ProvenancePath) error {
	requirement, err := pathRequirement(path.Requirement)
	if err != nil {
		return failure(ReasonAmbiguousPath, err.Error())
	}
	if normalized, normalizeErr := path.Path.Normalized(); normalizeErr == nil {
		if topologyErr := validatePathTopology(normalized, requirement); topologyErr != nil {
			return topologyErr
		}
	}
	evaluation := pathclosure.Evaluate(path.Path, []pathclosure.Requirement{requirement})
	if evaluation.Status == pathclosure.PASS && len(evaluation.Complete) == 1 {
		return nil
	}
	if evaluation.Code == pathclosure.CodeDuplicate {
		return failure(ReasonDuplicateID, evaluation.Code)
	}
	if evaluation.Code == pathclosure.CodeMissingRecord {
		return failure(ReasonUnknownPath, evaluation.Code)
	}
	if evaluation.Code == pathclosure.CodeMissingEvidence || evaluation.Code == pathclosure.CodeMissingSnapshot {
		return failure(ReasonDanglingReference, evaluation.Code)
	}
	if strings.Contains(strings.ToLower(evaluation.Code), "malformed") {
		return failure(ReasonAmbiguousPath, evaluation.Code)
	}
	return failure(ReasonEvaluatorError, evaluation.Code)
}

func validatePathTopology(path semantic.InferencePathV1, requirement pathclosure.Requirement) error {
	byID := make(map[semantic.ID]semantic.InferenceEdge, len(path.Edges))
	for _, edge := range path.Edges {
		byID[edge.RecordID] = edge
	}
	edges := make([]semantic.InferenceEdge, 0, len(requirement.RecordIDs))
	for _, recordID := range requirement.RecordIDs {
		edge, ok := byID[recordID]
		if !ok {
			return nil
		}
		edges = append(edges, edge)
	}
	if _, err := semantic.NewInferencePathChain(edges...); err != nil {
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "cycle") || strings.Contains(message, "path_orphan") {
			return failure(ReasonCycle, err.Error())
		}
		if strings.Contains(message, "path_ambiguity") {
			return failure(ReasonAmbiguousPath, err.Error())
		}
	}
	return nil
}

func pathRequirement(raw PathRequirement) (pathclosure.Requirement, error) {
	pathID, err := semantic.ParseIdentity(raw.PathID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	start, err := semantic.ParseIdentity(raw.StartID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	end, err := semantic.ParseIdentity(raw.EndID)
	if err != nil {
		return pathclosure.Requirement{}, err
	}
	records := make([]semantic.ID, len(raw.RecordIDs))
	kinds := make([]semantic.InferenceKind, len(raw.ExpectedKinds))
	for i := range raw.RecordIDs {
		records[i], err = semantic.ParseIdentity(raw.RecordIDs[i])
		if err != nil {
			return pathclosure.Requirement{}, err
		}
		kinds[i] = semantic.InferenceKind(raw.ExpectedKinds[i])
	}
	return pathclosure.Requirement{PathID: pathID, RecordIDs: records, ExpectedKinds: kinds, StartID: start, EndID: end}, nil
}
