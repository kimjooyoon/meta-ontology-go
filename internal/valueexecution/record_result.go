package valueexecution

import "maps"

// RecordFields is data only. Field names and required presence are checked
// against the entity schema compiled from the Gooo source.
type RecordFields map[string]string

type RecordEvidence struct {
	ProducerActivityID  string             `json:"producer_activity_id"`
	ProducerActivity    string             `json:"producer_activity"`
	OutputEntityID      string             `json:"output_entity_id"`
	SourceDigest        string             `json:"source_digest"`
	SemanticFingerprint string             `json:"semantic_fingerprint"`
	OperationSpecDigest string             `json:"operation_spec_digest"`
	RootActivity        string             `json:"root_activity"`
	RootInputDigest     string             `json:"root_input_digest"`
	ParentResultDigest  string             `json:"parent_result_digest"`
	InputOrigin         string             `json:"input_origin"`
	Scope               string             `json:"scope"`
	Authority           OperationAuthority `json:"authority"`
	Fields              RecordFields       `json:"fields"`
	ResultDigest        string             `json:"result_digest"`
}

// ProducedRecord is an opaque runtime result. Its provenance proves which
// compiled activity transported data, not whether the data's claims are true.
type ProducedRecord struct {
	authority    resultAuthority
	fields       RecordFields
	rootActivity string
	rootDigest   string
	parentDigest string
	digest       string
}

func (result *ProducedRecord) UnmarshalJSON(_ []byte) error {
	if result != nil {
		*result = ProducedRecord{}
	}
	return failAt(ReasonResultHandleInvalid, "RESULT", "unmarshal-produced-record", "record evidence cannot restore runtime authority")
}

func issueProducedRecord(authority resultAuthority, fields RecordFields, rootActivity, rootDigest, parentDigest string) ProducedRecord {
	result := ProducedRecord{authority: authority, fields: maps.Clone(fields),
		rootActivity: rootActivity, rootDigest: rootDigest, parentDigest: parentDigest}
	result.digest = digestValue(result.detachedEvidence())
	return result
}

func (result ProducedRecord) detachedEvidence() RecordEvidence {
	return RecordEvidence{
		ProducerActivityID: result.authority.activityID, ProducerActivity: result.authority.activityName,
		OutputEntityID: result.authority.outputEntityID, SourceDigest: result.authority.sourceDigest,
		SemanticFingerprint: result.authority.semanticFingerprint, OperationSpecDigest: result.authority.operationSpecDigest,
		RootActivity: result.rootActivity, RootInputDigest: result.rootDigest, ParentResultDigest: result.parentDigest,
		InputOrigin: "CALLER_SUPPLIED_DATA", Scope: RecordTransportScope, Fields: maps.Clone(result.fields),
	}
}

func (result ProducedRecord) validate() error {
	if !result.authority.valid() || result.rootActivity == "" || !validDigest(result.rootDigest) ||
		result.parentDigest != "" && !validDigest(result.parentDigest) || result.fields == nil ||
		!validDigest(result.digest) || result.digest != digestValue(result.detachedEvidence()) {
		return failAt(ReasonResultHandleInvalid, "RESULT", "validate-produced-record", "record result has no valid private authority")
	}
	return nil
}

func (result ProducedRecord) Valid() bool { return result.validate() == nil }

func (result ProducedRecord) Evidence() RecordEvidence {
	if result.validate() != nil {
		return RecordEvidence{}
	}
	evidence := result.detachedEvidence()
	evidence.ResultDigest = result.digest
	return evidence
}
