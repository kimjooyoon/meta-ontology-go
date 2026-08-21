package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const SchemaVersion = "gooo-selective-ci-evidence/v1"

type DecisionStatus string

const (
	Verified   DecisionStatus = "VERIFIED"
	Unknown    DecisionStatus = "UNKNOWN"
	FailClosed DecisionStatus = "FAIL_CLOSED"
)

type FallbackMode string

const (
	NoFallback FallbackMode = "NONE"
	FullSuite  FallbackMode = "FULL_SUITE"
)

type ReceiptStatus string

const (
	ReceiptVerified  ReceiptStatus = "VERIFIED"
	ReceiptCandidate ReceiptStatus = "CANDIDATE"
	ReceiptDeferred  ReceiptStatus = "DEFERRED"
	ReceiptNotRun    ReceiptStatus = "NOT_RUN"
)
const (
	CodeVerified        = "SELECTIVE_CI_V1_VERIFIED"
	CodeMissing         = "SELECTIVE_CI_V1_MISSING_INPUT"
	CodeStaleSnapshot   = "SELECTIVE_CI_V1_STALE_SNAPSHOT"
	CodeCandidate       = "SELECTIVE_CI_V1_CANDIDATE_ONLY"
	CodeDuplicate       = "SELECTIVE_CI_V1_DUPLICATE"
	CodeAmbiguous       = "SELECTIVE_CI_V1_AMBIGUOUS"
	CodeDisconnected    = "SELECTIVE_CI_V1_DISCONNECTED"
	CodeWrongEndpoint   = "SELECTIVE_CI_V1_WRONG_ENDPOINT"
	CodeCycle           = "SELECTIVE_CI_V1_CYCLE"
	CodeReceiptMismatch = "SELECTIVE_CI_V1_RECEIPT_MISMATCH"
	CodeDigestMismatch  = "SELECTIVE_CI_V1_DIGEST_MISMATCH"
	CodeMalformed       = "SELECTIVE_CI_V1_MALFORMED"
)

type SnapshotBinding struct {
	Base semantic.SnapshotDigests
	Head semantic.SnapshotDigests
}
type Path struct {
	PathID        semantic.ID
	RootID        semantic.ID
	ObligationID  semantic.ID
	CommandID     semantic.ID
	ReceiptID     semantic.ID
	RecordIDs     []semantic.ID
	ExpectedKinds []semantic.InferenceKind
}
type CommandReceipt struct {
	CommandID             semantic.ID
	ReceiptID             semantic.ID
	Status                ReceiptStatus
	ProviderReceiptDigest string
	PhaseReceiptDigest    string
	ResourceReceiptDigest string
	RegistryDigest        string
	PlanDigest            string
	Digest                string
}
