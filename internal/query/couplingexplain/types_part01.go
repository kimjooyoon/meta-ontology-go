package couplingexplain

type Status string

const (
	StatusPass       Status = "PASS"
	StatusFailClosed Status = "FAIL_CLOSED"
	StatusUnknown    Status = "UNKNOWN"
)

// LinkReason is the closed set of reasons for withholding an authority link.
type LinkReason string

const (
	ReasonAmbiguous    LinkReason = "AMBIGUOUS"
	ReasonStale        LinkReason = "STALE"
	ReasonUnregistered LinkReason = "UNREGISTERED"
	ReasonMissing      LinkReason = "MISSING"
	ReasonNotVerified  LinkReason = "NOT_VERIFIED"
)

type View string

const (
	ViewCompact  View = "COMPACT"
	ViewExpanded View = "EXPANDED"
)

// ChangeClaim is deliberately separate from semantic.SemanticChangeKind.
type ChangeClaim string

const (
	ClaimDelta   ChangeClaim = "DELTA"
	ClaimNoDelta ChangeClaim = "NO_DELTA"
)

type VerifierState string

const (
	VerifierPass       VerifierState = "PASS"
	VerifierFailClosed VerifierState = "FAIL_CLOSED"
	VerifierUnknown    VerifierState = "UNKNOWN"
)

type EnvelopeVerdict string

const (
	VerdictVerified   EnvelopeVerdict = "VERIFIED"
	VerdictFailClosed EnvelopeVerdict = "FAIL_CLOSED"
	VerdictUnknown    EnvelopeVerdict = "UNKNOWN"
)

// Request binds a query to one exact verified envelope and snapshot.
type Request struct {
	CodeSymbolID         string  `json:"code_symbol_id"`
	SnapshotDigest       string  `json:"snapshot_digest"`
	RegistryDigest       string  `json:"registry_digest"`
	SourceMapDigest      string  `json:"source_map_digest"`
	ManifestDigest       string  `json:"manifest_digest"`
	ToolchainDigest      string  `json:"toolchain_digest"`
	ProfileDigest        string  `json:"profile_digest"`
	DetectorInputDigest  string  `json:"detector_input_digest"`
	DetectorResultDigest string  `json:"detector_result_digest"`
	VerifierResultDigest string  `json:"verifier_result_digest"`
	EnvelopeDigest       string  `json:"envelope_digest"`
	Control              Control `json:"control"`
}
