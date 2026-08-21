package bidir

// FieldOrigin identifies the provenance class of a latent field carrier.
// Only Source values are authoritative for public source readback.
type FieldOrigin string

const (
	FieldOriginSource      FieldOrigin = "Source"
	FieldOriginGenerated   FieldOrigin = "Generated"
	FieldOriginSynthesized FieldOrigin = "Synthesized"
	FieldOriginDeferred    FieldOrigin = "Deferred"
	FieldOriginUnsupported FieldOrigin = "Unsupported"
)

// TypeRefForm records the source/BX spelling form of a nominal type use.
// Form and Spelling are presentation provenance; ResolvedID is the semantic
// identity resolved from that presentation.
type TypeRefForm string

const (
	TypeRefFormLookup   TypeRefForm = "Lookup"
	TypeRefFormStableID TypeRefForm = "StableID"
)

// TypeRefUse preserves the original source spelling independently from the
// semantic TypeRef. It must never be reconstructed from a registry name.
type TypeRefUse struct {
	Form       TypeRefForm
	Spelling   string
	ResolvedID ID
	Span       SourceSpan
}

func validFieldOrigin(origin FieldOrigin) bool {
	switch origin {
	case FieldOriginSource, FieldOriginGenerated, FieldOriginSynthesized, FieldOriginDeferred, FieldOriginUnsupported:
		return true
	default:
		return false
	}
}
