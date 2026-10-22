package bidir

// ReferenceResolutionError retains the authored occurrence that could not be
// resolved. Callers can unwrap the cause without guessing a location from its
// message or from another occurrence of the same display name.
type ReferenceResolutionError struct {
	Span  SourceSpan
	Cause error
}

func (e *ReferenceResolutionError) Error() string { return e.Cause.Error() }

func (e *ReferenceResolutionError) Unwrap() error { return e.Cause }

func referenceResolutionError(reference Reference, cause error) error {
	return &ReferenceResolutionError{Span: reference.Span, Cause: cause}
}
