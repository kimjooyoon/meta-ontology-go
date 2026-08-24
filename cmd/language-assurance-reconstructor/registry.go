package main

const (
	transactionSchema = "gooo/language-assurance-transaction/v1"
	receiptSchema = "gooo/language-assurance-raw-reconstruction/v1"
	verifierID = "gooo-independent-json-reconstructor-v1"
	denominatorDigest = "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02"
	allowLimited = "ALLOW_LIMITED"
	block = "BLOCK"
	failClosed = "FAIL_CLOSED"
	reasonClear = "IMPLEMENTED_ASSURANCE_BOUNDARY_CLEAR"
	reasonEvidence = "ASSURANCE_EVIDENCE_UNKNOWN"
	reasonUnknown = "ASSURANCE_TOP_DECISION_UNKNOWN"
	reasonSnapshot = "ASSURANCE_SNAPSHOT_BINDING_MISMATCH"
	reasonGovernance = "ASSURANCE_GOVERNANCE_VIOLATION"
	exact = "EXACT"
	invariantOnly = "INVARIANT_ONLY"
	unknown = "UNKNOWN"
)

var conflictPairs = [][2]string{{"CONTRACT_AUTHOR", "EVALUATOR_AUTHOR"}, {"IMPLEMENTER", "PROMOTER"}, {"EVALUATOR_AUTHOR", "AUDITOR"}, {"POLICY_ADOPTER", "PROMOTER"}, {"ADAPTER_AUTHOR", "AUDITOR"}}
var launderingOutputs = map[string]bool{"PASS": true, "FIXED_POINT": true, "AUTHORIZED": true, "ALLOW": true}
var snapshotEvidenceIDs = map[string]bool{"authority_routes": true, "role_bindings": true, "decision_transitions": true}
