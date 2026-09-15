package semantic

// ActivityContract is the compact semantic form from which deterministic PROV
// facts can be derived. It does not depend on any parser or AST type.
type ActivityContract struct {
	Activity ID
	Inputs   []ID
	Outputs  []ID
	Agents   []ID
	Span     Span
}
