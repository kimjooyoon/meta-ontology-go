package coupling

// LocationLink is the standard LSP definition response shape.
type LocationLink struct {
	OriginSelectionRange *Range `json:"originSelectionRange,omitempty"`
	TargetURI            string `json:"targetUri"`
	TargetRange          Range  `json:"targetRange"`
	TargetSelectionRange Range  `json:"targetSelectionRange"`
}
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}
type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           int                            `json:"severity,omitempty"`
	Code               string                         `json:"code,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

const (
	DiagnosticError       = 1
	DiagnosticWarning     = 2
	DiagnosticInformation = 3
)
const (
	DiagnosticExplanation          = "gooo.coupling.explanation"
	DiagnosticMissingCancellation  = "gooo.coupling.missing-cancellation"
	DiagnosticMissingSnapshot      = "gooo.coupling.missing-snapshot"
	DiagnosticMissingVersion       = "gooo.coupling.missing-document-version"
	DiagnosticDocumentMismatch     = "gooo.coupling.document-mismatch"
	DiagnosticWrongVersion         = "gooo.coupling.wrong-document-version"
	DiagnosticStaleSnapshot        = "gooo.coupling.stale-snapshot"
	DiagnosticCancelled            = "gooo.coupling.cancelled"
	DiagnosticAmbiguous            = "gooo.coupling.ambiguous"
	DiagnosticUpstreamUnknown      = "gooo.coupling.upstream-unknown"
	DiagnosticUpstreamFail         = "gooo.coupling.upstream-fail-closed"
	DiagnosticNoBinding            = "gooo.coupling.no-binding"
	DiagnosticInvalidEnvelope      = "gooo.coupling.invalid-envelope"
	DiagnosticInvalidPosition      = "gooo.coupling.invalid-position"
	DiagnosticLiveInvalidEnvelope  = "gooo.coupling.query-invalid-envelope"
	DiagnosticLiveInvalidBinding   = "gooo.coupling.query-invalid-binding"
	DiagnosticLiveMissingLocations = "gooo.coupling.missing-source-locations"
	DiagnosticLiveMissingMaterial  = "gooo.coupling.missing-verified-material"
	DiagnosticLiveNotVerified      = "gooo.coupling.not-independently-verified"
)

// Result is an in-process aggregation. Its fields are independently
// serializable standard LSP values; Result itself is not a custom wire
// envelope.
type Result struct {
	Outcome     Outcome
	Links       []LocationLink
	Hover       *Hover
	Diagnostics []Diagnostic
}
type Adapter struct {
	envelope Envelope
	raw      []byte
}
