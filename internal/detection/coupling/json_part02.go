package coupling

import (
	"encoding/json"
	"fmt"
)

func DecodeResult(data []byte) (Result, error) {
	var result Result
	if err := strictDecode(data, &result); err != nil {
		return Result{}, &JSONError{Code: ReasonMalformedBinding, Detail: err.Error()}
	}
	if result.Schema != ResultSchemaV1 || (result.Status != StatusPass && result.Status != StatusFailClosed && result.Status != StatusUnknown) {
		return Result{}, &JSONError{Code: ReasonMalformedBinding, Detail: "result schema or status"}
	}
	if err := validateResultShape(result); err != nil {
		return Result{}, &JSONError{Code: ReasonMalformedBinding, Detail: err.Error()}
	}
	if !validDigest(result.InputDigest) || !validDigest(result.Digest) || stableDigest(resultCanonical(result)) != result.Digest {
		return Result{}, &JSONError{Code: ReasonMalformedBinding, Detail: "result digest"}
	}
	return result, nil
}
func validateResultShape(result Result) error {
	validReasons := map[ReasonCode]struct{}{
		ReasonMalformedBinding: {}, ReasonRequiredInputMissing: {}, ReasonDuplicateSurface: {},
		ReasonSurfaceUnregistered: {}, ReasonDuplicateReceipt: {}, ReasonOrphanReceipt: {},
		ReasonStaleInput: {}, ReasonDigestMismatch: {}, ReasonSourceMapMismatch: {},
		ReasonContradictoryReceipt: {}, ReasonDeltaWithoutSource: {}, ReasonNoDeltaWithoutEquality: {},
		ReasonCandidateOnlyPath: {}, ReasonInferencePathMalformed: {}, ReasonMissingAuthorityPath: {},
		ReasonMissingVerification: {}, ReasonExternalReceiptMissing: {}, ReasonAuthorityInputSelfBound: {},
	}
	seenIDs := make(map[string]struct{}, len(result.AcceptedSurfaceIDs))
	for _, id := range result.AcceptedSurfaceIDs {
		if _, issue := normalizeID(id, "accepted surface ID"); issue != nil {
			return fmt.Errorf("invalid accepted surface ID")
		}
		if _, duplicate := seenIDs[id.String()]; duplicate {
			return fmt.Errorf("duplicate accepted surface ID")
		}
		seenIDs[id.String()] = struct{}{}
	}
	seenReasons := make(map[string]struct{}, len(result.Reasons))
	for _, reason := range result.Reasons {
		if _, ok := validReasons[reason.Code]; !ok || reason.Detail == "" {
			return fmt.Errorf("invalid result reason")
		}
		key := string(reason.Code) + "\x00" + reason.Detail
		if _, duplicate := seenReasons[key]; duplicate {
			return fmt.Errorf("duplicate result reason")
		}
		seenReasons[key] = struct{}{}
	}
	if result.Status == StatusPass {
		if len(result.Reasons) != 0 || result.FullSuiteRequired {
			return fmt.Errorf("PASS result has failure state")
		}
	} else if len(result.AcceptedSurfaceIDs) != 0 || len(result.Reasons) == 0 || !result.FullSuiteRequired {
		return fmt.Errorf("non-PASS result has invalid closure state")
	}
	for _, dimension := range []CountDimension{
		result.Observation.ChangedSurfaces, result.Observation.Receipts, result.Observation.InferenceRecords,
		result.Observation.InferencePaths, result.Observation.DeterministicWork, result.Observation.ResourceWork,
		result.Observation.CPU, result.Observation.Memory,
	} {
		if !dimension.Known && dimension.Value != 0 {
			return fmt.Errorf("unknown dimension carries a value")
		}
	}
	return nil
}
func EncodeResult(result Result) ([]byte, error) { return json.Marshal(result) }
