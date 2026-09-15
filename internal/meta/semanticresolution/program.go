package semanticresolution

const (
	Schema    = "gooo/meta-semantic-resolution-program/v1"
	Authority = "semanticresolution.CanonicalProgram"
)

func CanonicalProgram() Program {
	return Program{
		Schema: Schema, Authority: Authority,
		Resolutions: CanonicalResolutions(), MetaOperations: CanonicalOperations(),
		Indicators: CanonicalIndicators(),
	}
}
