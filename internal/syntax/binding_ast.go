package syntax

// BindingDecl declares an explicit data-flow edge between activity ports.
type BindingDecl struct {
	Span Span
	SourceActivity, SourcePort string
	TargetActivity, TargetPort string
}
