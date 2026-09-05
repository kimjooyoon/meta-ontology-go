package valueexecution

import "fmt"

// IntegerValue is the native value produced by the currently registered
// Integer operation. It is returned only through a validated ProducedResult.
type IntegerValue int64

// ResultEvidence is a detached, data-only view of a produced result. It is
// not a constructor and carries no execution authority.
type ResultEvidence struct {
	ProducerActivityID  string `json:"producer_activity_id"`
	ProducerActivity    string `json:"producer_activity"`
	OutputEntityID      string `json:"output_entity_id"`
	OutputEntity        string `json:"output_entity"`
	SourceDigest        string `json:"source_digest"`
	SemanticFingerprint string `json:"semantic_fingerprint"`
	OperationSpecDigest string `json:"operation_spec_digest"`
	Value               int64  `json:"value"`
	ResultDigest        string `json:"result_digest"`
}

// ProducedResult is opaque outside this package. Its zero value and any JSON
// decoded value are invalid because its private authority cannot be supplied
// by ordinary exported-field construction.
type ProducedResult struct {
	authority    resultAuthority
	value        int64
	resultDigest string
}

// UnmarshalJSON deliberately rejects every JSON input. ResultEvidence may be
// serialized as detached data, but JSON never restores a ProducedResult
// authority. A valid JSON attempt clears the destination before returning a
// structured failure; syntax rejected before this method may leave a
// destination unchanged.
func (result *ProducedResult) UnmarshalJSON(_ []byte) error {
	if result != nil {
		*result = ProducedResult{}
	}
	return failAt(ReasonResultHandleInvalid, "RESULT", "unmarshal-produced-result", "ProducedResult does not support JSON deserialization")
}

type producedResultDigestInput struct {
	ProducerActivityID  string `json:"producer_activity_id"`
	ProducerActivity    string `json:"producer_activity"`
	OutputEntityID      string `json:"output_entity_id"`
	OutputEntity        string `json:"output_entity"`
	SourceDigest        string `json:"source_digest"`
	SemanticFingerprint string `json:"semantic_fingerprint"`
	OperationSpecDigest string `json:"operation_spec_digest"`
	Value               int64  `json:"value"`
}

// The digest is content identity for replay comparison only. It intentionally
// carries no invocation counter or attempt identity.
func issueProducedResult(authority resultAuthority, value int64) ProducedResult {
	result := ProducedResult{authority: authority, value: value}
	result.resultDigest = digestValue(producedResultDigestInput{
		ProducerActivityID: authority.activityID, ProducerActivity: authority.activityName,
		OutputEntityID: authority.outputEntityID, OutputEntity: authority.outputEntityName,
		SourceDigest: authority.sourceDigest, SemanticFingerprint: authority.semanticFingerprint,
		OperationSpecDigest: authority.operationSpecDigest, Value: value,
	})
	return result
}

func (result ProducedResult) validate() error {
	if !result.authority.valid() || result.resultDigest == "" {
		return failAt(ReasonResultHandleInvalid, "RESULT", "validate-produced-result", "result handle has no valid private authority")
	}
	want := digestValue(producedResultDigestInput{
		ProducerActivityID: result.authority.activityID, ProducerActivity: result.authority.activityName,
		OutputEntityID: result.authority.outputEntityID, OutputEntity: result.authority.outputEntityName,
		SourceDigest: result.authority.sourceDigest, SemanticFingerprint: result.authority.semanticFingerprint,
		OperationSpecDigest: result.authority.operationSpecDigest, Value: result.value,
	})
	if result.resultDigest != want || !validDigest(result.resultDigest) {
		return failAt(ReasonResultHandleInvalid, "RESULT", "validate-produced-result", "result content digest is invalid")
	}
	return nil
}

// Valid reports whether result is a sealed, non-zero produced result.
func (result ProducedResult) Valid() bool { return result.validate() == nil }

// Evidence returns detached provenance data. Mutating the returned value does
// not mutate the opaque result authority.
func (result ProducedResult) Evidence() ResultEvidence {
	if result.validate() != nil {
		return ResultEvidence{}
	}
	return ResultEvidence{
		ProducerActivityID: result.authority.activityID, ProducerActivity: result.authority.activityName,
		OutputEntityID: result.authority.outputEntityID, OutputEntity: result.authority.outputEntityName,
		SourceDigest: result.authority.sourceDigest, SemanticFingerprint: result.authority.semanticFingerprint,
		OperationSpecDigest: result.authority.operationSpecDigest, Value: result.value, ResultDigest: result.resultDigest,
	}
}

// Integer returns the validated Integer value carried by the result.
func (result ProducedResult) Integer() (IntegerValue, error) {
	if err := result.validate(); err != nil {
		return 0, err
	}
	if result.authority.outputEntityName != IntegerEntity {
		return 0, failAt(ReasonResultHandleInvalid, "RESULT", "read-integer-result", fmt.Sprintf("output entity %q is not %q", result.authority.outputEntityName, IntegerEntity))
	}
	return IntegerValue(result.value), nil
}
