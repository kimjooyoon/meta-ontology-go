package adapter

// ProvenanceReceiptSchema is the append-only receipt wire schema.
const ProvenanceReceiptSchema = "gooo/provenance-receipt/v1"

// ReceiptProvenanceStatus records whether the receipt is verified or deferred.
type ReceiptProvenanceStatus string

const (
	ReceiptProvenanceVerified ReceiptProvenanceStatus = "verified"
	ReceiptProvenanceDeferred ReceiptProvenanceStatus = "deferred"
)

// ReceiptOutcome describes the transaction represented by a receipt.
type ReceiptOutcome string

const (
	ReceiptOutcomeAccepted  ReceiptOutcome = "accepted"
	ReceiptOutcomeRejected  ReceiptOutcome = "rejected"
	ReceiptOutcomeCancelled ReceiptOutcome = "cancelled"
	ReceiptOutcomeClosed    ReceiptOutcome = "closed"
)

// ReceiptWriteEffect is an observer-level summary of filesystem effects.
type ReceiptWriteEffect string

const (
	ReceiptWriteEffectNone     ReceiptWriteEffect = "none"
	ReceiptWriteEffectObserved ReceiptWriteEffect = "observed"
)

// ReceiptJob is one caller-supplied check in the six-job tuple.
type ReceiptJob struct {
	Name       string `json:"name"`
	HeadSHA    string `json:"head_sha"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

// ReceiptPredecessor identifies an earlier append-only evidence record.
type ReceiptPredecessor struct {
	EventRef string `json:"event_ref"`
	Digest   string `json:"digest"`
}

// ProvenanceReceipt is an append-only, deterministic verification record.
// Repository, refs, run IDs, and jobs are caller-supplied; no CI state is
// inferred or embedded by the serializer.
type ProvenanceReceipt struct {
	Schema             string                  `json:"schema"`
	Repository         string                  `json:"repository"`
	BaseSHA            string                  `json:"base_sha"`
	HeadSHA            string                  `json:"head_sha"`
	EventRef           string                  `json:"event_ref"`
	CheckoutRef        string                  `json:"checkout_ref"`
	Run                string                  `json:"run"`
	Attempt            int                     `json:"attempt"`
	ArtifactCount      int                     `json:"artifact_count"`
	Jobs               []ReceiptJob            `json:"jobs"`
	Binding            ObservationBinding      `json:"binding"`
	PreconditionDigest string                  `json:"precondition_sha256"`
	BeforeStateDigest  string                  `json:"before_state_sha256"`
	AfterStateDigest   string                  `json:"after_state_sha256"`
	Outcome            ReceiptOutcome          `json:"outcome"`
	WriteEffect        ReceiptWriteEffect      `json:"write_effect"`
	Predecessors       []ReceiptPredecessor    `json:"predecessors"`
	ProvenanceStatus   ReceiptProvenanceStatus `json:"provenance_status"`
}

var receiptJobNames = [...]string{
	"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy",
}
