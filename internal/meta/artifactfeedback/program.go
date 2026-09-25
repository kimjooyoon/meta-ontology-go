package artifactfeedback

const (
	Schema    = "gooo/meta-operation-artifact-feedback-program/v1"
	Authority = "artifactfeedback.CanonicalProgram"
)

func CanonicalProgram() Program {
	return Program{
		Schema: Schema, Authority: Authority,
		MetaOperations: CanonicalOperations(), Indicators: CanonicalIndicators(),
	}
}
